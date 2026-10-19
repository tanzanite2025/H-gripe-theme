package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLVisualShowcaseDraftUploadReceiptEnvironment(t *testing.T) (*gorm.DB, *HomeVisualTileService, *visualShowcaseDraftUploadControlledStorage) {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, base.db.AutoMigrate(&homevisualtile.Tile{}))
	svc, provider := configureVisualShowcaseDraftUploadControlledStorage(t, base.db, base.storage.StorageService)
	return base.db, svc, provider
}

func TestPostgreSQLVisualShowcaseDraftUploadCommitsReceiptBeforeIOAndKeepsLocksUntilCanceledProviderReturns(t *testing.T) {
	db, svc, provider := createIsolatedPostgreSQLVisualShowcaseDraftUploadReceiptEnvironment(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, releaseChannel := make(chan struct{}), make(chan struct{})
	var releaseOnce, captureOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
	defer release()
	pidChannel := make(chan int, 1)
	callback := "capture_visual_showcase_draft_writer_backend_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
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
	t.Cleanup(func() { _ = db.Callback().Query().Remove(callback) })
	provider.afterWrite = func() { close(entered); <-releaseChannel }
	file := multipartFileHeader(t, "showcase.png", "image/png", testOpaquePNG(t, 600, 600))
	uploadResult := make(chan error, 1)
	go func() { _, err := svc.UploadAdminImage(ctx, "home-hero", "en", file); uploadResult <- err }()
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("provider did not start")
	}
	var pid int
	select {
	case pid = <-pidChannel:
	case <-time.After(time.Second):
		t.Fatal("writer backend missing")
	}
	// A different connection sees the independently committed receipt during IO.
	receipt, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, provider.writtenKeys[0])
	require.Equal(t, "uploading", payload.UploadState)
	repo := repository.NewOutboxRepository(db)
	claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "visual-writer-race", 100, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed, "worker skips the receipt row locked by ongoing IO")
	cancel()
	cleanupResult := make(chan error, 1)
	go func() {
		cleanupResult <- NewObjectStorageCleanupOutboxHandler(nil, nil, svc, nil).Handle(t.Context(), receipt)
	}()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, t.Context(), db, pid)
	release()
	select {
	case err := <-uploadResult:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(10 * time.Second):
		t.Fatal("canceled upload stuck")
	}
	select {
	case err := <-cleanupResult:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("cleanup stuck")
	}
	retained, retainedPayload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, provider.writtenKeys[0])
	require.Equal(t, receipt.ID, retained.ID)
	require.Equal(t, "uploading", retainedPayload.UploadState)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, provider.writtenKeys[0], false)
}

func TestPostgreSQLVisualShowcaseDraftWaitingWriterIsFencedByCleanupClaimCompletionRetryExpiryMissingOrChangedReceipt(t *testing.T) {
	for _, fence := range []string{"claimed", "completed", "manually retried", "expired", "missing", "changed payload", "extra payload field"} {
		t.Run(fence, func(t *testing.T) {
			db, svc, provider := createIsolatedPostgreSQLVisualShowcaseDraftUploadReceiptEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			locked, releaseChannel := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
			defer release()
			blockerResult := make(chan error, 1)
			go func() {
				blockerResult <- repository.NewHomeVisualTileRepository(db).WithLockedVisualShowcaseReferencesForObjectCleanup(ctx, func(*repository.HomeVisualTileRepository, *repository.MediaRepository) error {
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
				t.Fatal("namespace blocker did not start")
			}
			file := multipartFileHeader(t, "showcase.png", "image/png", testOpaquePNG(t, 600, 600))
			uploadResult := make(chan error, 1)
			go func() { _, err := svc.UploadAdminImage(ctx, "home-hero", "en", file); uploadResult <- err }()
			var receipt outbox.Event
			require.Eventually(t, func() bool {
				return db.Where("event_key LIKE ?", homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix+"%").First(&receipt).Error == nil
			}, 5*time.Second, 20*time.Millisecond)
			repo := repository.NewOutboxRepository(db)
			switch fence {
			case "claimed", "completed", "manually retried":
				claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "visual-draft-fence", 100, time.Minute)
				require.NoError(t, err)
				require.Len(t, claimed, 1)
				if fence == "completed" {
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "visual-draft-fence", time.Now()))
				}
				if fence == "manually retried" {
					require.NoError(t, db.Model(&receipt).Update("status", outbox.EventStatusFailed).Error)
					require.NoError(t, repo.RetryFailedEvent(receipt.ID, "retry visual draft cleanup", time.Now()))
				}
			case "expired":
				require.NoError(t, db.Model(&receipt).Update("available_at", time.Now().Add(-time.Minute)).Error)
			case "missing":
				require.NoError(t, db.Unscoped().Delete(&receipt).Error)
			case "changed payload":
				var payload homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload
				require.NoError(t, json.Unmarshal(receipt.Payload, &payload))
				payload.RequestedAt = payload.RequestedAt.Add(time.Second)
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				require.NoError(t, db.Model(&receipt).Update("payload", encoded).Error)
			case "extra payload field":
				var document map[string]interface{}
				require.NoError(t, json.Unmarshal(receipt.Payload, &document))
				document["unexpected_receipt_mutation"] = true
				encoded, err := json.Marshal(document)
				require.NoError(t, err)
				require.NoError(t, db.Model(&receipt).Update("payload", encoded).Error)
			}
			release()
			require.NoError(t, <-blockerResult)
			select {
			case err := <-uploadResult:
				require.ErrorIs(t, err, repository.ErrVisualShowcaseDraftUploadReceiptUnavailable)
			case <-ctx.Done():
				t.Fatal("fenced writer stuck")
			}
			require.Empty(t, provider.writtenKeys)
		})
	}
}

