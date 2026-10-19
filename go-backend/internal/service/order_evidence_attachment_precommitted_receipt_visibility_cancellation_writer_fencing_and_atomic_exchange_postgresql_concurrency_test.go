package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLOrderEvidencePrecommittedReceiptVisibleBeforeIOAndCancellationHoldsLocksUntilProviderReturns(t *testing.T) {
	env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, releaseChannel := make(chan struct{}), make(chan struct{})
	var releaseOnce, captureOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
	defer release()
	pidChannel := make(chan int, 1)
	callback := "capture_evidence_receipt_writer_backend_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			if _, active := tx.Statement.ConnPool.(gorm.TxCommitter); active {
				captureOnce.Do(func() {
					var pid int
					if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
						tx.AddError(err)
					} else {
						pidChannel <- pid
					}
				})
			}
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callback) })
	env.provider.afterPreallocatedWrite = func(context.Context, string) { close(entered); <-releaseChannel }
	file := newEvidencePNGFileHeader(t, "proof.png")
	result := make(chan error, 1)
	go func() { _, err := env.service.Upload(ctx, env.item.OrderID, env.item.ID, file, 7); result <- err }()
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("provider did not enter")
	}
	var pid int
	select {
	case pid = <-pidChannel:
	case <-time.After(time.Second):
		t.Fatal("receipt lock holder missing")
	}
	receipt, payload := loadOrderEvidencePrecommittedUploadReceiptAndPayload(t, env.db)
	require.Equal(t, env.provider.preallocatedKeys, payload.ObjectKeys)
	claimed, err := repository.NewOutboxRepository(env.db).ClaimReadyEvents(time.Now().Add(6*time.Minute), "evidence-visible-receipt", 100, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)
	cancel()
	cleanup := make(chan error, 1)
	handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
	handler.ConfigureOrderEvidenceAttachmentService(env.service)
	go func() { cleanup <- handler.Handle(t.Context(), receipt) }()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, env.db, pid)
	release()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(10 * time.Second):
		t.Fatal("canceled writer stuck")
	}
	select {
	case err := <-cleanup:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("cleanup stuck")
	}
	assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], false)
	var count int64
	require.NoError(t, env.db.Unscoped().Model(&orderevidence.OrderEvidenceAttachment{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestPostgreSQLOrderEvidencePrecommittedWaitingWriterRejectsClaimedCompletedRetriedExpiredMissingOrChangedReceipts(t *testing.T) {
	for _, fence := range []string{"claimed", "completed", "manually retried", "expired", "missing", "changed payload"} {
		t.Run(fence, func(t *testing.T) {
			env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			locked, releaseChannel := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
			defer release()
			blocker := make(chan error, 1)
			go func() {
				blocker <- env.service.evidenceRepo.WithLockedOrderEvidenceAttachmentLifecycleTransaction(ctx, func(context.Context, *repository.OrderEvidenceRepository) error {
					close(locked)
					select {
					case <-releaseChannel:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				})
			}()
			select {
			case <-locked:
			case <-ctx.Done():
				t.Fatal("namespace blocker missing")
			}
			file := newEvidencePNGFileHeader(t, "proof.png")
			result := make(chan error, 1)
			go func() { _, err := env.service.Upload(ctx, env.item.OrderID, env.item.ID, file, 7); result <- err }()
			var receipt outbox.Event
			require.Eventually(t, func() bool {
				return env.db.Where("event_key LIKE ?", repository.OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix+"%").First(&receipt).Error == nil
			}, 5*time.Second, 20*time.Millisecond)
			repo := repository.NewOutboxRepository(env.db)
			switch fence {
			case "claimed", "completed", "manually retried":
				claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "evidence-fenced-receipt", 100, time.Minute)
				require.NoError(t, err)
				require.Len(t, claimed, 1)
				if fence == "completed" {
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "evidence-fenced-receipt", time.Now()))
				}
				if fence == "manually retried" {
					require.NoError(t, env.db.Model(&receipt).Update("status", outbox.EventStatusFailed).Error)
					require.NoError(t, repo.RetryFailedEvent(receipt.ID, "retry cleanup", time.Now()))
				}
			case "expired":
				require.NoError(t, env.db.Model(&receipt).Update("available_at", time.Now().Add(-time.Minute)).Error)
			case "missing":
				require.NoError(t, env.db.Unscoped().Delete(&receipt).Error)
			case "changed payload":
				var payload map[string]interface{}
				require.NoError(t, json.Unmarshal(receipt.Payload, &payload))
				payload["unexpected_mutation"] = true
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				require.NoError(t, env.db.Model(&receipt).Update("payload", encoded).Error)
			}
			release()
			require.NoError(t, <-blocker)
			select {
			case err := <-result:
				require.ErrorIs(t, err, repository.ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable)
			case <-ctx.Done():
				t.Fatal("fenced writer stuck")
			}
			require.Empty(t, env.provider.preallocatedKeys)
			require.Zero(t, env.provider.deleteCalls)
		})
	}
}

func TestPostgreSQLOrderEvidencePrecommittedWaitingUploadRechecksLockedSupersededOrChangedEvidenceBeforeWriting(t *testing.T) {
	for _, mutation := range []string{"locked", "superseded", "item type changed"} {
		t.Run(mutation, func(t *testing.T) {
			env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			locked, releaseChannel := make(chan int, 1), make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(releaseChannel) }) }
			defer release()
			blocker := make(chan error, 1)
			go func() {
				blocker <- env.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
					return env.service.evidenceRepo.WithTx(tx).WithLockedOrderEvidenceAttachmentLifecycleTransaction(ctx, func(context.Context, *repository.OrderEvidenceRepository) error {
						var pid int
						if err := tx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
							return err
						}
						locked <- pid
						select {
						case <-releaseChannel:
						case <-ctx.Done():
							return ctx.Err()
						}
						if mutation == "item type changed" {
							return tx.Exec("UPDATE order_evidence_items SET item_type = ? WHERE id = ?", orderevidence.EvidenceItemTypeSignedPOD, env.item.ID).Error
						}
						return tx.Exec("UPDATE order_evidence_packages SET status = ? WHERE id = ?", mutation, env.pkg.ID).Error
					})
				})
			}()
			var pid int
			select {
			case pid = <-locked:
			case <-ctx.Done():
				t.Fatal("package mutation blocker missing")
			}
			file := newEvidencePNGFileHeader(t, "proof.png")
			result := make(chan error, 1)
			go func() { _, err := env.service.Upload(ctx, env.item.OrderID, env.item.ID, file, 7); result <- err }()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, env.db, pid)
			receipt, payload := loadOrderEvidencePrecommittedUploadReceiptAndPayload(t, env.db)
			require.NotZero(t, receipt.ID)
			release()
			require.NoError(t, <-blocker)
			err := <-result
			switch mutation {
			case "locked":
				require.ErrorIs(t, err, orderevidence.ErrOrderEvidencePackageLocked)
			case "superseded":
				require.ErrorIs(t, err, orderevidence.ErrOrderEvidencePackageSuperseded)
			default:
				require.ErrorIs(t, err, ErrOrderEvidenceAttachmentOwnership)
			}
			require.Empty(t, env.provider.preallocatedKeys)
			assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], false)
			require.Equal(t, 1, processOrderEvidencePrecommittedUploadReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute)).Processed)
		})
	}
}

func TestPostgreSQLOrderEvidencePrecommittedReceiptExchangeFailureRollsBackAttachmentAndPreservesOriginalReceipt(t *testing.T) {
	env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
	failure := errors.New("reject evidence receipt exchange")
	callback := "reject_postgresql_evidence_receipt_exchange_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			tx.AddError(failure)
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
	_, err := env.service.Upload(t.Context(), env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
	require.ErrorIs(t, err, failure)
	var count int64
	require.NoError(t, env.db.Unscoped().Model(&orderevidence.OrderEvidenceAttachment{}).Count(&count).Error)
	require.Zero(t, count)
	receipt, payload := loadOrderEvidencePrecommittedUploadReceiptAndPayload(t, env.db)
	require.Equal(t, outbox.EventStatusPending, receipt.Status)
	require.Zero(t, env.provider.deleteCalls)
	assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], true)
	require.Equal(t, 1, processOrderEvidencePrecommittedUploadReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute)).Processed)
	assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], false)
}
