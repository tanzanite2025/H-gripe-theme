package service

import (
	"path/filepath"
	"testing"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductProfitabilityServicePreviewDoesNotWriteDatabase(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	results, err := service.Preview([]ProfitabilityItemInput{{
		ProductCode:     "SKU-PREVIEW",
		ProductName:     "Preview item",
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		SalePriceMinor:  int64PointerForServiceTest(9000),
		UnitCostMinor:   int64PointerForServiceTest(5000),
		UnitCostKnown:   true,
	}})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, suppliercostdomain.ProfitStatusReady, results[0].Status)
	require.Equal(t, int64(4000), *results[0].GrossProfitMinor)

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityServiceBulkUpsertSkipsUnknownPurchaseAndPersistsExplicitZero(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	result, err := service.BulkUpsert([]ProfitabilityItemInput{
		{
			ProductCode:     "SKU-MISSING",
			ProductName:     "Missing cost",
			SellingCurrency: "USD",
			ListPriceMinor:  10000,
		},
		{
			ProductCode:     "SKU-ZERO",
			ProductName:     "Zero cost",
			SellingCurrency: "USD",
			ListPriceMinor:  10000,
			UnitCostMinor:   int64PointerForServiceTest(0),
			UnitCostKnown:   true,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Records, 1)
	require.Len(t, result.Skipped, 1)
	require.Equal(t, "SKU-MISSING", result.Skipped[0].ProductCode)
	require.Equal(t, suppliercostdomain.ProfitStatusMissingUnitCost, result.Skipped[0].Status)
	require.Equal(t, "SKU-ZERO", result.Records[0].ProductCode)
	require.Equal(t, int64(0), result.Records[0].UnitCostMinor)
	require.Equal(t, int64(10000), result.Records[0].GrossProfitMinor)

	_, err = repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-MISSING")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductProfitabilityServiceClearsOldSnapshotWhenUnitCostBecomesUnknown(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	_, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:     "SKU-CLEAR",
		ProductName:     "Clear cost",
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		UnitCostMinor:   int64PointerForServiceTest(4000),
		UnitCostKnown:   true,
		SupplierCostDetails: &ProfitabilitySupplierCostDetailsInput{
			SupplierName:         "Clear supplier",
			LeadTimeDays:         10,
			MinimumOrderQuantity: 25,
		},
	}})
	require.NoError(t, err)

	result, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:     "SKU-CLEAR",
		ProductName:     "Clear cost",
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
	}})
	require.NoError(t, err)
	require.Len(t, result.Records, 0)
	require.Len(t, result.Skipped, 1)
	require.Equal(t, "SKU-CLEAR", result.Skipped[0].ProductCode)

	_, err = repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-CLEAR")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-CLEAR")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductProfitabilityServiceBulkUpsertUpdatesSameSKU(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	first, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:                  "SKU-UPDATE",
		ProductName:                  "First name",
		SellingCurrency:              "USD",
		ListPriceMinor:               10000,
		UnitCostMinor:                int64PointerForServiceTest(6000),
		UnitCostKnown:                true,
		InboundShippingUnitCostMinor: 100,
		PackagingUnitCostMinor:       300,
		OtherUnitCostMinor:           400,
	}})
	require.NoError(t, err)
	require.Len(t, first.Records, 1)
	firstID := first.Records[0].ID

	second, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:                  "SKU-UPDATE",
		ProductName:                  "Second name",
		SellingCurrency:              "USD",
		ListPriceMinor:               10000,
		SalePriceMinor:               int64PointerForServiceTest(9000),
		UnitCostMinor:                int64PointerForServiceTest(4000),
		UnitCostKnown:                true,
		InboundShippingUnitCostMinor: 500,
		PackagingUnitCostMinor:       700,
		OtherUnitCostMinor:           800,
	}})
	require.NoError(t, err)
	require.Len(t, second.Records, 1)
	require.Equal(t, firstID, second.Records[0].ID)
	require.Equal(t, "Second name", second.Records[0].ProductName)
	require.Equal(t, int64(3000), second.Records[0].GrossProfitMinor)
	require.Equal(t, int64(500), second.Records[0].InboundShippingUnitCostMinor)
	require.Equal(t, int64(700), second.Records[0].PackagingUnitCostMinor)
	require.Equal(t, int64(800), second.Records[0].OtherUnitCostMinor)
	require.Equal(t, int64(6000), second.Records[0].LandedCostMinor)

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).
		Where("product_code = ?", "SKU-UPDATE").
		Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestProductProfitabilityServiceInvalidBatchDoesNotWriteValidItems(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	_, err := service.BulkUpsert([]ProfitabilityItemInput{
		{
			ProductCode:     "SKU-VALID",
			ProductName:     "Valid item",
			SellingCurrency: "USD",
			ListPriceMinor:  10000,
			UnitCostMinor:   int64PointerForServiceTest(4000),
			UnitCostKnown:   true,
		},
		{
			ProductCode:     "SKU-INVALID",
			ProductName:     "Invalid item",
			SellingCurrency: "USD",
			CostCurrency:    "CNY",
			ListPriceMinor:  10000,
			UnitCostMinor:   int64PointerForServiceTest(4000),
			UnitCostKnown:   true,
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrProductProfitabilityInvalid)

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)

	_, err = repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-VALID")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductProfitabilityServiceRejectsDuplicateCodesWithoutWriting(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityService(repository.NewProductProfitCalculationRepository(db))

	_, err := service.BulkUpsert([]ProfitabilityItemInput{
		{
			ProductCode:     "SKU-DUPLICATE",
			ProductName:     "First item",
			SellingCurrency: "USD",
			ListPriceMinor:  10000,
			UnitCostMinor:   int64PointerForServiceTest(4000),
			UnitCostKnown:   true,
		},
		{
			ProductCode:     " SKU-DUPLICATE ",
			ProductName:     "Second item",
			SellingCurrency: "USD",
			ListPriceMinor:  10000,
			UnitCostMinor:   int64PointerForServiceTest(3500),
			UnitCostKnown:   true,
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrProductProfitabilityInvalid)
	var validationErr *ProfitabilityBatchValidationError
	require.ErrorAs(t, err, &validationErr)
	require.Len(t, validationErr.Items, 1)
	require.Contains(t, validationErr.Items[0].Reason, "duplicates item 1")

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityServicePreviewTreatsUnmarkedUnitCostAsUnknown(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityService(repository.NewProductProfitCalculationRepository(db))

	results, err := service.Preview([]ProfitabilityItemInput{{
		ProductCode:     "SKU-MISSING-FLAG",
		ProductName:     "Missing flag",
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		UnitCostMinor:   int64PointerForServiceTest(2000),
	}})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, suppliercostdomain.ProfitStatusMissingUnitCost, results[0].Status)
	require.Nil(t, results[0].GrossProfitMinor)
}

func TestProductProfitabilityServiceBulkUpsertPersistsSupplierCostRecordAndProfitSnapshotTogether(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	result, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:                  "SKU-SUPPLIER-COST-ATOMIC",
		ProductName:                  "Atomic supplier cost item",
		SellingCurrency:              "USD",
		ListPriceMinor:               10000,
		UnitCostMinor:                int64PointerForServiceTest(4000),
		UnitCostKnown:                true,
		InboundShippingUnitCostMinor: 200,
		PackagingUnitCostMinor:       400,
		OtherUnitCostMinor:           500,
		SupplierCostDetails: &ProfitabilitySupplierCostDetailsInput{
			SupplierName:         "Atomic Supplier",
			SupplierContactName:  "Lina",
			SupplierPhone:        "+86-123",
			SupplierEmail:        "lina@example.com",
			LeadTimeDays:         14,
			MinimumOrderQuantity: 20,
		},
	}})
	require.NoError(t, err)
	require.Len(t, result.Records, 1)

	supplierCostRecord, err := repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-SUPPLIER-COST-ATOMIC")
	require.NoError(t, err)
	require.Equal(t, "Atomic Supplier", supplierCostRecord.SupplierName)
	require.Equal(t, int64(4000), supplierCostRecord.UnitCostMinor)
	require.Equal(t, int64(200), supplierCostRecord.InboundShippingUnitCostMinor)
	require.Equal(t, int64(400), supplierCostRecord.PackagingUnitCostMinor)
	require.Equal(t, int64(500), supplierCostRecord.OtherUnitCostMinor)
	require.Equal(t, 14, supplierCostRecord.LeadTimeDays)
	require.Equal(t, 20, supplierCostRecord.MinimumOrderQuantity)

	profitRecord, err := repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-SUPPLIER-COST-ATOMIC")
	require.NoError(t, err)
	require.Equal(t, int64(4900), profitRecord.GrossProfitMinor)
}

