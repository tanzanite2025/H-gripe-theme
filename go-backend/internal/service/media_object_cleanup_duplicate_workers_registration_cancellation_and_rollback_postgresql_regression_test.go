package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mediaObjectPhysicalCleanupStorageWithAtomicDeletionCount struct {
	storage.StorageService
	deleteCalls atomic.Int32
}

func (s *mediaObjectPhysicalCleanupStorageWithAtomicDeletionCount) Delete(ctx context.Context, key string) error {
	s.deleteCalls.Add(1)
	return s.StorageService.Delete(ctx, key)
}

func createRetiredSyntheticMediaObjectAndOriginalDurableCleanupEventForObjectOwnershipTest(t *testing.T, db *gorm.DB, assets *MediaService, root, key string) (*media.MediaAsset, outbox.Event) {
	t.Helper()
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, key)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
	require.NoError(t, assets.repo.WithTransaction(func(repo *repository.MediaRepository, tx *gorm.DB) error {
		current, err := repo.FindAssetByIDForUpdate(asset.ID)
		if err != nil {
			return err
		}
		return assets.retireUnreferencedMediaAssetWithinTransaction(repo, tx, current)
	}))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).First(&event).Error)
	return asset, event
}

func TestPostgreSQLMediaObjectCleanupDuplicateWorkersCanceledRegistrationAndSameFilenameNeighborRemainIndependent(t *testing.T) {
	for _, competingOperation := range []string{"duplicate-cleanup-worker", "canceled-registration", "same-filename-neighbor"} {
		t.Run(competingOperation, func(t *testing.T) {
			env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			db, provider := env.db.WithContext(ctx), env.storage.StorageService
			assets := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
			assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			asset, event := createRetiredSyntheticMediaObjectAndOriginalDurableCleanupEventForObjectOwnershipTest(t, db, assets, env.publicRoot, "catalog/first/image.png")
			counter := &mediaObjectPhysicalCleanupStorageWithAtomicDeletionCount{StorageService: provider}
			assets.storage = counter
			acquired, releaseFirst := make(chan int, 1), make(chan struct{})
			var releaseOnce, pauseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
			t.Cleanup(release)
			callbackName := "pause_media_cleanup_before_hard_delete_" + uuid.NewString()
			require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
				if tx.Statement.Table == "media_assets" && tx.Statement.Unscoped {
					pauseOnce.Do(func() {
						pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, releaseFirst)
					})
				}
			}))
			t.Cleanup(func() { _ = db.Callback().Delete().Remove(callbackName) })
			cleanupResult := make(chan error, 1)
			go func() { cleanupResult <- assets.cleanupDeletedAsset(ctx, asset.ID) }()
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
			secondResult := make(chan error, 1)
			switch competingOperation {
			case "duplicate-cleanup-worker":
				go func() { secondResult <- NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil).Handle(ctx, event) }()
				waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "media_assets")
			case "canceled-registration":
				writerContext, cancelWriter := context.WithCancel(ctx)
				defer cancelWriter()
				writer := NewMediaService(repository.NewMediaRepository(db.WithContext(writerContext)), provider, nil, "", 20<<30)
				go func() {
					secondResult <- writer.repo.CreateAsset(&media.MediaAsset{Filename: "canceled.png", URL: provider.GetURL(asset.StorageKey) + "?canceled=1", StorageKey: asset.StorageKey})
				}()
				waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
				cancelWriter()
				require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, secondResult), context.Canceled)
			case "same-filename-neighbor":
				key := "catalog/second/image.png"
				writeMediaCleanupOwnershipConflictTestObjects(t, env.publicRoot, key)
				neighbor := &media.MediaAsset{Filename: "image.png", URL: provider.GetURL(key), StorageKey: key}
				writer := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
				go func() { secondResult <- writer.repo.CreateAsset(neighbor) }()
				require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, secondResult), "unrelated paths must complete while cleanup still holds its locks")
				require.NotZero(t, neighbor.ID)
				require.FileExists(t, filepath.Join(env.publicRoot, filepath.FromSlash(key)))
			}
			release()
			require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult))
			if competingOperation == "duplicate-cleanup-worker" {
				require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, secondResult))
			}
			expectedDeleteAttempts := 1
			if competingOperation == "duplicate-cleanup-worker" {
				// After reloading the removed registration, the durable receipt
				// retries its exact key to collect any late arriving object bytes.
				expectedDeleteAttempts++
			}
			require.EqualValues(t, expectedDeleteAttempts, counter.deleteCalls.Load(), "the durable receipt must preserve exact-key cleanup after registration removal")
			require.NoFileExists(t, filepath.Join(env.publicRoot, filepath.FromSlash(asset.StorageKey)))
		})
	}
}

func TestPostgreSQLMediaObjectCleanupWaitsForRolledBackRegistrationThenRetriesTheOriginalEvent(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	db, provider := env.db.WithContext(ctx), env.storage.StorageService
	assets := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	asset, event := createRetiredSyntheticMediaObjectAndOriginalDurableCleanupEventForObjectOwnershipTest(t, db, assets, env.publicRoot, "catalog/rollback-registration.png")
	acquired, releaseFirst := make(chan int, 1), make(chan struct{})
	var releaseOnce, pauseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	t.Cleanup(release)
	registrationFailure := errors.New("synthetic registration transaction rollback")
	callbackName := "pause_and_rollback_media_object_registration_" + uuid.NewString()
	require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "media_assets" {
			pauseOnce.Do(func() {
				pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, releaseFirst)
				tx.AddError(registrationFailure)
			})
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Create().Remove(callbackName) })
	registerResult, cleanupResult := make(chan error, 1), make(chan error, 1)
	candidate := &media.MediaAsset{Filename: "rollback.png", URL: provider.GetURL(asset.StorageKey) + "?rollback=1", StorageKey: asset.StorageKey}
	go func() { registerResult <- assets.repo.CreateAsset(candidate) }()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
	go func() {
		cleanupResult <- NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil).Handle(ctx, event)
	}()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
	release()
	require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, registerResult), registrationFailure)
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult))
	var count int64
	require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("url = ?", candidate.URL).Count(&count).Error)
	require.Zero(t, count)
	require.NoFileExists(t, filepath.Join(env.publicRoot, filepath.FromSlash(asset.StorageKey)))
}

func TestMediaObjectRegistrationRejectsMissingBytesAfterCleanupDatabaseRollbackAndOriginalEventStillRetries(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	root := t.TempDir()
	assets := newMediaDeleteTestService(t, db, root)
	asset, event := createRetiredSyntheticMediaObjectAndOriginalDurableCleanupEventForObjectOwnershipTest(t, db, assets, root, "catalog/cleanup-database-rollback.png")
	cleanupFailure := errors.New("synthetic cleanup database failure after object deletion")
	callbackName := "fail_media_object_hard_delete_" + uuid.NewString()
	var injectedFailure atomic.Bool
	require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "media_assets" && tx.Statement.Unscoped && injectedFailure.CompareAndSwap(false, true) {
			tx.AddError(cleanupFailure)
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Delete().Remove(callbackName) })
	handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), cleanupFailure)
	require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(asset.StorageKey)))
	retained, err := assets.repo.FindAssetByIDUnscoped(asset.ID)
	require.NoError(t, err)
	require.True(t, retained.DeletedAt.Valid)
	err = assets.repo.CreateAsset(&media.MediaAsset{Filename: "late.png", URL: assets.storage.GetURL(asset.StorageKey) + "?late=1", StorageKey: asset.StorageKey})
	require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
}
