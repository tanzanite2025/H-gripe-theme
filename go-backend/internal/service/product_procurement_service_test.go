package service

import (
	"testing"
	"time"

	procurementdomain "commerce-platform/internal/domain/procurement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestProductProcurementServiceCreatePersistsExtraCostFieldsFromCatalogSKU(t *testing.T) {
	db := newProductProcurementServiceTestDB(t)
	service := newProductProcurementServiceForTest(db)

	record, err := service.Create(ProductProcurementCreateInput{
		SKU: "SKU-PROC-EXTRA",
		ProductProcurementDetailsInput: ProductProcurementDetailsInput{
			PurchasePrice:           float64PointerForServiceTest(30),
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

	stored, err := repository.NewProductProcurementRepository(db).FindByProductCode("SKU-PROC-EXTRA")
	require.NoError(t, err)
	require.Equal(t, record.ID, stored.ID)
	require.Equal(t, 2.0, stored.InboundShippingUnitCost)
	require.Equal(t, 4.0, stored.PackagingUnitCost)
	require.Equal(t, 5.0, stored.OtherUnitCost)
}

func TestProductProcurementServiceUpdateKeepsIndependentSnapshot(t *testing.T) {
	db := newProductProcurementServiceTestDB(t)
	service := newProductProcurementServiceForTest(db)

	record, err := service.Create(ProductProcurementCreateInput{
		SKU: "SKU-PROC-UPDATE",
		ProductProcurementDetailsInput: ProductProcurementDetailsInput{
			PurchasePrice:        float64PointerForServiceTest(20),
			Currency:             "USD",
			SupplierName:         "Supplier A",
			LeadTimeDays:         10,
			MinimumOrderQuantity: 1,
		},
	})
	require.NoError(t, err)

	updated, err := service.Update(record.ID, ProductProcurementUpdateInput{
		ProductProcurementDetailsInput: ProductProcurementDetailsInput{
			PurchasePrice:           float64PointerForServiceTest(25),
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

	stored, err := repository.NewProductProcurementRepository(db).FindByProductCode("SKU-PROC-UPDATE")
	require.NoError(t, err)
	require.Equal(t, "Catalog update item", stored.ProductName)
	require.Equal(t, 25.0, stored.PurchasePrice)
	require.Equal(t, "Supplier B", stored.SupplierName)
}

func TestProductProcurementServiceSyncsAndClearsProfitSnapshotBySKU(t *testing.T) {
	db := newProductProcurementServiceTestDB(t)
	profitRepo := repository.NewProductProfitCalculationRepository(db)
	service := newProductProcurementServiceForTest(db)

	require.NoError(t, profitRepo.BulkUpsert([]procurementdomain.ProductProfitCalculation{{
		ProductCode:           "SKU-PROC-SYNC",
		ProductName:           "Old name",
		Currency:              "USD",
		ListPrice:             100,
		EffectiveSellingPrice: 100,
		PurchasePrice:         20,
		LandedCost:            20,
		GrossProfit:           80,
		GrossMarginBPS:        8000,
		CalculationStatus:     procurementdomain.ProfitStatusReady,
		FormulaVersion:        procurementdomain.ProfitFormulaVersion,
		WarningsData:          datatypes.JSON([]byte(`[]`)),
		CalculatedAt:          time.Now().UTC(),
	}}))

	created, err := service.Create(ProductProcurementCreateInput{
		SKU: "SKU-PROC-SYNC",
		ProductProcurementDetailsInput: ProductProcurementDetailsInput{
			PurchasePrice:           float64PointerForServiceTest(30),
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
	require.Equal(t, 30.0, profitRecord.PurchasePrice)
	require.Equal(t, 1.0, profitRecord.InboundShippingUnitCost)
	require.Equal(t, 3.0, profitRecord.PackagingUnitCost)
	require.Equal(t, 4.0, profitRecord.OtherUnitCost)
	require.Equal(t, 38.0, profitRecord.LandedCost)
	require.Equal(t, 62.0, profitRecord.GrossProfit)
	require.Equal(t, procurementdomain.ProfitStatusWarning, profitRecord.CalculationStatus)

	updated, err := service.Update(created.ID, ProductProcurementUpdateInput{
		ProductProcurementDetailsInput: ProductProcurementDetailsInput{
			PurchasePrice:           float64PointerForServiceTest(35),
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
	require.Equal(t, 35.0, profitRecord.PurchasePrice)
	require.Equal(t, 5.0, profitRecord.InboundShippingUnitCost)
	require.Equal(t, 7.0, profitRecord.PackagingUnitCost)
	require.Equal(t, 8.0, profitRecord.OtherUnitCost)
	require.Equal(t, 45.0, profitRecord.GrossProfit)

	require.NoError(t, service.Delete(updated.ID))

	_, err = profitRepo.FindByProductCode("SKU-PROC-SYNC")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repository.NewProductProcurementRepository(db).FindByProductCode("SKU-PROC-SYNC")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func newProductProcurementServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
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
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProcurement{}))
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProfitCalculation{}))
	return db
}

func newProductProcurementServiceForTest(db *gorm.DB) *ProductProcurementService {
	service := NewProductProcurementServiceWithProfitability(
		repository.NewProductProcurementRepository(db),
		repository.NewProductProfitCalculationRepository(db),
	)
	service.ConfigureCatalogRepository(repository.NewProductProcurementCatalogRepository(db))
	return service
}
