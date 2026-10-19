package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOriginalStorageCleanupReconciliationSelectsOnlySurvivingAuthorizedTerminalReceipts(t *testing.T) {
	for _, eventType := range []string{outbox.EventTypeObjectStorageCleanup, outbox.EventTypeCustomerServiceAvatarCleanup, outbox.EventTypePaymentRefundExecutionRequested} {
		for _, status := range []string{"pending", "processing", "failed", "processed", "unknown", "dead_letter", "ignored"} {
			t.Run(eventType+"/"+status, func(t *testing.T) {
				db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				require.NoError(t, err)
				sqlDB, err := db.DB()
				require.NoError(t, err)
				t.Cleanup(func() { _ = sqlDB.Close() })
				require.NoError(t, db.AutoMigrate(&outbox.Event{}))
				now := time.Now().UTC()
				event := outbox.Event{EventKey: "original", EventType: eventType, AggregateType: "test", AggregateID: "1", Status: status, AvailableAt: now, CreatedAt: now, UpdatedAt: now}
				require.NoError(t, db.Create(&event).Error)
				claimed, err := NewOutboxRepository(db).ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now.Add(time.Second), "unique-pass", 10, time.Minute)
				require.NoError(t, err)
				if IsObjectStorageCleanupReconciliationEligible(event) {
					require.Len(t, claimed, 1)
				} else {
					require.Empty(t, claimed)
				}
			})
		}
	}
}

func TestOriginalStorageCleanupReconciliationRetainsFutureRetryAndOriginalWriterFenceAcrossLeaseRecovery(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	now := time.Now().UTC()
	originalAttempt := now.Add(-time.Hour)
	event := outbox.Event{EventKey: "original", EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "media_asset", AggregateID: "1", Payload: []byte(`{"object_keys":["exact/original.png"]}`), Status: "processed", Attempts: 1_000_001, MaxAttempts: 1_000_000, LastAttemptAt: &originalAttempt, ProcessedAt: &originalAttempt, LastError: "original dispatch evidence", UpdatedAt: originalAttempt, CreatedAt: originalAttempt}
	require.NoError(t, db.Create(&event).Error)
	repo := NewOutboxRepository(db)
	claimed, err := repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now, "first-pass", 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	claimed, err = repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now.Add(30*time.Second), "other-pass", 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)
	claimed, err = repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now.Add(2*time.Minute), "replacement-pass", 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.ErrorIs(t, repo.RefreshOriginalObjectStorageCleanupReconciliationLease(event.ID, "first-pass", now), ErrOutboxOwnershipLost)
	require.ErrorIs(t, repo.CompleteOriginalObjectStorageCleanupReconciliation(event.ID, "first-pass", now, now.Add(time.Hour), ""), ErrOutboxOwnershipLost)
	require.NoError(t, repo.RefreshOriginalObjectStorageCleanupReconciliationLease(event.ID, "replacement-pass", now.Add(2*time.Minute)))
	require.Error(t, repo.CompleteOriginalObjectStorageCleanupReconciliation(event.ID, "replacement-pass", now, now, ""))
	require.NoError(t, repo.CompleteOriginalObjectStorageCleanupReconciliation(event.ID, "replacement-pass", now.Add(3*time.Minute), now.Add(time.Hour), "historical reference retained"))
	var saved outbox.Event
	require.NoError(t, db.First(&saved, event.ID).Error)
	require.Equal(t, event.Payload, saved.Payload)
	require.Equal(t, event.Status, saved.Status)
	require.Equal(t, event.Attempts, saved.Attempts)
	require.WithinDuration(t, *event.LastAttemptAt, *saved.LastAttemptAt, time.Microsecond)
	require.WithinDuration(t, event.UpdatedAt, saved.UpdatedAt, time.Microsecond)
	require.Equal(t, event.LastError, saved.LastError)
	require.Equal(t, "historical reference retained", saved.StorageCleanupReconciliationError)
	require.Empty(t, saved.LockedBy)
	claimed, err = repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now.Add(59*time.Minute), "too-early", 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)
	// Soft deletion represents consumed ownership; never reconstruct it.
	require.NoError(t, db.Delete(&saved).Error)
	claimed, err = repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now.Add(2*time.Hour), "deleted", 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)
}
