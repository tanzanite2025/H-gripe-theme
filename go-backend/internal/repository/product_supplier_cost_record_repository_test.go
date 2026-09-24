package repository

import (
	"testing"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductSupplierCostRecordRepositoryFindByProductCodesUsesSKUAsStableKey(t *testing.T) {
	db := newProductSupplierCostRecordTestDB(t)
	repo := NewProductSupplierCostRecordRepository(db)

	require.NoError(t, repo.Create(&suppliercostdomain.ProductSupplierCostRecord{
		ProductCode:          "SKU-PROC-002",
		ProductName:          "Second item",
		UnitCostMinor:        2000,
		Currency:             "USD",
		SupplierName:         "Supplier B",
		MinimumOrderQuantity: 1,
	}))
	require.NoError(t, repo.Create(&suppliercostdomain.ProductSupplierCostRecord{
		ProductCode:          "SKU-PROC-001",
		ProductName:          "First item",
		UnitCostMinor:        1000,
		Currency:             "USD",
		SupplierName:         "Supplier A",
		MinimumOrderQuantity: 1,
	}))

	records, err := repo.FindByProductCodes([]string{" SKU-PROC-002 ", "SKU-PROC-001", "SKU-PROC-001", ""})
	require.NoError(t, err)
	require.Len(t, records, 2)
	require.Equal(t, "SKU-PROC-001", records[0].ProductCode)
	require.Equal(t, "SKU-PROC-002", records[1].ProductCode)

	var productTableCount int64
	require.NoError(t, db.Raw(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('products', 'product_variants')",
	).Scan(&productTableCount).Error)
	require.Equal(t, int64(0), productTableCount)
}

func newProductSupplierCostRecordTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductSupplierCostRecord{}))
	return db
}
