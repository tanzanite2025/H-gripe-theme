package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenWarehouseCatalogRepositoryUpsertsWarehousesPerEnvironment(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenWarehouseCatalogEntry{}))

	repository := NewYanwenWarehouseCatalogRepository(db)
	syncedAt := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	stats, err := repository.UpsertYanwenWarehouseCatalogEntries("fat", []shipping.YanwenWarehouseCatalogEntry{
		{WarehouseCode: "01", Name: "北京燕文", Area: "华北"},
		{WarehouseCode: "02", Name: "上海燕文", Area: "华东"},
	}, syncedAt)
	require.NoError(t, err)
	require.Equal(t, YanwenWarehouseCatalogSyncStats{Scanned: 2, Added: 2, Updated: 0}, stats)

	stats, err = repository.UpsertYanwenWarehouseCatalogEntries("fat", []shipping.YanwenWarehouseCatalogEntry{
		{WarehouseCode: "01", Name: "北京燕文更新", Area: "华北"},
	}, syncedAt.Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, YanwenWarehouseCatalogSyncStats{Scanned: 1, Added: 0, Updated: 1}, stats)

	entries, err := repository.FindYanwenWarehouseCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "北京燕文更新", entries[0].Name)
	require.Equal(t, "华北", entries[0].Area)

	productionEntries, err := repository.FindYanwenWarehouseCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	require.Empty(t, productionEntries)
}
