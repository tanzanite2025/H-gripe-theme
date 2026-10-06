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

func TestYanwenProductCatalogRepositoryUpsertsProductsPerEnvironment(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenProductCatalogEntry{}))

	productCatalogRepository := NewYanwenProductCatalogRepository(db)
	syncedAt := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	stats, err := productCatalogRepository.UpsertYanwenProductCatalogEntries("fat", []shipping.YanwenProductCatalogEntry{
		{ProductID: "481", NameChinese: "普货", NameEnglish: "Tracked"},
		{ProductID: "484", NameChinese: "特货", NameEnglish: "Sensitive"},
	}, syncedAt)
	require.NoError(t, err)
	require.Equal(t, YanwenProductCatalogSyncStats{Scanned: 2, Added: 2, Updated: 0}, stats)

	stats, err = productCatalogRepository.UpsertYanwenProductCatalogEntries("fat", []shipping.YanwenProductCatalogEntry{
		{ProductID: "481", NameChinese: "普货更新", NameEnglish: "Tracked Updated"},
		{ProductID: "484", NameChinese: "特货", NameEnglish: "Sensitive"},
	}, syncedAt.Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, YanwenProductCatalogSyncStats{Scanned: 2, Added: 0, Updated: 2}, stats)

	fatEntries, err := productCatalogRepository.FindYanwenProductCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, fatEntries, 2)
	require.Equal(t, "普货更新", fatEntries[0].NameChinese)
	require.True(t, fatEntries[0].LastSyncedAt.Equal(syncedAt.Add(time.Hour)))

	stats, err = productCatalogRepository.UpsertYanwenProductCatalogEntries("fat", []shipping.YanwenProductCatalogEntry{
		{ProductID: "481", NameChinese: "普货最终名称", NameEnglish: "Tracked Final"},
	}, syncedAt.Add(2*time.Hour))
	require.NoError(t, err)
	require.Equal(t, YanwenProductCatalogSyncStats{Scanned: 1, Added: 0, Updated: 1}, stats)
	fatEntries, err = productCatalogRepository.FindYanwenProductCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, fatEntries, 1, "products absent from the latest official response must be removed")
	require.Equal(t, "481", fatEntries[0].ProductID)

	productionEntries, err := productCatalogRepository.FindYanwenProductCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	require.Empty(t, productionEntries)
}
