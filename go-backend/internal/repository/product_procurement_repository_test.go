package repository

import (
	"testing"

	procurementdomain "commerce-platform/internal/domain/procurement"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductProcurementRepositoryFindByProductCodesUsesSKUAsStableKey(t *testing.T) {
	db := newProductProcurementTestDB(t)
	repo := NewProductProcurementRepository(db)

	require.NoError(t, repo.Create(&procurementdomain.ProductProcurement{
		ProductCode:          "SKU-PROC-002",
		ProductName:          "Second item",
		PurchasePrice:        20,
		Currency:             "USD",
		SupplierName:         "Supplier B",
		MinimumOrderQuantity: 1,
	}))
	require.NoError(t, repo.Create(&procurementdomain.ProductProcurement{
		ProductCode:          "SKU-PROC-001",
		ProductName:          "First item",
		PurchasePrice:        10,
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

func newProductProcurementTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProcurement{}))
	return db
}
