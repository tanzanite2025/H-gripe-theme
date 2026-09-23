package repository

import (
	"testing"
	"time"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestProductProfitCalculationRepositoryBulkUpsertUsesProductCodeAsStableKey(t *testing.T) {
	db := newProductProfitCalculationTestDB(t)
	repo := NewProductProfitCalculationRepository(db)

	first := suppliercostdomain.ProductProfitCalculation{
		ProductCode:                  " SKU-001 ",
		ProductName:                  "Initial name",
		Currency:                     "usd",
		ListPriceMinor:               10000,
		SalePriceMinor:               int64PointerForRepositoryTest(9000),
		EffectiveSellingPriceMinor:   9000,
		UnitCostMinor:                5000,
		InboundShippingUnitCostMinor: 200,
		PackagingUnitCostMinor:       50,
		OtherUnitCostMinor:           25,
		LandedCostMinor:              5000,
		GrossProfitMinor:             4000,
		GrossMarginBPS:               4444,
		CalculationStatus:            suppliercostdomain.ProfitStatusReady,
		FormulaVersion:               suppliercostdomain.ProfitFormulaVersion,
		WarningsData:                 datatypes.JSON([]byte(`[]`)),
		CalculatedAt:                 time.Now().UTC(),
	}
	require.NoError(t, repo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{first}))

	updated := first
	updated.ProductName = "Updated name"
	updated.UnitCostMinor = 6000
	updated.InboundShippingUnitCostMinor = 400
	updated.PackagingUnitCostMinor = 200
	updated.OtherUnitCostMinor = 100
	updated.LandedCostMinor = 7000
	updated.GrossProfitMinor = 2000
	updated.GrossMarginBPS = 3333
	updated.CalculatedAt = time.Now().UTC().Add(time.Minute)
	require.NoError(t, repo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{updated}))

	var count int64
	require.NoError(t, db.Model(&suppliercostdomain.ProductProfitCalculation{}).
		Where("product_code = ?", "SKU-001").
		Count(&count).Error)
	require.Equal(t, int64(1), count)

	record, err := repo.FindByProductCode(" SKU-001 ")
	require.NoError(t, err)
	require.Equal(t, "SKU-001", record.ProductCode)
	require.Equal(t, "Updated name", record.ProductName)
	require.Equal(t, int64(6000), record.UnitCostMinor)
	require.Equal(t, int64(400), record.InboundShippingUnitCostMinor)
	require.Equal(t, int64(200), record.PackagingUnitCostMinor)
	require.Equal(t, int64(100), record.OtherUnitCostMinor)
	require.Equal(t, int64(7000), record.LandedCostMinor)
	require.Equal(t, int64(2000), record.GrossProfitMinor)
	require.Equal(t, "USD", record.Currency)

	records, err := repo.FindByProductCodes([]string{" SKU-001 ", "SKU-001", "", " "})
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "SKU-001", records[0].ProductCode)
}

func TestProductProfitCalculationRepositoryDoesNotRequireCatalogTables(t *testing.T) {
	db := newProductProfitCalculationTestDB(t)
	repo := NewProductProfitCalculationRepository(db)

	require.NoError(t, repo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{{
		ProductCode:                "SKU-ISOLATED",
		ProductName:                "Isolated item",
		Currency:                   "USD",
		ListPriceMinor:             1000,
		EffectiveSellingPriceMinor: 1000,
		UnitCostMinor:              500,
		LandedCostMinor:            500,
		GrossProfitMinor:           500,
		GrossMarginBPS:             5000,
		CalculationStatus:          suppliercostdomain.ProfitStatusReady,
		FormulaVersion:             suppliercostdomain.ProfitFormulaVersion,
		WarningsData:               datatypes.JSON([]byte(`[]`)),
		CalculatedAt:               time.Now().UTC(),
	}}))

	var tableCount int64
	require.NoError(t, db.Raw(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('products', 'product_variants')",
	).Scan(&tableCount).Error)
	require.Equal(t, int64(0), tableCount)
}

func TestProductProfitCalculationRepositoryReplaceCurrentSnapshotsClearsUnknownCostCodes(t *testing.T) {
	db := newProductProfitCalculationTestDB(t)
	repo := NewProductProfitCalculationRepository(db)

	record := suppliercostdomain.ProductProfitCalculation{
		ProductCode:                "SKU-CLEAR",
		ProductName:                "Clear cost",
		Currency:                   "USD",
		ListPriceMinor:             10000,
		EffectiveSellingPriceMinor: 10000,
		UnitCostMinor:              4000,
		LandedCostMinor:            4000,
		GrossProfitMinor:           6000,
		GrossMarginBPS:             6000,
		CalculationStatus:          suppliercostdomain.ProfitStatusReady,
		FormulaVersion:             suppliercostdomain.ProfitFormulaVersion,
		WarningsData:               datatypes.JSON([]byte(`[]`)),
		CalculatedAt:               time.Now().UTC(),
	}
	require.NoError(t, repo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{record}))
	require.NoError(t, repo.ReplaceCurrentSnapshots(nil, []string{" SKU-CLEAR "}))

	_, err := repo.FindByProductCode("SKU-CLEAR")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func newProductProfitCalculationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&suppliercostdomain.ProductProfitCalculation{}))
	return db
}

func int64PointerForRepositoryTest(value int64) *int64 {
	return &value
}
