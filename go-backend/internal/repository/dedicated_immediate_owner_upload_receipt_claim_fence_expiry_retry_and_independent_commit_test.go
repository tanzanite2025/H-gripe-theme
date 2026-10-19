package repository

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createDedicatedImmediateOwnerUploadReceiptFenceTestRepository(t *testing.T) *OutboxRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	return NewOutboxRepository(db)
}

func newDedicatedImmediateOwnerUploadReceiptForFenceTest() *outbox.Event {
	return &outbox.Event{EventKey: DedicatedImmediateOwnerUploadReceiptEventKeyPrefix + uuid.NewString(), EventType: outbox.EventTypeObjectStorageCleanup,
		AggregateType: outbox.AggregateTypeSiteLogo, AggregateID: "receipt-fence-test", Payload: []byte(`{"object_keys":["site-logo/preallocated.webp"]}`), AvailableAt: time.Now().UTC().Add(5 * time.Minute)}
}

func TestDedicatedImmediateOwnerUploadRejectsLateWriterAfterCleanupClaimCompletionExpiryOrManualRetry(t *testing.T) {
	for _, state := range []string{"claimed", "completed", "expired", "failed", "manually retried", "missing receipt", "changed payload"} {
		t.Run(state, func(t *testing.T) {
			repo := createDedicatedImmediateOwnerUploadReceiptFenceTestRepository(t)
			receipt := newDedicatedImmediateOwnerUploadReceiptForFenceTest()
			require.NoError(t, repo.CreateEvent(receipt))
			switch state {
			case "claimed", "completed", "failed", "manually retried":
				claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "cleanup-worker", 10, time.Minute)
				require.NoError(t, err)
				require.Len(t, claimed, 1)
				if state == "completed" {
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "cleanup-worker", time.Now()))
				}
				if state == "failed" || state == "manually retried" {
					require.NoError(t, repo.MarkFailedByWorker(receipt.ID, "cleanup-worker", outbox.EventStatusFailed, "retry", time.Now(), time.Now()))
				}
				if state == "manually retried" {
					require.NoError(t, repo.RetryFailedEvent(receipt.ID, "retry cleanup", time.Now().Add(time.Minute)))
				}
			case "expired":
				require.NoError(t, repo.db.Model(receipt).Update("available_at", time.Now().Add(-time.Second)).Error)
			case "missing receipt":
				require.NoError(t, repo.db.Unscoped().Delete(receipt).Error)
			case "changed payload":
				require.NoError(t, repo.db.Model(&outbox.Event{}).Where("id = ?", receipt.ID).Update("payload", []byte(`{"object_keys":["site-logo/foreign.webp"]}`)).Error)
			}
			writes := 0
			err := repo.runDedicatedImmediateOwnerUploadWithLockedUnclaimedReceipt(t.Context(), dedicatedSiteLogoObjectNamespace, receipt, func(*gorm.DB) error { writes++; return nil })
			require.ErrorIs(t, err, ErrDedicatedImmediateOwnerUploadReceiptNoLongerWritable)
			require.Zero(t, writes)
		})
	}
}

func TestDedicatedImmediateOwnerUploadRejectsOuterTransactionBeforeReceiptPersistenceOrProviderIO(t *testing.T) {
	repo := createDedicatedImmediateOwnerUploadReceiptFenceTestRepository(t)
	writes := 0
	require.NoError(t, repo.db.Transaction(func(tx *gorm.DB) error {
		err := repo.WithTx(tx).RunDedicatedImmediateOwnerUploadWithPrecommittedCleanupReceipt(context.Background(), dedicatedSiteLogoObjectNamespace, newDedicatedImmediateOwnerUploadReceiptForFenceTest(), func(*gorm.DB) error { writes++; return nil })
		require.ErrorContains(t, err, "independent committed transaction")
		return nil
	}))
	require.Zero(t, writes)
	var count int64
	require.NoError(t, repo.db.Model(&outbox.Event{}).Count(&count).Error)
	require.Zero(t, count)
}
