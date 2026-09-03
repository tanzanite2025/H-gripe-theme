package repository

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductProcurementCatalogRepositoryReturnsMinimalVariantProjection(t *testing.T) {
	db := newProductProcurementCatalogTestDB(t)
	repo := NewProductProcurementCatalogRepository(db)

	_, total, err := repo.ListOptions(ProductProcurementCatalogFilter{
		Page:     1,
		PageSize: 20,
		Search:   "picker",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)

	options, _, err := repo.ListOptions(ProductProcurementCatalogFilter{
		Page:     1,
		PageSize: 20,
		Search:   "picker",
	})
	require.NoError(t, err)
	require.Len(t, options, 1)
	require.Equal(t, "Picker product", options[0].ProductName)
	require.Equal(t, "Black / 700c", options[0].VariantTitle)
	require.Equal(t, "SKU-PICKER-BLACK", options[0].SKU)
	require.True(t, options[0].Available)

	encoded, err := json.Marshal(options[0])
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "stock")
	require.NotContains(t, string(encoded), "product_id")
	require.NotContains(t, string(encoded), "variant_id")
	require.NotContains(t, string(encoded), "currency")
	require.NotContains(t, string(encoded), "price")
	require.NotContains(t, string(encoded), "description")
	require.NotContains(t, string(encoded), "media")
	require.NotContains(t, string(encoded), "category")
}

func TestProductProcurementCatalogRepositoryExactSKUCanReopenInactiveHistory(t *testing.T) {
	db := newProductProcurementCatalogTestDB(t)
	repo := NewProductProcurementCatalogRepository(db)

	options, total, err := repo.ListOptions(ProductProcurementCatalogFilter{
		Page:     1,
		PageSize: 20,
		SKU:      "SKU-PICKER-INACTIVE",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, options, 1)
	require.False(t, options[0].Available)
}

func TestProductProcurementCatalogRepositoryDefaultListExcludesUnavailableVariants(t *testing.T) {
	db := newProductProcurementCatalogTestDB(t)
	repo := NewProductProcurementCatalogRepository(db)

	options, total, err := repo.ListOptions(ProductProcurementCatalogFilter{
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, options, 1)
	require.Equal(t, "SKU-PICKER-BLACK", options[0].SKU)
}

func newProductProcurementCatalogTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			deleted_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE product_variants (
			id INTEGER PRIMARY KEY,
			product_id INTEGER NOT NULL,
			sku TEXT NOT NULL,
			title TEXT,
			currency TEXT NOT NULL,
			price REAL NOT NULL,
			sale_price REAL,
			stock INTEGER NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			deleted_at DATETIME
		)
	`).Error)

	require.NoError(t, db.Exec(`
		INSERT INTO products (id, name, status, deleted_at) VALUES
			(11, 'Picker product', 'active', NULL),
			(12, 'Inactive product', 'inactive', NULL),
			(13, 'Deleted product', 'active', CURRENT_TIMESTAMP)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO product_variants (id, product_id, sku, title, currency, price, sale_price, stock, is_active, deleted_at) VALUES
			(101, 11, 'SKU-PICKER-BLACK', 'Black / 700c', 'USD', 100, 90, 999, TRUE, NULL),
			(102, 12, 'SKU-PICKER-INACTIVE', 'Inactive', 'USD', 80, NULL, 50, FALSE, NULL),
			(103, 13, 'SKU-PICKER-DELETED', 'Deleted', 'USD', 70, NULL, 30, TRUE, NULL)
	`).Error)
	return db
}
