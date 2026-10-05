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

func TestYanwenCountryCatalogRepositoryUpsertsCountriesPerEnvironment(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenCountryCatalogEntry{}))

	repository := NewYanwenCountryCatalogRepository(db)
	syncedAt := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	stats, err := repository.UpsertYanwenCountryCatalogEntries("fat", []shipping.YanwenCountryCatalogEntry{
		{CountryID: "1", CountryCode: "AW", NameChinese: "阿鲁巴", NameEnglish: "ARUBA"},
		{CountryID: "45", CountryCode: "CA", NameChinese: "加拿大", NameEnglish: "CANADA"},
	}, syncedAt)
	require.NoError(t, err)
	require.Equal(t, YanwenCountryCatalogSyncStats{Scanned: 2, Added: 2, Updated: 0}, stats)

	stats, err = repository.UpsertYanwenCountryCatalogEntries("fat", []shipping.YanwenCountryCatalogEntry{
		{CountryID: "1", CountryCode: "AW", NameChinese: "阿鲁巴更新", NameEnglish: "ARUBA"},
	}, syncedAt.Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, YanwenCountryCatalogSyncStats{Scanned: 1, Added: 0, Updated: 1}, stats)

	entries, err := repository.FindYanwenCountryCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "阿鲁巴更新", entries[0].NameChinese)
	require.Equal(t, "AW", entries[0].CountryCode)

	productionEntries, err := repository.FindYanwenCountryCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	require.Empty(t, productionEntries)
}
