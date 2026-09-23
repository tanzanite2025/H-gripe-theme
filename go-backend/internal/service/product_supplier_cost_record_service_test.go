package service

import (
	"path/filepath"
	"testing"
	"time"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductSupplierCostRecordServiceCreatePersistsExtraCostFieldsFromCatalogSKU(t *testing.T) {
	db := newProductSupplierCostRecordServiceTestDB(t)
	service := newProductSupplierCostRecordServiceForTest(db)

	record, err := service.Create(ProductSupplierCostRecordCreateInput{
		SKU: "SKU-PROC-EXTRA",
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCostMinor:                int64PointerForServiceTest(3000),
			Currency:                     "USD",
			SupplierName:                 "Supplier X",
			SupplierContactName:          "Alice",
			SupplierPhone:                "+1-555-001",
			SupplierEmail:                "alice@example.com",
			LeadTimeDays:                 12,
			MinimumOrderQuantity:         8,
			InboundShippingUnitCostMinor: 200,
			PackagingUnitCostMinor:       400,
			OtherUnitCostMinor:           500,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "SKU-PROC-EXTRA", record.ProductCode)
	require.Equal(t, "Catalog extra item", record.ProductName)
	require.Equal(t, int64(200), record.InboundShippingUnitCostMinor)
	require.Equal(t, int64(400), record.PackagingUnitCostMinor)
	require.Equal(t, int64(500), record.OtherUnitCostMinor)

	stored, err := repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-PROC-EXTRA")
	require.NoError(t, err)
	require.Equal(t, record.ID, stored.ID)
	require.Equal(t, int64(200), stored.InboundShippingUnitCostMinor)
	require.Equal(t, int64(400), stored.PackagingUnitCostMinor)
	require.Equal(t, int64(500), stored.OtherUnitCostMinor)
}

func TestProductSupplierCostRecordServiceUpdateKeepsIndependentSnapshot(t *testing.T) {
	db := newProductSupplierCostRecordServiceTestDB(t)
	service := newProductSupplierCostRecordServiceForTest(db)

	record, err := service.Create(ProductSupplierCostRecordCreateInput{
		SKU: "SKU-PROC-UPDATE",
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCostMinor:        int64PointerForServiceTest(2000),
			Currency:             "USD",
			SupplierName:         "Supplier A",
			LeadTimeDays:         10,
			MinimumOrderQuantity: 1,
		},
	})
	require.NoError(t, err)

	updated, err := service.Update(record.ID, ProductSupplierCostRecordUpdateInput{
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCostMinor:                int64PointerForServiceTest(2500),
			Currency:                     "USD",
			SupplierName:                 "Supplier B",
			SupplierContactName:          "Bob",
			SupplierPhone:                "+1-555-002",
			SupplierEmail:                "bob@example.com",
			LeadTimeDays:                 18,
			MinimumOrderQuantity:         3,
			InboundShippingUnitCostMinor: 100,
			PackagingUnitCostMinor:       300,
			OtherUnitCostMinor:           400,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "SKU-PROC-UPDATE", updated.ProductCode)
	require.Equal(t, "Catalog update item", updated.ProductName)
	require.Equal(t, "Supplier B", updated.SupplierName)
	require.Equal(t, int64(100), updated.InboundShippingUnitCostMinor)
	require.Equal(t, int64(300), updated.PackagingUnitCostMinor)
	require.Equal(t, int64(400), updated.OtherUnitCostMinor)

	stored, err := repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-PROC-UPDATE")
	require.NoError(t, err)
	require.Equal(t, "Catalog update item", stored.ProductName)
	require.Equal(t, int64(2500), stored.UnitCostMinor)
	require.Equal(t, "Supplier B", stored.SupplierName)
}

func TestProductSupplierCostRecordServiceSyncsAndClearsProfitSnapshotBySKU(t *testing.T) {
	db := newProductSupplierCostRecordServiceTestDB(t)
	profitRepo := repository.NewProductProfitCalculationRepository(db)
	service := newProductSupplierCostRecordServiceForTest(db)

	require.NoError(t, profitRepo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{{
		ProductCode:                "SKU-PROC-SYNC",
		ProductName:                "Old name",
		Currency:                   "USD",
		ListPriceMinor:             10000,
		EffectiveSellingPriceMinor: 10000,
		UnitCostMinor:              2000,
		LandedCostMinor:            2000,
		GrossProfitMinor:           8000,
		GrossMarginBPS:             8000,
		CalculationStatus:          suppliercostdomain.ProfitStatusReady,
		FormulaVersion:             suppliercostdomain.ProfitFormulaVersion,
		WarningsData:               datatypes.JSON([]byte(`[]`)),
		CalculatedAt:               time.Now().UTC(),
	}}))

	created, err := service.Create(ProductSupplierCostRecordCreateInput{
		SKU: "SKU-PROC-SYNC",
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCostMinor:                int64PointerForServiceTest(3000),
			Currency:                     "USD",
			SupplierName:                 "Supplier Sync",
			LeadTimeDays:                 9,
			MinimumOrderQuantity:         6,
			InboundShippingUnitCostMinor: 100,
			PackagingUnitCostMinor:       300,
			OtherUnitCostMinor:           400,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Catalog sync item", created.ProductName)

	profitRecord, err := profitRepo.FindByProductCode("SKU-PROC-SYNC")
	require.NoError(t, err)
	require.Equal(t, "Catalog sync item", profitRecord.ProductName)
	require.Equal(t, int64(3000), profitRecord.UnitCostMinor)
	require.Equal(t, int64(100), profitRecord.InboundShippingUnitCostMinor)
	require.Equal(t, int64(300), profitRecord.PackagingUnitCostMinor)
	require.Equal(t, int64(400), profitRecord.OtherUnitCostMinor)
	require.Equal(t, int64(3800), profitRecord.LandedCostMinor)
	require.Equal(t, int64(6200), profitRecord.GrossProfitMinor)
	require.Equal(t, suppliercostdomain.ProfitStatusWarning, profitRecord.CalculationStatus)

	updated, err := service.Update(created.ID, ProductSupplierCostRecordUpdateInput{
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCostMinor:                int64PointerForServiceTest(3500),
			Currency:                     "USD",
			SupplierName:                 "Supplier Sync 2",
			LeadTimeDays:                 11,
			MinimumOrderQuantity:         6,
			InboundShippingUnitCostMinor: 500,
			PackagingUnitCostMinor:       700,
			OtherUnitCostMinor:           800,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Supplier Sync 2", updated.SupplierName)

	profitRecord, err = profitRepo.FindByProductCode("SKU-PROC-SYNC")
	require.NoError(t, err)
	require.Equal(t, "Catalog sync item", profitRecord.ProductName)
	require.Equal(t, int64(3500), profitRecord.UnitCostMinor)
	require.Equal(t, int64(500), profitRecord.InboundShippingUnitCostMinor)
	require.Equal(t, int64(700), profitRecord.PackagingUnitCostMinor)
	require.Equal(t, int64(800), profitRecord.OtherUnitCostMinor)
	require.Equal(t, int64(4500), profitRecord.GrossProfitMinor)

	require.NoError(t, service.Delete(updated.ID))

	_, err = profitRepo.FindByProductCode("SKU-PROC-SYNC")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-PROC-SYNC")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func newProductSupplierCostRecordServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "product-supplier-cost-records.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			deleted_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO products (id, name, status, deleted_at) VALUES
			(1, 'Catalog extra item', 'active', NULL),
			(2, 'Catalog update item', 'active', NULL),
			(3, 'Catalog sync item', 'active', NULL)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO product_variants (id, product_id, sku, title, is_active, deleted_at) VALUES
			(11, 1, 'SKU-PROC-EXTRA', 'Default', TRUE, NULL),
			(12, 2, 'SKU-PROC-UPDATE', 'Default', TRUE, NULL),
			(13, 3, 'SKU-PROC-SYNC', 'Default', TRUE, NULL)
	`).Error)
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductSupplierCostRecord{}))
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductProfitCalculation{}))
	return db
}

func newProductSupplierCostRecordServiceForTest(db *gorm.DB) *ProductSupplierCostRecordService {
	service := NewProductSupplierCostRecordServiceWithProfitability(
		repository.NewProductSupplierCostRecordRepository(db),
		repository.NewProductProfitCalculationRepository(db),
	)
	service.ConfigureCatalogRepository(repository.NewProductSupplierCostCatalogRepository(db))
	return service
}