func TestProductProfitabilityServiceBulkUpsertRollsBackProfitSnapshotWhenSupplierCostRecordWriteFails(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER force_supplier_cost_record_write_failure
		BEFORE INSERT ON product_procurement_records
		WHEN NEW.product_code = 'SKU-SUPPLIER-COST-ROLLBACK'
		BEGIN
			SELECT RAISE(ABORT, 'forced supplier cost record write failure');
		END
	`).Error)

	_, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:     "SKU-SUPPLIER-COST-ROLLBACK",
		ProductName:     "Rollback item",
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		UnitCostMinor:   int64PointerForServiceTest(4000),
		UnitCostKnown:   true,
		SupplierCostDetails: &ProfitabilitySupplierCostDetailsInput{
			SupplierName: "Unavailable supplier",
		},
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "forced supplier cost record write failure")

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).
		Where("product_code = ?", "SKU-SUPPLIER-COST-ROLLBACK").
		Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityServiceBulkUpsertUpdatesSupplierCostRecordBySKU(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithSupplierCostRecords(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductSupplierCostRecordRepository(db),
	)

	first, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:                  "SKU-SUPPLIER-COST-UPDATE",
		ProductName:                  "First item",
		SellingCurrency:              "USD",
		ListPriceMinor:               10000,
		UnitCostMinor:                int64PointerForServiceTest(6000),
		UnitCostKnown:                true,
		InboundShippingUnitCostMinor: 100,
		PackagingUnitCostMinor:       300,
		OtherUnitCostMinor:           400,
		SupplierCostDetails: &ProfitabilitySupplierCostDetailsInput{
			SupplierName: "First supplier",
		},
	}})
	require.NoError(t, err)
	require.Len(t, first.Records, 1)

	second, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:                  "SKU-SUPPLIER-COST-UPDATE",
		ProductName:                  "Second item",
		SellingCurrency:              "USD",
		ListPriceMinor:               10000,
		UnitCostMinor:                int64PointerForServiceTest(5000),
		UnitCostKnown:                true,
		InboundShippingUnitCostMinor: 500,
		PackagingUnitCostMinor:       700,
		OtherUnitCostMinor:           800,
		SupplierCostDetails: &ProfitabilitySupplierCostDetailsInput{
			SupplierName: "Second supplier",
			LeadTimeDays: 7,
		},
	}})
	require.NoError(t, err)
	require.Len(t, second.Records, 1)

	supplierCostRecord, err := repository.NewProductSupplierCostRecordRepository(db).FindByProductCode("SKU-SUPPLIER-COST-UPDATE")
	require.NoError(t, err)
	require.Equal(t, "Second supplier", supplierCostRecord.SupplierName)
	require.Equal(t, int64(500), supplierCostRecord.InboundShippingUnitCostMinor)
	require.Equal(t, int64(700), supplierCostRecord.PackagingUnitCostMinor)
	require.Equal(t, int64(800), supplierCostRecord.OtherUnitCostMinor)
	require.Equal(t, 7, supplierCostRecord.LeadTimeDays)

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductSupplierCostRecord{}).
		Where("product_code = ?", "SKU-SUPPLIER-COST-UPDATE").
		Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func newProductProfitabilityServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "product-profitability-with-supplier-cost.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductProfitCalculation{}))
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductSupplierCostRecord{}))
	return db
}

func float64PointerForServiceTest(value float64) *float64 {
	return &value
}

func int64PointerForServiceTest(value int64) *int64 {
	return &value
}
