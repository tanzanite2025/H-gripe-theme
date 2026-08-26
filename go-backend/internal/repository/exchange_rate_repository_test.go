package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExchangeRateSyncLeaseAllowsSingleActiveOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&currency.ExchangeRateSyncLease{}))

	repo := NewExchangeRateRepository(db)
	now := time.Now().UTC()

	acquired, err := repo.TryAcquireSyncLease("daily", "worker-a", now, time.Hour)
	require.NoError(t, err)
	require.True(t, acquired)

	acquired, err = repo.TryAcquireSyncLease("daily", "worker-b", now.Add(time.Minute), time.Hour)
	require.NoError(t, err)
	require.False(t, acquired)

	acquired, err = repo.TryAcquireSyncLease("daily", "worker-a", now.Add(time.Minute), time.Hour)
	require.NoError(t, err)
	require.True(t, acquired)

	acquired, err = repo.TryAcquireSyncLease("daily", "worker-b", now.Add(2*time.Hour), time.Hour)
	require.NoError(t, err)
	require.True(t, acquired)

	require.NoError(t, repo.ReleaseSyncLease("daily", "worker-b", now.Add(2*time.Hour)))
	acquired, err = repo.TryAcquireSyncLease("daily", "worker-c", now.Add(2*time.Hour+time.Second), time.Hour)
	require.NoError(t, err)
	require.True(t, acquired)
}

func TestExchangeRateUpsertMatchesSoftDeletePartialUniqueIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.Exec(`
		CREATE TABLE currency_exchange_rates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			base_currency TEXT NOT NULL,
			quote_currency TEXT NOT NULL,
			rate REAL NOT NULL,
			source TEXT NOT NULL DEFAULT '',
			fetched_at DATETIME NOT NULL,
			expires_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX idx_currency_exchange_rate_pair
		ON currency_exchange_rates(base_currency, quote_currency)
		WHERE deleted_at IS NULL
	`).Error)

	repo := NewExchangeRateRepository(db)
	fetchedAt := time.Now().UTC()
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{{
		BaseCurrency:  "CNY",
		QuoteCurrency: "USD",
		Rate:          7.1,
		Source:        "test",
		FetchedAt:     fetchedAt,
	}}))
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{{
		BaseCurrency:  "CNY",
		QuoteCurrency: "USD",
		Rate:          7.2,
		Source:        "test-refresh",
		FetchedAt:     fetchedAt.Add(time.Hour),
	}}))

	var stored currency.ExchangeRate
	require.NoError(t, db.First(&stored, "base_currency = ? AND quote_currency = ?", "CNY", "USD").Error)
	require.Equal(t, 7.2, stored.Rate)
	require.Equal(t, "test-refresh", stored.Source)

	var count int64
	require.NoError(t, db.Model(&currency.ExchangeRate{}).
		Where("base_currency = ? AND quote_currency = ?", "CNY", "USD").
		Count(&count).Error)
	require.Equal(t, int64(1), count)
}
