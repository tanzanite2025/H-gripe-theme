package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createPublicMediaDerivativePostgreSQLReceiptTestEnvironment(t *testing.T) (*gorm.DB, *MediaService, *publicMediaDerivativeReceiptProviderWithInjectedFailures, *media.MediaAsset, []byte) {
	t.Helper()
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	service := NewMediaService(repository.NewMediaRepository(env.db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
	payload := testOpaquePNG(t, 80, 40)
	parent, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "postgres-public-parent.png", "image/png", payload), MediaType: "image", UploaderID: 424242})
	require.NoError(t, err)
	provider := &publicMediaDerivativeReceiptProviderWithInjectedFailures{StorageService: env.storage.StorageService}
	service.storage = provider
	return env.db, service, provider, parent, payload
}

func capturePublicMediaDerivativePostgreSQLReceiptLockTransactionPID(t *testing.T, db *gorm.DB) <-chan int {
	t.Helper()
	captured := make(chan int, 1)
	var once sync.Once
	name := "capture_public_derivative_receipt_writer_pid_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table != "outbox_events" {
			return
		}
		if _, ok := tx.Statement.ConnPool.(gorm.TxCommitter); !ok {
			return
		}
		once.Do(func() {
			var pid int
			err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error
			if err != nil {
				tx.AddError(err)
			} else {
				captured <- pid
			}
		})
	}))
	t.Cleanup(func() { _ = db.Callback().Query().Remove(name) })
	return captured
}

func TestPostgreSQLPublicMediaDerivativeBatchReceiptsAreVisibleSkippedByWorkerAndHoldCleanupUntilProviderReturns(t *testing.T) {
	for _, cancelWriter := range []bool{false, true} {
		t.Run(map[bool]string{false: "successful registration", true: "canceled request"}[cancelWriter], func(t *testing.T) {
			db, service, provider, parent, payload := createPublicMediaDerivativePostgreSQLReceiptTestEnvironment(t)
			pidChannel := capturePublicMediaDerivativePostgreSQLReceiptLockTransactionPID(t, db)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			entered := make(chan string, 1)
			releaseChannel := make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(releaseChannel) }) }
			defer release()
			var first sync.Once
			provider.afterWrite = func(key string) { first.Do(func() { entered <- key; <-releaseChannel }) }
			result := make(chan error, 1)
			go func() {
				_, err := generatePublicMediaDerivativeReceiptTestBatch(ctx, service, parent, payload)
				result <- err
			}()
			var key string
			select {
			case key = <-entered:
			case <-time.After(10 * time.Second):
				t.Fatal("derivative provider did not start")
			}
			var pid int
			select {
			case pid = <-pidChannel:
			case <-time.After(time.Second):
				t.Fatal("writer transaction PID unavailable")
			}
			receipts := findPublicMediaDerivativeReceiptsForTest(t, db)
			require.Len(t, receipts, 3)
			claimed, err := repository.NewOutboxRepository(db).ClaimReadyEvents(time.Now().Add(6*time.Minute), "public-derivative-skip-worker", 100, time.Minute)
			require.NoError(t, err)
			for _, receipt := range receipts {
				for _, event := range claimed {
					require.NotEqual(t, receipt.ID, event.ID)
				}
			}
			var exact outbox.Event
			for _, receipt := range receipts {
				if stringContainsExactDerivativeReceiptKeyForTest(receipt, key) {
					exact = receipt
				}
			}
			require.NotZero(t, exact.ID)
			if cancelWriter {
				cancel()
			}
			cleanup := make(chan error, 1)
			go func() {
				cleanup <- NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil).Handle(t.Context(), exact)
			}()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, t.Context(), db, pid)
			select {
			case err := <-cleanup:
				t.Fatalf("cleanup escaped in-flight derivative writer: %v", err)
			default:
			}
			release()
			select {
			case err := <-result:
				if cancelWriter {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("writer did not finish")
			}
			select {
			case err := <-cleanup:
				if cancelWriter {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, err, ErrMediaAssetInUse)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("cleanup did not resume")
			}
			var count int64
			require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ?", parent.ID).Count(&count).Error)
			if cancelWriter {
				require.Zero(t, count)
				require.Equal(t, receipts, findPublicMediaDerivativeReceiptsForTest(t, db))
				_, err := provider.Open(t.Context(), key)
				require.Error(t, err)
			} else {
				require.EqualValues(t, 3, count)
				require.Empty(t, findPublicMediaDerivativeReceiptsForTest(t, db))
				object, err := provider.Open(t.Context(), key)
				require.NoError(t, err)
				require.NoError(t, object.ReadCloser.Close())
			}
		})
	}
}

