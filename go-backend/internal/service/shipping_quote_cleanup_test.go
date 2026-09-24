package service

import (
	"testing"
	"time"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCleanupExpiredShippingQuoteSnapshotsRemovesExpiredRowsOnly(t *testing.T) {
	db := newShippingQuoteCleanupTestDB(t)
	repo := repository.NewShippingRepository(db)
	service := NewShippingService(repo)
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&shippingdomain.QuoteSnapshot{
		ID: "expired", RequestHash: "expired-request", RateVersion: "v1",
		QuoteData: []byte(`{"quote":"expired"}`), ExpiresAt: now.Add(-time.Second), CreatedAt: now.Add(-time.Hour),
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.QuoteSnapshot{
		ID: "active", RequestHash: "active-request", RateVersion: "v1",
		QuoteData: []byte(`{"quote":"active"}`), ExpiresAt: now.Add(time.Hour), CreatedAt: now,
	}).Error)

	deleted, err := service.CleanupExpiredShippingQuoteSnapshots(now, 10)
	require.NoError(t, err)
	require.EqualValues(t, 1, deleted)

	var remaining []shippingdomain.QuoteSnapshot
	require.NoError(t, db.Find(&remaining).Error)
	require.Len(t, remaining, 1)
	require.Equal(t, "active", remaining[0].ID)
}

func TestCleanupExpiredShippingQuoteSnapshotsHonorsBatchLimit(t *testing.T) {
	db := newShippingQuoteCleanupTestDB(t)
	repo := repository.NewShippingRepository(db)
	service := NewShippingService(repo)
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)

	for i, id := range []string{"expired-1", "expired-2", "expired-3"} {
		require.NoError(t, db.Create(&shippingdomain.QuoteSnapshot{
			ID: id, RequestHash: "request-" + id, RateVersion: "v1",
			QuoteData: []byte(`{"quote":"expired"}`), ExpiresAt: now.Add(-time.Duration(i+1) * time.Minute), CreatedAt: now.Add(-time.Hour),
		}).Error)
	}

	deleted, err := service.CleanupExpiredShippingQuoteSnapshots(now, 2)
	require.NoError(t, err)
	require.EqualValues(t, 2, deleted)

	var remaining []shippingdomain.QuoteSnapshot
	require.NoError(t, db.Find(&remaining).Error)
	require.Len(t, remaining, 1)
	// The repository deletes the oldest expired rows first. Assert by ID
	// instead of relying on the database's unspecified result ordering.
	var survivor shippingdomain.QuoteSnapshot
	require.NoError(t, db.First(&survivor, "id = ?", "expired-1").Error)
	require.Equal(t, "expired-1", survivor.ID)
}

func newShippingQuoteCleanupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shippingdomain.QuoteSnapshot{}))
	return db
}
