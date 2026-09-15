package repository

import (
	"strconv"
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestProductRepositoryLoadsDisplayPriceReadModelOnlyWhenSourceMatches(t *testing.T) {
	db := newProductVariantTestDB(t)
	require.NoError(t, db.AutoMigrate(&product.ProductDisplayPriceSnapshot{}))
	repo := NewProductRepository(db)
	item := &product.Product{SKU: "DISPLAY-READ-MODEL", Name: "Display read model", Slug: "display-read-model", Currency: "USD", PriceMinor: 1000, DisplayPriceData: datatypes.JSON([]byte(`[{"amount":1}]`))}
	require.NoError(t, repo.Create(item))
	require.NoError(t, db.Create(&product.ProductDisplayPriceSnapshot{
		ScopeKey:         "product:" + strconv.FormatUint(uint64(item.ID), 10),
		ProductID:        item.ID,
		SourceCurrency:   "USD",
		SourcePriceMinor: 1000,
		DisplayPriceData: datatypes.JSON([]byte(`[ {"amount": 0.92, "currency": "EUR"} ]`)),
	}).Error)

	var loaded product.Product
	require.NoError(t, db.First(&loaded, item.ID).Error)
	require.JSONEq(t, `[{"amount":0.92,"currency":"EUR"}]`, string(loaded.DisplayPriceData))

	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", item.ID).Update("price_minor", 1200).Error)
	require.NoError(t, db.First(&loaded, item.ID).Error)
	require.JSONEq(t, `[{"amount":1}]`, string(loaded.DisplayPriceData))
}

func TestProductRepositorySEOUpdateDoesNotRestoreConcurrentStock(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	record := &product.Product{
		SKU:      "SEO-STOCK-RACE",
		Name:     "SEO Stock Race",
		Slug:     "seo-stock-race",
		Currency: "USD",
		Price:    100,
		Stock:    20,
	}
	require.NoError(t, repo.Create(record))

	var stale product.Product
	require.NoError(t, db.First(&stale, record.ID).Error)
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Update("stock", 15).Error)
	require.NoError(t, repo.UpdateSEO(record.ID, "Updated title", "Updated description"))

	var stored product.Product
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Equal(t, 15, stored.Stock)
	require.Equal(t, "Updated title", stored.MetaTitle)
	require.Equal(t, "Updated description", stored.MetaDesc)

	stale.MetaTitle = "Stale full update title"
	require.NoError(t, repo.Update(&stale))
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Equal(t, 15, stored.Stock)
}

func TestProductRepositoryDisplayPriceRefreshUsesSourcePriceCAS(t *testing.T) {
	db := newProductVariantTestDB(t)
	require.NoError(t, db.AutoMigrate(&product.ProductDisplayPriceSnapshot{}))
	repo := NewProductRepository(db)
	record := &product.Product{
		SKU: "DISPLAY-PRICE-CAS", Name: "Display Price CAS", Slug: "display-price-cas",
		Currency: "USD", PriceMinor: 100000, Price: 1000, DisplayPriceData: datatypes.JSON([]byte(`[{"amount":1000}]`)),
	}
	require.NoError(t, repo.Create(record))
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Updates(map[string]interface{}{"price": 1500, "price_minor": 150000}).Error)

	err := repo.UpdateDisplayPriceSnapshots([]ProductDisplayPriceSnapshotUpdate{{
		ProductID: record.ID, SourceCurrency: "USD", SourcePriceMinor: 100000, UpdateProduct: true,
		DisplayPriceData: datatypes.JSON([]byte(`[{"amount":900}]`)),
	}})
	require.NoError(t, err)

	var stored product.Product
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Equal(t, 1500.0, stored.Price)
	require.JSONEq(t, `[{"amount":1000}]`, string(stored.DisplayPriceData))
	var storedSnapshot product.ProductDisplayPriceSnapshot
	require.Error(t, db.Where("product_id = ?", record.ID).First(&storedSnapshot).Error)

	variant := product.ProductVariant{
		ProductID: record.ID, SKU: "DISPLAY-PRICE-CAS-VAR", Currency: "USD", PriceMinor: 10000, Price: 100,
		Stock: 1, IsActive: true, IsDefault: true,
		DisplayPriceData: datatypes.JSON([]byte(`[{"amount":100}]`)),
	}
	require.NoError(t, db.Create(&variant).Error)
	require.NoError(t, db.Model(&product.ProductVariant{}).Where("id = ?", variant.ID).Updates(map[string]interface{}{"price": 120, "price_minor": 12000}).Error)
	require.NoError(t, repo.UpdateDisplayPriceSnapshots([]ProductDisplayPriceSnapshotUpdate{{
		ProductID: record.ID,
		VariantSnapshotUpdates: []ProductVariantDisplayPriceSnapshotUpdate{{
			VariantID: variant.ID, SourceCurrency: "USD", SourcePriceMinor: 10000,
			DisplayPriceData: datatypes.JSON([]byte(`[{"amount":90}]`)),
		}},
	}}))
	var storedVariant product.ProductVariant
	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 120.0, storedVariant.Price)
	require.JSONEq(t, `[{"amount":100}]`, string(storedVariant.DisplayPriceData))
}
