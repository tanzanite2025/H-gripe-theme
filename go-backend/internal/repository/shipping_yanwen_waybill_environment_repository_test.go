package repository

import (
	"testing"

	"commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenWaybillRepositoryScopesIdempotencyByEnvironment(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenWaybill{}))

	repo := NewYanwenWaybillRepository(db)
	fatWaybill := newEnvironmentScopedYanwenWaybill("fat", "YE-FAT-481")
	productionWaybill := newEnvironmentScopedYanwenWaybill("production", "YE-PRODUCTION-481")
	require.NoError(t, repo.CreateYanwenWaybill(&fatWaybill))
	require.NoError(t, repo.CreateYanwenWaybill(&productionWaybill))

	foundFat, err := repo.FindYanwenWaybillByEnvironmentAndOrderIDAndProductCode("fat", fatWaybill.OrderID, fatWaybill.ProductCode)
	require.NoError(t, err)
	require.Equal(t, fatWaybill.WaybillNumber, foundFat.WaybillNumber)

	foundProduction, err := repo.FindYanwenWaybillByEnvironmentAndOrderIDAndProductCode("production", productionWaybill.OrderID, productionWaybill.ProductCode)
	require.NoError(t, err)
	require.Equal(t, productionWaybill.WaybillNumber, foundProduction.WaybillNumber)
}

func newEnvironmentScopedYanwenWaybill(environment, waybillNumber string) shipping.YanwenWaybill {
	return shipping.YanwenWaybill{
		Environment: environment, OrderID: 23, OrderNumber: "ORDER-ENVIRONMENT-SCOPE",
		ProductCode: "481", ChannelName: "燕文普货", WarehouseCode: "WH01",
		DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Bicycle spokes",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: waybillNumber,
		Status: shipping.YanwenWaybillStatusCreated,
	}
}
