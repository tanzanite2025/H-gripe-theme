package repository

import (
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductRepositoryGetStatsUsesVariantInventoryInsteadOfProductSummary(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	lowStockProduct := seedInventoryStatsProduct(t, db, "LOW", product.FulfillmentModeStock, 99)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID: lowStockProduct.ID,
		SKU:       "LOW-VAR",
		Currency:  "USD",
		Price:     100,
		Stock:     5,
		IsActive:  true,
		IsDefault: true,
	}).Error)

	outOfStockProduct := seedInventoryStatsProduct(t, db, "OUT", product.FulfillmentModeStock, 5)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID: outOfStockProduct.ID,
		SKU:       "OUT-VAR",
		Currency:  "USD",
		Price:     100,
		Stock:     0,
		IsActive:  true,
		IsDefault: true,
	}).Error)

	masterProduct := seedInventoryStatsProduct(t, db, "MASTER", product.FulfillmentModeStock, 0)
	masterVariant := product.ProductVariant{
		ProductID: masterProduct.ID,
		SKU:       "MASTER-VAR",
		Currency:  "USD",
		Price:     100,
		Stock:     12,
		IsActive:  true,
		IsDefault: true,
	}
	require.NoError(t, db.Create(&masterVariant).Error)

	translatedProduct := seedInventoryStatsProduct(t, db, "TRANSLATED", product.FulfillmentModeStock, 0)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:       translatedProduct.ID,
		MasterVariantID: &masterVariant.ID,
		SKU:             "TRANSLATED-VAR",
		Currency:        "USD",
		Price:           100,
		Stock:           0,
		IsActive:        true,
		IsDefault:       true,
	}).Error)

	madeToOrderProduct := seedInventoryStatsProduct(t, db, "MTO", product.FulfillmentModeMadeToOrder, 0)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID: madeToOrderProduct.ID,
		SKU:       "MTO-VAR",
		Currency:  "USD",
		Price:     100,
		Stock:     0,
		IsActive:  true,
		IsDefault: true,
	}).Error)

	stats, err := repo.GetStats()
	require.NoError(t, err)
	require.Equal(t, int64(1), stats["low_stock"])
	require.Equal(t, int64(1), stats["out_of_stock"])
}

func seedInventoryStatsProduct(t *testing.T, db *gorm.DB, key, fulfillmentMode string, summaryStock int) *product.Product {
	t.Helper()
	item := &product.Product{
		SKU:             key,
		Name:            key,
		Slug:            "inventory-stats-" + key,
		Currency:        "USD",
		Price:           100,
		Stock:           summaryStock,
		Status:          "active",
		FulfillmentMode: fulfillmentMode,
	}
	require.NoError(t, db.Create(item).Error)
	return item
}
