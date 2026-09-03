package service

import (
	"testing"

	procurementdomain "commerce-platform/internal/domain/procurement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductProfitabilityServicePreviewDoesNotWriteDatabase(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	results, err := service.Preview([]ProfitabilityItemInput{{
		ProductCode:        "SKU-PREVIEW",
		ProductName:        "Preview item",
		SellingCurrency:    "USD",
		ListPrice:          100,
		SalePrice:          float64PointerForServiceTest(90),
		PurchasePrice:      float64PointerForServiceTest(50),
		PurchasePriceKnown: true,
	}})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, procurementdomain.ProfitStatusReady, results[0].Status)
	require.Equal(t, 40.0, *results[0].GrossProfit)

	var count int64
	require.NoError(t, db.Model(&procurementdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityServiceBulkUpsertSkipsUnknownPurchaseAndPersistsExplicitZero(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	result, err := service.BulkUpsert([]ProfitabilityItemInput{
		{
			ProductCode:     "SKU-MISSING",
			ProductName:     "Missing cost",
			SellingCurrency: "USD",
			ListPrice:       100,
		},
		{
			ProductCode:        "SKU-ZERO",
			ProductName:        "Zero cost",
			SellingCurrency:    "USD",
			ListPrice:          100,
			PurchasePrice:      float64PointerForServiceTest(0),
			PurchasePriceKnown: true,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Records, 1)
	require.Len(t, result.Skipped, 1)
	require.Equal(t, "SKU-MISSING", result.Skipped[0].ProductCode)
	require.Equal(t, procurementdomain.ProfitStatusMissingPurchase, result.Skipped[0].Status)
	require.Equal(t, "SKU-ZERO", result.Records[0].ProductCode)
	require.Equal(t, 0.0, result.Records[0].PurchasePrice)
	require.Equal(t, 100.0, result.Records[0].GrossProfit)

	_, err = repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-MISSING")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductProfitabilityServiceClearsOldSnapshotWhenPurchasePriceBecomesUnknown(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	_, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:        "SKU-CLEAR",
		ProductName:        "Clear cost",
		SellingCurrency:    "USD",
		ListPrice:          100,
		PurchasePrice:      float64PointerForServiceTest(40),
		PurchasePriceKnown: true,
		Procurement: &ProfitabilityProcurementInput{
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
		ListPrice:       100,
	}})
	require.NoError(t, err)
	require.Len(t, result.Records, 0)
	require.Len(t, result.Skipped, 1)
	require.Equal(t, "SKU-CLEAR", result.Skipped[0].ProductCode)

	_, err = repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-CLEAR")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repository.NewProductProcurementRepository(db).FindByProductCode("SKU-CLEAR")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductProfitabilityServiceBulkUpsertUpdatesSameSKU(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	first, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:             "SKU-UPDATE",
		ProductName:             "First name",
		SellingCurrency:         "USD",
		ListPrice:               100,
		PurchasePrice:           float64PointerForServiceTest(60),
		PurchasePriceKnown:      true,
		InboundShippingUnitCost: 1,
		PackagingUnitCost:       3,
		OtherUnitCost:           4,
	}})
	require.NoError(t, err)
	require.Len(t, first.Records, 1)
	firstID := first.Records[0].ID

	second, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:             "SKU-UPDATE",
		ProductName:             "Second name",
		SellingCurrency:         "USD",
		ListPrice:               100,
		SalePrice:               float64PointerForServiceTest(90),
		PurchasePrice:           float64PointerForServiceTest(40),
		PurchasePriceKnown:      true,
		InboundShippingUnitCost: 5,
		PackagingUnitCost:       7,
		OtherUnitCost:           8,
	}})
	require.NoError(t, err)
	require.Len(t, second.Records, 1)
	require.Equal(t, firstID, second.Records[0].ID)
	require.Equal(t, "Second name", second.Records[0].ProductName)
	require.Equal(t, 30.0, second.Records[0].GrossProfit)
	require.Equal(t, 5.0, second.Records[0].InboundShippingUnitCost)
	require.Equal(t, 7.0, second.Records[0].PackagingUnitCost)
	require.Equal(t, 8.0, second.Records[0].OtherUnitCost)
	require.Equal(t, 60.0, second.Records[0].LandedCost)

	var count int64
	require.NoError(t, db.Model(&procurementdomain.ProductProfitCalculation{}).
		Where("product_code = ?", "SKU-UPDATE").
		Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestProductProfitabilityServiceInvalidBatchDoesNotWriteValidItems(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	_, err := service.BulkUpsert([]ProfitabilityItemInput{
		{
			ProductCode:        "SKU-VALID",
			ProductName:        "Valid item",
			SellingCurrency:    "USD",
			ListPrice:          100,
			PurchasePrice:      float64PointerForServiceTest(40),
			PurchasePriceKnown: true,
		},
		{
			ProductCode:        "SKU-INVALID",
			ProductName:        "Invalid item",
			SellingCurrency:    "USD",
			CostCurrency:       "CNY",
			ListPrice:          100,
			PurchasePrice:      float64PointerForServiceTest(40),
			PurchasePriceKnown: true,
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrProductProfitabilityInvalid)

	var count int64
	require.NoError(t, db.Model(&procurementdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)

	_, err = repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-VALID")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProductProfitabilityServiceRejectsDuplicateCodesWithoutWriting(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityService(repository.NewProductProfitCalculationRepository(db))

	_, err := service.BulkUpsert([]ProfitabilityItemInput{
		{
			ProductCode:        "SKU-DUPLICATE",
			ProductName:        "First item",
			SellingCurrency:    "USD",
			ListPrice:          100,
			PurchasePrice:      float64PointerForServiceTest(40),
			PurchasePriceKnown: true,
		},
		{
			ProductCode:        " SKU-DUPLICATE ",
			ProductName:        "Second item",
			SellingCurrency:    "USD",
			ListPrice:          100,
			PurchasePrice:      float64PointerForServiceTest(35),
			PurchasePriceKnown: true,
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrProductProfitabilityInvalid)
	var validationErr *ProfitabilityBatchValidationError
	require.ErrorAs(t, err, &validationErr)
	require.Len(t, validationErr.Items, 1)
	require.Contains(t, validationErr.Items[0].Reason, "duplicates item 1")

	var count int64
	require.NoError(t, db.Model(&procurementdomain.ProductProfitCalculation{}).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityServicePreviewTreatsUnmarkedPurchasePriceAsUnknown(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityService(repository.NewProductProfitCalculationRepository(db))

	results, err := service.Preview([]ProfitabilityItemInput{{
		ProductCode:     "SKU-MISSING-FLAG",
		ProductName:     "Missing flag",
		SellingCurrency: "USD",
		ListPrice:       100,
		PurchasePrice:   float64PointerForServiceTest(20),
	}})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, procurementdomain.ProfitStatusMissingPurchase, results[0].Status)
	require.Nil(t, results[0].GrossProfit)
}

func TestProductProfitabilityServiceBulkUpsertPersistsProcurementAndProfitTogether(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	result, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:             "SKU-PROCUREMENT-ATOMIC",
		ProductName:             "Atomic procurement item",
		SellingCurrency:         "USD",
		ListPrice:               100,
		PurchasePrice:           float64PointerForServiceTest(40),
		PurchasePriceKnown:      true,
		InboundShippingUnitCost: 2,
		PackagingUnitCost:       4,
		OtherUnitCost:           5,
		Procurement: &ProfitabilityProcurementInput{
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

	procurementRecord, err := repository.NewProductProcurementRepository(db).FindByProductCode("SKU-PROCUREMENT-ATOMIC")
	require.NoError(t, err)
	require.Equal(t, "Atomic Supplier", procurementRecord.SupplierName)
	require.Equal(t, 40.0, procurementRecord.PurchasePrice)
	require.Equal(t, 2.0, procurementRecord.InboundShippingUnitCost)
	require.Equal(t, 4.0, procurementRecord.PackagingUnitCost)
	require.Equal(t, 5.0, procurementRecord.OtherUnitCost)
	require.Equal(t, 14, procurementRecord.LeadTimeDays)
	require.Equal(t, 20, procurementRecord.MinimumOrderQuantity)

	profitRecord, err := repository.NewProductProfitCalculationRepository(db).FindByProductCode("SKU-PROCUREMENT-ATOMIC")
	require.NoError(t, err)
	require.Equal(t, 49.0, profitRecord.GrossProfit)
}

func TestProductProfitabilityServiceBulkUpsertRollsBackProfitWhenProcurementWriteFails(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)
	require.NoError(t, db.Migrator().DropTable(&procurementdomain.ProductProcurement{}))

	_, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:        "SKU-PROCUREMENT-ROLLBACK",
		ProductName:        "Rollback item",
		SellingCurrency:    "USD",
		ListPrice:          100,
		PurchasePrice:      float64PointerForServiceTest(40),
		PurchasePriceKnown: true,
		Procurement: &ProfitabilityProcurementInput{
			SupplierName: "Unavailable supplier",
		},
	}})
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&procurementdomain.ProductProfitCalculation{}).
		Where("product_code = ?", "SKU-PROCUREMENT-ROLLBACK").
		Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestProductProfitabilityServiceBulkUpsertUpdatesProcurementBySKU(t *testing.T) {
	db := newProductProfitabilityServiceTestDB(t)
	service := NewProductProfitabilityServiceWithProcurement(
		repository.NewProductProfitCalculationRepository(db),
		repository.NewProductProcurementRepository(db),
	)

	first, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:             "SKU-PROCUREMENT-UPDATE",
		ProductName:             "First item",
		SellingCurrency:         "USD",
		ListPrice:               100,
		PurchasePrice:           float64PointerForServiceTest(60),
		PurchasePriceKnown:      true,
		InboundShippingUnitCost: 1,
		PackagingUnitCost:       3,
		OtherUnitCost:           4,
		Procurement: &ProfitabilityProcurementInput{
			SupplierName: "First supplier",
		},
	}})
	require.NoError(t, err)
	require.Len(t, first.Records, 1)

	second, err := service.BulkUpsert([]ProfitabilityItemInput{{
		ProductCode:             "SKU-PROCUREMENT-UPDATE",
		ProductName:             "Second item",
		SellingCurrency:         "USD",
		ListPrice:               100,
		PurchasePrice:           float64PointerForServiceTest(50),
		PurchasePriceKnown:      true,
		InboundShippingUnitCost: 5,
		PackagingUnitCost:       7,
		OtherUnitCost:           8,
		Procurement: &ProfitabilityProcurementInput{
			SupplierName: "Second supplier",
			LeadTimeDays: 7,
		},
	}})
	require.NoError(t, err)
	require.Len(t, second.Records, 1)

	procurementRecord, err := repository.NewProductProcurementRepository(db).FindByProductCode("SKU-PROCUREMENT-UPDATE")
	require.NoError(t, err)
	require.Equal(t, "Second supplier", procurementRecord.SupplierName)
	require.Equal(t, 5.0, procurementRecord.InboundShippingUnitCost)
	require.Equal(t, 7.0, procurementRecord.PackagingUnitCost)
	require.Equal(t, 8.0, procurementRecord.OtherUnitCost)
	require.Equal(t, 7, procurementRecord.LeadTimeDays)

	var count int64
	require.NoError(t, db.Model(&procurementdomain.ProductProcurement{}).
		Where("product_code = ?", "SKU-PROCUREMENT-UPDATE").
		Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func newProductProfitabilityServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProfitCalculation{}))
	require.NoError(t, db.AutoMigrate(&procurementdomain.ProductProcurement{}))
	return db
}

func float64PointerForServiceTest(value float64) *float64 {
	return &value
}
