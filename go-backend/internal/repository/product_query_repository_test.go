package repository

import (
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/product"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSearchPublicUsesStableIDTieBreakerForPagination(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)
	updatedAt := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	products := seedPublicProductsWithSameUpdatedAt(t, db, updatedAt, 3)

	firstPage, total, err := repo.SearchPublic(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 0,
		Limit:  2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)

	secondPage, total, err := repo.SearchPublic(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 2,
		Limit:  2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)

	requireProductIDs(t, firstPage, products[2].ID, products[1].ID)
	requireProductIDs(t, secondPage, products[0].ID)
}

func TestSearchPublicCompactUsesStableIDTieBreakerForPagination(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)
	updatedAt := time.Date(2026, 9, 3, 13, 0, 0, 0, time.UTC)

	products := seedPublicProductsWithSameUpdatedAt(t, db, updatedAt, 3)

	firstPage, err := repo.SearchPublicCompact(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 0,
		Limit:  2,
	})
	require.NoError(t, err)

	secondPage, err := repo.SearchPublicCompact(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 2,
		Limit:  2,
	})
	require.NoError(t, err)

	requireProductIDs(t, firstPage, products[2].ID, products[1].ID)
	requireProductIDs(t, secondPage, products[0].ID)
}

func newProductQueryTestDB(t *testing.T) *gorm.DB {
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

	require.NoError(t, db.AutoMigrate(
		&product.ProductBrand{},
		&product.ProductSpecificationTemplate{},
		&product.SpecDefinition{},
		&product.CustomsClassificationProfile{},
		&product.ProductInformationTemplate{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
	))
	return db
}

func seedPublicProductsWithSameUpdatedAt(t *testing.T, db *gorm.DB, updatedAt time.Time, count int) []product.Product {
	t.Helper()

	products := make([]product.Product, 0, count)
	for index := 0; index < count; index++ {
		item := product.Product{
			SKU:       fmt.Sprintf("STABLE-PAGE-%d", index),
			Name:      fmt.Sprintf("Stable Page %d", index),
			Slug:      fmt.Sprintf("stable-page-%d", index),
			Status:    "active",
			Locale:    "en",
			Price:     100,
			Stock:     1,
			CreatedAt: updatedAt,
			UpdatedAt: updatedAt,
		}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Create(&product.ProductVariant{
			ProductID: item.ID,
			SKU:       fmt.Sprintf("STABLE-PAGE-%d-VAR", index),
			Title:     "Default",
			Price:     100,
			Stock:     1,
			IsActive:  true,
			IsDefault: true,
			CreatedAt: updatedAt,
			UpdatedAt: updatedAt,
		}).Error)
		products = append(products, item)
	}
	require.NoError(t, db.Model(&product.Product{}).Where("id IN ?", productIDs(products)).UpdateColumn("updated_at", updatedAt).Error)
	return products
}

func productIDs(products []product.Product) []uint {
	ids := make([]uint, 0, len(products))
	for _, item := range products {
		ids = append(ids, item.ID)
	}
	return ids
}

func requireProductIDs(t *testing.T, products []product.Product, expectedIDs ...uint) {
	t.Helper()

	actualIDs := make([]uint, 0, len(products))
	for _, item := range products {
		actualIDs = append(actualIDs, item.ID)
	}
	require.Equal(t, expectedIDs, actualIDs)
}
