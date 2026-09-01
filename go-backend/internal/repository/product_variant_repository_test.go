package repository

import (
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductRepositoryDecrementVariantStocksUsesConditionalUpdate(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	p := &product.Product{
		SKU:      "LIMITED-RIM",
		Name:     "Limited Rim",
		Slug:     "limited-rim",
		Currency: "USD",
		Price:    199.99,
		Stock:    1,
	}
	require.NoError(t, repo.Create(p))

	variant := product.ProductVariant{
		ProductID: p.ID,
		SKU:       "LIMITED-RIM-01",
		Title:     "Limited Rim 01",
		Currency:  "USD",
		Price:     199.99,
		Stock:     1,
		IsActive:  true,
		IsDefault: true,
	}
	require.NoError(t, db.Create(&variant).Error)

	productIDs, err := repo.DecrementVariantStocks(map[uint]int{variant.ID: 1})
	require.NoError(t, err)
	require.Len(t, productIDs, 1)
	require.Equal(t, p.ID, productIDs[0])

	var storedVariant product.ProductVariant
	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 0, storedVariant.Stock)

	var storedProduct product.Product
	require.NoError(t, db.First(&storedProduct, p.ID).Error)
	require.Equal(t, 0, storedProduct.Stock)

	_, err = repo.DecrementVariantStocks(map[uint]int{variant.ID: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient stock")

	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 0, storedVariant.Stock)
}

func newProductVariantTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&product.Product{}, &product.ProductVariant{}))
	return db
}
