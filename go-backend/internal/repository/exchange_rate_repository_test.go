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
