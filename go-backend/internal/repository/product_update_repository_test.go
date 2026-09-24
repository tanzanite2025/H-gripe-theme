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
	item := &product.Product{SKU: "DISPLAY-READ-MODEL", Name: "Display read model", Slug: "display-read-model", Currency: "USD", PriceMinor: 1000, DisplayPriceData: datatypes.JSON([]byte(`[{"amount_decimal":"1.00"}]`))}
	require.NoError(t, repo.Create(item))
	require.NoError(t, db.Create(&product.ProductDisplayPriceSnapshot{
		ScopeKey:         "product:" + strconv.FormatUint(uint64(item.ID), 10),
		ProductID:        item.ID,
		SourceCurrency:   "USD",
		SourcePriceMinor: 1000,
		DisplayPriceData: datatypes.JSON([]byte(`[ {"amount_decimal": "0.92", "currency": "EUR"} ]`)),
	}).Error)

	var loaded product.Product
	require.NoError(t, db.First(&loaded, item.ID).Error)
	require.NoError(t, repo.attachProductDisplayPriceSnapshot(&loaded))
	require.JSONEq(t, `[ {"amount_decimal":"0.92","currency":"EUR"} ]`, string(loaded.DisplayPriceData))

	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", item.ID).Update("price_minor", 1200).Error)
	require.NoError(t, db.First(&loaded, item.ID).Error)
	require.NoError(t, repo.attachProductDisplayPriceSnapshot(&loaded))
	require.JSONEq(t, `[]`, string(loaded.DisplayPriceData))
}

func TestProductRepositoryRefreshProjectionLoadsVariantSelectionFields(t *testing.T) {
	db := newProductVariantTestDB(t)
	require.NoError(t, db.AutoMigrate(&product.ProductDisplayPriceSnapshot{}))
	repo := NewProductRepository(db)
	item := &product.Product{SKU: "DISPLAY-PROJECTION", Name: "Display projection", Slug: "display-projection", Currency: "USD", PriceMinor: 1000}
	require.NoError(t, repo.Create(item))
	variant := &product.ProductVariant{
		ProductID: item.ID, SKU: "DISPLAY-PROJECTION-V", Currency: "CNY", PriceMinor: 69900,
		OptionValues: "{}", IsActive: true, IsDefault: true,
	}
	require.NoError(t, db.Create(variant).Error)
	require.NoError(t, db.Create(&product.ProductDisplayPriceSnapshot{
		ScopeKey: "product:" + strconv.FormatUint(uint64(item.ID), 10), ProductID: item.ID,
		SourceCurrency: "CNY", SourcePriceMinor: variant.PriceMinor,
		DisplayPriceData: datatypes.JSON([]byte(`[{"amount_decimal":"97.86","currency":"USD"}]`)),
	}).Error)

	products, err := repo.ListProductsForDisplayPriceRefresh()
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.JSONEq(t, `[{"amount_decimal":"97.86","currency":"USD"}]`, string(products[0].DisplayPriceData))
}

func TestProductRepositorySEOUpdateDoesNotPersistProductStockSummary(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	record := &product.Product{
		SKU:        "SEO-STOCK-RACE",
		Name:       "SEO Stock Race",
		Slug:       "seo-stock-race",
		Currency:   "USD",
		PriceMinor: 10000,
		Stock:      20,
	}
	require.NoError(t, repo.Create(record))

	var stale product.Product
	require.NoError(t, db.First(&stale, record.ID).Error)
	require.NoError(t, repo.UpdateSEO(record.ID, "Updated title", "Updated description"))

	var stored product.Product
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Zero(t, stored.Stock)
	require.Equal(t, "Updated title", stored.MetaTitle)
	require.Equal(t, "Updated description", stored.MetaDesc)

	stale.MetaTitle = "Stale full update title"
	require.NoError(t, repo.Update(&stale))
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Zero(t, stored.Stock)
}

func TestProductRepositoryDisplayPriceRefreshUsesSourcePriceCAS(t *testing.T) {
	db := newProductVariantTestDB(t)
	require.NoError(t, db.AutoMigrate(&product.ProductDisplayPriceSnapshot{}))
	repo := NewProductRepository(db)
	record := &product.Product{
		SKU: "DISPLAY-PRICE-CAS", Name: "Display Price CAS", Slug: "display-price-cas",
		Currency: "USD", PriceMinor: 100000, DisplayPriceData: datatypes.JSON([]byte(`[{"amount_decimal":"1000.00"}]`)),
	}
	require.NoError(t, repo.Create(record))
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Updates(map[string]interface{}{"price_minor": 150000}).Error)

	err := repo.UpdateDisplayPriceSnapshots([]ProductDisplayPriceSnapshotUpdate{{
		ProductID: record.ID, SourceCurrency: "USD", SourcePriceMinor: 100000, UpdateProduct: true,
		DisplayPriceData: datatypes.JSON([]byte(`[{"amount_decimal":"900.00"}]`)),
	}})
	require.NoError(t, err)

	var stored product.Product
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Equal(t, int64(150000), stored.PriceMinor)
	require.Empty(t, stored.DisplayPriceData)
	var storedSnapshot product.ProductDisplayPriceSnapshot
	require.Error(t, db.Where("product_id = ?", record.ID).First(&storedSnapshot).Error)

	variant := product.ProductVariant{
		ProductID: record.ID, SKU: "DISPLAY-PRICE-CAS-VAR", Currency: "USD", PriceMinor: 10000,
		Stock: 1, IsActive: true, IsDefault: true,
	}
	require.NoError(t, db.Create(&variant).Error)
	require.NoError(t, db.Model(&product.ProductVariant{}).Where("id = ?", variant.ID).Updates(map[string]interface{}{"price_minor": 12000}).Error)
	require.NoError(t, repo.UpdateDisplayPriceSnapshots([]ProductDisplayPriceSnapshotUpdate{{
		ProductID: record.ID,
		VariantSnapshotUpdates: []ProductVariantDisplayPriceSnapshotUpdate{{
			VariantID: variant.ID, SourceCurrency: "USD", SourcePriceMinor: 10000,
			DisplayPriceData: datatypes.JSON([]byte(`[{"amount_decimal":"90.00"}]`)),
		}},
	}}))
	var storedVariant product.ProductVariant
	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, int64(12000), storedVariant.PriceMinor)
	require.Empty(t, storedVariant.DisplayPriceData)
}