func TestPostgreSQLVisualShowcaseDraftSaveAndExpiryWorkerSerializeClaimAndOwnerAdmission(t *testing.T) {
	for _, firstOperation := range []string{"save", "worker"} {
		for _, saveMode := range []string{"single card", "whole group"} {
			t.Run(firstOperation+" first/"+saveMode, func(t *testing.T) {
				db, svc, _ := createIsolatedPostgreSQLVisualShowcaseDraftUploadReceiptEnvironment(t)
				input := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
				receipt, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, input.StorageKey)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				locked, releaseChannel := make(chan int, 1), make(chan struct{})
				var releaseOnce, pauseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
				defer release()
				callback := "pause_visual_draft_save_expiry_race_" + uuid.NewString()
				if firstOperation == "save" {
					require.NoError(t, db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "visual_showcase_items" {
							pauseOnce.Do(func() {
								pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Create().Remove(callback) })
				} else {
					require.NoError(t, db.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "visual_showcase_items" {
							pauseOnce.Do(func() {
								pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Query().Remove(callback) })
				}
				saveResult, workerResult := make(chan error, 1), make(chan error, 1)
				save := func() {
					var err error
					if saveMode == "single card" {
						_, err = svc.SaveSingleVisualShowcaseAdminItem(ctx, "home-hero", "en", input)
					} else {
						_, err = svc.ReplaceAdminItems(ctx, "home-hero", "en", []HomeVisualTileInput{input})
					}
					saveResult <- err
				}
				worker := NewOutboxService(repository.NewOutboxRepository(db))
				worker.RegisterHandler(receipt.EventType, NewObjectStorageCleanupOutboxHandler(nil, nil, svc, nil).Handle)
				cleanup := func() {
					result, err := worker.ProcessPending(ctx, payload.DraftExpiresAt.Add(time.Second), 100)
					if err == nil && result.Processed != 1 {
						err = errors.New("worker did not complete exactly the draft cleanup")
					}
					workerResult <- err
				}
				if firstOperation == "save" {
					go save()
				} else {
					go cleanup()
				}
				var pid int
				select {
				case pid = <-locked:
				case <-ctx.Done():
					t.Fatal("first operation did not reach lock")
				}
				if firstOperation == "save" {
					claimed, err := repository.NewOutboxRepository(db).ClaimReadyEvents(payload.DraftExpiresAt.Add(time.Second), "visual-save-row-lock-race", 100, time.Minute)
					require.NoError(t, err)
					require.Empty(t, claimed)
					release()
					require.NoError(t, <-saveResult)
					requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, input.StorageKey, true)
					result, err := worker.ProcessPending(ctx, payload.DraftExpiresAt.Add(time.Second), 100)
					require.NoError(t, err)
					require.Zero(t, result.Processed)
				} else {
					go save()
					requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, db, pid)
					release()
					require.NoError(t, <-workerResult)
					require.ErrorIs(t, <-saveResult, ErrHomeVisualTileImageInvalid)
					requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, input.StorageKey, false)
				}
			})
		}
	}
}

func TestPostgreSQLVisualShowcaseDraftOwnerTransactionRollbackRestoresOriginalAndThumbnailReceipts(t *testing.T) {
	for _, saveMode := range []string{"single card", "whole group"} {
		t.Run(saveMode, func(t *testing.T) {
			db, svc, _ := createIsolatedPostgreSQLVisualShowcaseDraftUploadReceiptEnvironment(t)
			original := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
			thumbnail := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 2)
			original.ThumbnailURL = thumbnail.ImageURL
			callback := "reject_visual_draft_postgresql_owner_after_create_" + uuid.NewString()
			failure := errors.New("owner persistence rejected after receipt exchange")
			require.NoError(t, db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == "visual_showcase_items" {
					tx.AddError(failure)
				}
			}))
			defer db.Callback().Create().Remove(callback)
			var err error
			if saveMode == "single card" {
				_, err = svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", original)
			} else {
				_, err = svc.ReplaceAdminItems(t.Context(), "home-hero", "en", []HomeVisualTileInput{original})
			}
			require.ErrorIs(t, err, failure)
			for _, key := range []string{original.StorageKey, thumbnail.StorageKey} {
				event, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, key)
				require.Equal(t, outbox.EventStatusPending, event.Status)
				require.Equal(t, "ready", payload.UploadState)
				requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, key, true)
			}
			items, err := svc.GetAdminItems("home-hero", "en")
			require.NoError(t, err)
			require.Empty(t, items)
		})
	}
}

var _ storage.PreallocatedVisualShowcaseDraftObjectWriter = (*visualShowcaseDraftUploadControlledStorage)(nil)
