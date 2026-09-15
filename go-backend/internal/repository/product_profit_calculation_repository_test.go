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
		ProductCode:             " SKU-001 ",
		ProductName:             "Initial name",
		Currency:                "usd",
		ListPrice:               100,
		SalePrice:               float64PointerForRepositoryTest(90),
		EffectiveSellingPrice:   90,
		UnitCost:                50,
		InboundShippingUnitCost: 2,
		PackagingUnitCost:       0.5,
		OtherUnitCost:           0.25,
		LandedCost:              50,
		GrossProfit:             40,
		GrossMarginBPS:          4444,
		CalculationStatus:       suppliercostdomain.ProfitStatusReady,
		FormulaVersion:          suppliercostdomain.ProfitFormulaVersion,
		WarningsData:            datatypes.JSON([]byte(`[]`)),
		CalculatedAt:            time.Now().UTC(),
	}
	require.NoError(t, repo.BulkUpsert([]suppliercostdomain.ProductProfitCalculation{first}))

	updated := first
	updated.ProductName = "Updated name"
	updated.UnitCost = 60
	updated.InboundShippingUnitCost = 4
	updated.PackagingUnitCost = 2
	updated.OtherUnitCost = 1
	updated.LandedCost = 70
	updated.GrossProfit = 20
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
	require.Equal(t, 60.0, record.UnitCost)
	require.Equal(t, 4.0, record.InboundShippingUnitCost)
	require.Equal(t, 2.0, record.PackagingUnitCost)
	require.Equal(t, 1.0, record.OtherUnitCost)
	require.Equal(t, 70.0, record.LandedCost)
	require.Equal(t, 20.0, record.GrossProfit)
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
		ProductCode:           "SKU-ISOLATED",
		ProductName:           "Isolated item",
		Currency:              "USD",
		ListPrice:             10,
		EffectiveSellingPrice: 10,
		UnitCost:              5,
		LandedCost:            5,
		GrossProfit:           5,
		GrossMarginBPS:        5000,
		CalculationStatus:     suppliercostdomain.ProfitStatusReady,
		FormulaVersion:        suppliercostdomain.ProfitFormulaVersion,
		WarningsData:          datatypes.JSON([]byte(`[]`)),
		CalculatedAt:          time.Now().UTC(),
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
		ProductCode:           "SKU-CLEAR",
		ProductName:           "Clear cost",
		Currency:              "USD",
		ListPrice:             100,
		EffectiveSellingPrice: 100,
		UnitCost:              40,
		LandedCost:            40,
		GrossProfit:           60,
		GrossMarginBPS:        6000,
		CalculationStatus:     suppliercostdomain.ProfitStatusReady,
		FormulaVersion:        suppliercostdomain.ProfitFormulaVersion,
		WarningsData:          datatypes.JSON([]byte(`[]`)),
		CalculatedAt:          time.Now().UTC(),
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

func float64PointerForRepositoryTest(value float64) *float64 {
	return &value
}
