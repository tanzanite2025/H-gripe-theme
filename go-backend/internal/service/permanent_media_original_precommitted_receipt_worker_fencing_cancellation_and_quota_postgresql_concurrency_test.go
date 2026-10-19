package service

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type permanentMediaOriginalPostgreSQLProviderWithWriteBarrier struct {
	storage.StorageService
	afterWrite func(string)
	writeCount atomic.Int32
}

func (s *permanentMediaOriginalPostgreSQLProviderWithWriteBarrier) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.writeCount.Add(1)
	reference, err := s.StorageService.(storage.PreallocatedPermanentMediaOriginalObjectWriter).WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx, reader, key)
	if err == nil && s.afterWrite != nil {
		s.afterWrite(key)
	}
	return reference, err
}
func (s *permanentMediaOriginalPostgreSQLProviderWithWriteBarrier) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func configurePermanentMediaOriginalPostgreSQLUploadTestService(t *testing.T) (*gorm.DB, *MediaService, *permanentMediaOriginalPostgreSQLProviderWithWriteBarrier, MediaUploadInput) {
	t.Helper()
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	provider := &permanentMediaOriginalPostgreSQLProviderWithWriteBarrier{StorageService: env.storage.StorageService}
	service := NewMediaService(repository.NewMediaRepository(env.db), provider, nil, "https://shop.example.test", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
	input := MediaUploadInput{File: multipartFileHeader(t, "permanent.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 424242}, UploadOrigin: media.MediaUploadOriginAdministratorCatalog}
	return env.db, service, provider, input
}

func capturePermanentMediaOriginalPostgreSQLWriterReceiptLockPID(t *testing.T, db *gorm.DB) <-chan int {
	t.Helper()
	captured := make(chan int, 1)
	var once sync.Once
	name := "capture_permanent_media_original_receipt_writer_" + uuid.NewString()
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

func TestPostgreSQLPermanentMediaOriginalReceiptVisibilityWorkerSkipAndCancellationKeepObjectLockUntilProviderReturns(t *testing.T) {
	db, service, provider, input := configurePermanentMediaOriginalPostgreSQLUploadTestService(t)
	pidChannel := capturePermanentMediaOriginalPostgreSQLWriterReceiptLockPID(t, db)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, releaseChannel := make(chan string, 1), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseChannel) }) }
	defer release()
	provider.afterWrite = func(key string) { entered <- key; <-releaseChannel }
	result := make(chan error, 1)
	go func() { _, err := service.UploadAsset(ctx, input); result <- err }()
	var key string
	select {
	case key = <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("writer did not reach provider")
	}
	var pid int
	select {
	case pid = <-pidChannel:
	case <-time.After(time.Second):
		t.Fatal("writer PID missing")
	}
	receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
	require.Contains(t, string(receipt.Payload), key)
	claimed, err := repository.NewOutboxRepository(db).ClaimReadyEvents(time.Now().Add(6*time.Minute), "permanent-original-skip-worker", 100, time.Minute)
	require.NoError(t, err)
	for _, event := range claimed {
		require.NotEqual(t, receipt.ID, event.ID, "active writer receipt must be skipped even when worker clock passes its deadline")
	}
	cancel()
	cleanup := make(chan error, 1)
	go func() {
		cleanup <- NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil).Handle(t.Context(), receipt)
	}()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, db, pid)
	select {
	case err := <-cleanup:
		t.Fatalf("cleanup escaped provider lock: %v", err)
	default:
	}
	release()
	select {
	case err := <-result:
		require.Error(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("canceled writer did not finish")
	}
	select {
	case err := <-cleanup:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("cleanup did not resume")
	}
	_, err = provider.Open(t.Context(), key)
	require.Error(t, err)
	var assets int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Where("uploader_id = ?", input.UploadActorIdentity.AccountID).Count(&assets).Error)
	require.Zero(t, assets)
	require.Equal(t, receipt.ID, permanentMediaOriginalUploadOnlyReceipt(t, service.repo).ID)
}

