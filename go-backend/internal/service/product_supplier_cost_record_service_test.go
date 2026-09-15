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
			UnitCost:                float64PointerForServiceTest(30),
			Currency:                "USD",
			SupplierName:            "Supplier X",
			SupplierContactName:     "Alice",
			SupplierPhone:           "+1-555-001",
			SupplierEmail:           "alice@example.com",
			LeadTimeDays:            12,
			MinimumOrderQuantity:    8,
			InboundShippingUnitCost: 2,
			PackagingUnitCost:       4,
			OtherUnitCost:           5,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "SKU-PROC-EXTRA", record.ProductCode)
	require.Equal(t, "Catalog extra item", record.ProductName)
	require.Equal(t, 2.0, record.InboundShippingUnitCost)
	require.Equal(t, 4.0, record.PackagingUnitCost)
	require.Equal(t, 5.0, record.OtherUnitCost)

	stored, err := repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-PROC-EXTRA")
	require.NoError(t, err)
	require.Equal(t, record.ID, stored.ID)
	require.Equal(t, 2.0, stored.InboundShippingUnitCost)
	require.Equal(t, 4.0, stored.PackagingUnitCost)
	require.Equal(t, 5.0, stored.OtherUnitCost)
}

func TestProductSupplierCostRecordServiceUpdateKeepsIndependentSnapshot(t *testing.T) {
	db := newProductSupplierCostRecordServiceTestDB(t)
	service := newProductSupplierCostRecordServiceForTest(db)

	record, err := service.Create(ProductSupplierCostRecordCreateInput{
		SKU: "SKU-PROC-UPDATE",
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCost:             float64PointerForServiceTest(20),
			Currency:             "USD",
			SupplierName:         "Supplier A",
			LeadTimeDays:         10,
			MinimumOrderQuantity: 1,
		},
	})
	require.NoError(t, err)

	updated, err := service.Update(record.ID, ProductSupplierCostRecordUpdateInput{
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCost:                float64PointerForServiceTest(25),
			Currency:                "USD",
			SupplierName:            "Supplier B",
			SupplierContactName:     "Bob",
			SupplierPhone:           "+1-555-002",
			SupplierEmail:           "bob@example.com",
			LeadTimeDays:            18,
			MinimumOrderQuantity:    3,
			InboundShippingUnitCost: 1,
			PackagingUnitCost:       3,
			OtherUnitCost:           4,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "SKU-PROC-UPDATE", updated.ProductCode)
	require.Equal(t, "Catalog update item", updated.ProductName)
	require.Equal(t, "Supplier B", updated.SupplierName)
	require.Equal(t, 1.0, updated.InboundShippingUnitCost)
	require.Equal(t, 3.0, updated.PackagingUnitCost)
	require.Equal(t, 4.0, updated.OtherUnitCost)

	stored, err := repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-PROC-UPDATE")
	require.NoError(t, err)
	require.Equal(t, "Catalog update item", stored.ProductName)
	require.Equal(t, 25.0, stored.UnitCost)
	require.Equal(t, "Supplier B", stored.SupplierName)
}

func TestProductSupplierCostRecordServiceSyncsAndClearsProfitSnapshotBySKU(t *testing.T) {
	db := newProductSupplierCostRecordServiceTestDB(t)
	profitRepo := repository.NewProductProfitCalculationRepository(db)
	service := newProductSupplierCostRecordServiceForTest(db)

	require.NoError(t, profitRepo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{{
		ProductCode:           "SKU-PROC-SYNC",
		ProductName:           "Old name",
		Currency:              "USD",
		ListPrice:             100,
		EffectiveSellingPrice: 100,
		UnitCost:              20,
		LandedCost:            20,
		GrossProfit:           80,
		GrossMarginBPS:        8000,
		CalculationStatus:     suppliercostdomain.ProfitStatusReady,
		FormulaVersion:        suppliercostdomain.ProfitFormulaVersion,
		WarningsData:          datatypes.JSON([]byte(`[]`)),
		CalculatedAt:          time.Now().UTC(),
	}}))

	created, err := service.Create(ProductSupplierCostRecordCreateInput{
		SKU: "SKU-PROC-SYNC",
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCost:                float64PointerForServiceTest(30),
			Currency:                "USD",
			SupplierName:            "Supplier Sync",
			LeadTimeDays:            9,
			MinimumOrderQuantity:    6,
			InboundShippingUnitCost: 1,
			PackagingUnitCost:       3,
			OtherUnitCost:           4,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Catalog sync item", created.ProductName)

	profitRecord, err := profitRepo.FindByProductCode("SKU-PROC-SYNC")
	require.NoError(t, err)
	require.Equal(t, "Catalog sync item", profitRecord.ProductName)
	require.Equal(t, 30.0, profitRecord.UnitCost)
	require.Equal(t, 1.0, profitRecord.InboundShippingUnitCost)
	require.Equal(t, 3.0, profitRecord.PackagingUnitCost)
	require.Equal(t, 4.0, profitRecord.OtherUnitCost)
	require.Equal(t, 38.0, profitRecord.LandedCost)
	require.Equal(t, 62.0, profitRecord.GrossProfit)
	require.Equal(t, suppliercostdomain.ProfitStatusWarning, profitRecord.CalculationStatus)

	updated, err := service.Update(created.ID, ProductSupplierCostRecordUpdateInput{
		ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
			UnitCost:                float64PointerForServiceTest(35),
			Currency:                "USD",
			SupplierName:            "Supplier Sync 2",
			LeadTimeDays:            11,
			MinimumOrderQuantity:    6,
			InboundShippingUnitCost: 5,
			PackagingUnitCost:       7,
			OtherUnitCost:           8,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Supplier Sync 2", updated.SupplierName)

	profitRecord, err = profitRepo.FindByProductCode("SKU-PROC-SYNC")
	require.NoError(t, err)
	require.Equal(t, "Catalog sync item", profitRecord.ProductName)
	require.Equal(t, 35.0, profitRecord.UnitCost)
	require.Equal(t, 5.0, profitRecord.InboundShippingUnitCost)
	require.Equal(t, 7.0, profitRecord.PackagingUnitCost)
	require.Equal(t, 8.0, profitRecord.OtherUnitCost)
	require.Equal(t, 45.0, profitRecord.GrossProfit)

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