func TestPostgreSQLPublicMediaDerivativeClaimCompletionOrParentRetirementFencesWaitingBatchBeforeIO(t *testing.T) {
	for _, scenario := range []string{"claimed", "processed", "parent-retired"} {
		t.Run(scenario, func(t *testing.T) {
			db, service, provider, parent, payload := createPublicMediaDerivativePostgreSQLReceiptTestEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			releaseChannel := make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(releaseChannel) }) }
			defer release()
			locked := make(chan int, 1)
			blockerResult := make(chan error, 1)
			go func() {
				blockerResult <- repository.NewMediaRepository(db).WithinMediaAssetTransactionWithRequestContext(ctx, func(repo *repository.MediaRepository, tx *gorm.DB) error {
					if _, err := repo.FindAssetByIDForUpdate(parent.ID); err != nil {
						return err
					}
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
					if scenario == "parent-retired" {
						return tx.Model(&media.MediaAsset{}).Where("id = ?", parent.ID).Update("status", "deleting").Error
					}
					return nil
				})
			}()
			var pid int
			select {
			case pid = <-locked:
			case <-ctx.Done():
				t.Fatal("parent blocker did not start")
			}
			result := make(chan error, 1)
			go func() {
				_, err := generatePublicMediaDerivativeReceiptTestBatch(ctx, service, parent, payload)
				result <- err
			}()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, db, pid)
			receipts := findPublicMediaDerivativeReceiptsForTest(t, db)
			require.Len(t, receipts, 3)
			if scenario != "parent-retired" {
				events := repository.NewOutboxRepository(db)
				claimed, err := events.ClaimReadyEvents(time.Now().Add(6*time.Minute), "public-derivative-fence-worker", 100, time.Minute)
				require.NoError(t, err)
				found := 0
				for _, receipt := range receipts {
					for _, event := range claimed {
						if event.ID == receipt.ID {
							found++
						}
					}
				}
				require.Equal(t, 3, found)
				if scenario == "processed" {
					for _, receipt := range receipts {
						require.NoError(t, events.MarkProcessedByWorker(receipt.ID, "public-derivative-fence-worker", time.Now()))
					}
				}
			}
			release()
			require.NoError(t, <-blockerResult)
			select {
			case err := <-result:
				require.Error(t, err)
			case <-ctx.Done():
				t.Fatal("fenced batch did not finish")
			}
			require.Empty(t, provider.keys)
			var count int64
			require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ?", parent.ID).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func stringContainsExactDerivativeReceiptKeyForTest(receipt outbox.Event, key string) bool {
	var payload outbox.ObjectStorageCleanupPayload
	if json.Unmarshal(receipt.Payload, &payload) != nil {
		return false
	}
	return len(payload.ObjectKeys) == 1 && payload.ObjectKeys[0] == key
}

func TestPostgreSQLPublicMediaDerivativeReceiptExchangeFailureRollsBackMetadataAndRetryPreservesOldTasks(t *testing.T) {
	db, service, provider, parent, payload := createPublicMediaDerivativePostgreSQLReceiptTestEnvironment(t)
	callback := "reject_second_public_derivative_receipt_exchange_" + uuid.NewString()
	deletes := 0
	injected := errors.New("injected PostgreSQL receipt exchange failure")
	require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			deletes++
			if deletes == 2 {
				tx.AddError(injected)
			}
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Delete().Remove(callback) })
	_, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
	require.ErrorIs(t, err, injected)
	previous := findPublicMediaDerivativeReceiptsForTest(t, db)
	require.Len(t, previous, 3)
	var count int64
	require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ?", parent.ID).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.Callback().Delete().Remove(callback))
	generated, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
	require.NoError(t, err)
	require.Len(t, generated, 3)
	require.Equal(t, previous, findPublicMediaDerivativeReceiptsForTest(t, db))
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	for _, receipt := range previous {
		require.NoError(t, handler.Handle(t.Context(), receipt))
	}
	for _, derivative := range generated {
		object, err := provider.Open(t.Context(), derivative.StorageKey)
		require.NoError(t, err)
		require.NoError(t, object.ReadCloser.Close())
	}
}