func TestPostgreSQLPermanentMediaOriginalCleanupClaimFencesWriterWaitingForExactObjectLock(t *testing.T) {
	for _, complete := range []bool{false, true} {
		t.Run(map[bool]string{false: "claimed", true: "processed"}[complete], func(t *testing.T) {
			db, service, provider, input := configurePermanentMediaOriginalPostgreSQLUploadTestService(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			releaseChannel := make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(releaseChannel) }) }
			defer release()
			blockerPID := make(chan int, 1)
			blockerResult := make(chan error, 1)
			name := "hold_exact_object_before_permanent_writer_" + uuid.NewString()
			require.NoError(t, db.Callback().Create().After("gorm:create").Register(name, func(tx *gorm.DB) {
				event, ok := tx.Statement.Dest.(*outbox.Event)
				if !ok || tx.Statement.Table != "outbox_events" {
					return
				}
				var payload outbox.ObjectStorageCleanupPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					tx.AddError(err)
					return
				}
				if len(payload.ObjectKeys) != 1 {
					return
				}
				locked := make(chan struct{})
				go func() {
					blockerResult <- repository.NewMediaRepository(db).WithinMediaAssetTransactionWithRequestContext(ctx, func(repo *repository.MediaRepository, lockTx *gorm.DB) error {
						if err := repo.LockExactMediaObjectBeforePhysicalCleanupWithinTransaction(payload.ObjectKeys[0]); err != nil {
							return err
						}
						var pid int
						if err := lockTx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
							return err
						}
						blockerPID <- pid
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
					tx.AddError(ctx.Err())
				}
			}))
			t.Cleanup(func() { _ = db.Callback().Create().Remove(name) })
			result := make(chan error, 1)
			go func() { _, err := service.UploadAsset(ctx, input); result <- err }()
			var pid int
			select {
			case pid = <-blockerPID:
			case <-ctx.Done():
				t.Fatal("object blocker did not start")
			}
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, db, pid)
			receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
			events := repository.NewOutboxRepository(db)
			claimed, err := events.ClaimReadyEvents(time.Now().Add(6*time.Minute), "permanent-original-fence-worker", 100, time.Minute)
			require.NoError(t, err)
			found := false
			for _, event := range claimed {
				if event.ID == receipt.ID {
					found = true
				}
			}
			require.True(t, found)
			if complete {
				require.NoError(t, events.MarkProcessedByWorker(receipt.ID, "permanent-original-fence-worker", time.Now()))
			}
			release()
			require.NoError(t, <-blockerResult)
			select {
			case err := <-result:
				require.ErrorIs(t, err, repository.ErrPermanentMediaOriginalUploadReceiptNoLongerWritable)
			case <-ctx.Done():
				t.Fatal("writer did not return")
			}
			require.Zero(t, provider.writeCount.Load())
		})
	}
}

func TestPostgreSQLPermanentMediaOriginalConcurrentQuotaRecheckAllowsOneWriteAndOneAtomicAssetExchange(t *testing.T) {
	db, service, provider, input := configurePermanentMediaOriginalPostgreSQLUploadTestService(t)
	service.accountStorageQuotaBytes = input.File.Size
	pidChannel := capturePermanentMediaOriginalPostgreSQLWriterReceiptLockPID(t, db)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	entered, releaseChannel := make(chan struct{}), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseChannel) }) }
	defer release()
	provider.afterWrite = func(string) { close(entered); <-releaseChannel }
	first, second := make(chan error, 1), make(chan error, 1)
	go func() { _, err := service.UploadAsset(ctx, input); first <- err }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("first writer did not start")
	}
	var pid int
	select {
	case pid = <-pidChannel:
	case <-ctx.Done():
		t.Fatal("writer PID missing")
	}
	go func() { _, err := service.UploadAsset(ctx, input); second <- err }()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, db, pid)
	var receipts int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_key LIKE ?", repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix+"%").Count(&receipts).Error)
	require.EqualValues(t, 2, receipts)
	release()
	select {
	case err := <-first:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("first writer did not commit")
	}
	select {
	case err := <-second:
		require.ErrorIs(t, err, ErrMediaAccountStorageQuotaExceeded)
	case <-ctx.Done():
		t.Fatal("second writer did not reject quota")
	}
	require.EqualValues(t, 1, provider.writeCount.Load())
	var assets int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Where("uploader_id = ?", input.UploadActorIdentity.AccountID).Count(&assets).Error)
	require.EqualValues(t, 1, assets)
	receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
	require.Equal(t, outbox.EventStatusPending, receipt.Status)
	require.NoError(t, NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil).Handle(t.Context(), receipt))
}
