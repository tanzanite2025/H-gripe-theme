package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
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

func createRetiredRegisteredCustomerServicePrivateOriginalAndDurableCleanupEventForTest(t *testing.T, env customerServicePrivateAttachmentAccessTestEnvironment, key string) outbox.Event {
	t.Helper()
	env.mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
	target := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.mediaService.storage, key)
	require.NoError(t, env.db.Model(target).Update("visibility", "private").Error)
	require.NoError(t, env.mediaService.repo.WithinMediaAssetTransactionWithRequestContext(t.Context(), func(repo *repository.MediaRepository, tx *gorm.DB) error {
		asset, err := repo.FindAssetByIDUnscopedForUpdate(target.ID)
		if err != nil {
			return err
		}
		return env.mediaService.retireUnreferencedMediaAssetWithinTransaction(repo, tx, asset)
	}))
	expected, err := newObjectStorageCleanupEvent(objectCleanupResourceMediaAsset, fmt.Sprint(target.ID), outbox.AggregateTypeMediaAsset, fmt.Sprint(target.ID), []string{key})
	require.NoError(t, err)
	stored, err := repository.NewOutboxRepository(env.db).FindEventByKey(expected.EventKey)
	require.NoError(t, err)
	return *stored
}

func TestPostgreSQLCustomerServiceUnregisteredOriginalCleanupSerializesReceiptAndMediaAdmissionInBothOrders(t *testing.T) {
	for _, claimKind := range []string{"receipt", "whole-asset-receipt", "original", "windows-key-original", "url-only-original", "derivative"} {
		for _, first := range []string{"claim", "cleanup"} {
			t.Run(claimKind+"/"+first+"-first", func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				db, provider := env.db.WithContext(ctx), env.storage.StorageService
				env.mediaService = NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
				key, event := createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t, env)
				if claimKind == "whole-asset-receipt" {
					event = createRetiredRegisteredCustomerServicePrivateOriginalAndDurableCleanupEventForTest(t, env, key)
				}
				receiptAdmission := claimKind == "receipt" || claimKind == "whole-asset-receipt"
				claim := media.MediaAsset{Filename: "concurrent-original.png", URL: provider.GetURL(key) + "?new-claim=1", StorageKey: key, Visibility: "private", Status: "active"}
				if claimKind == "windows-key-original" {
					claim.StorageKey = strings.ReplaceAll(key, "/", "\\")
				} else if claimKind == "url-only-original" {
					claim.StorageKey = ""
				}
				receipt := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
					BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
					ObjectKey: key, Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
				derivative := media.MediaAssetDerivative{MediaAssetID: env.asset.ID, Preset: "concurrent-claim", PresetVersion: 1, URL: claim.URL, StorageKey: key}
				register := func() error {
					switch claimKind {
					case "receipt", "whole-asset-receipt":
						return env.mediaService.repo.CreateCustomerServicePendingUploadSession(ctx, &receipt)
					case "derivative":
						return env.mediaService.repo.CreateAssetDerivatives([]media.MediaAssetDerivative{derivative})
					default:
						return env.mediaService.repo.CreateAsset(&claim)
					}
				}
				acquired, releaseFirst := make(chan int, 1), make(chan struct{})
				var releaseOnce, pauseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
				t.Cleanup(release)
				callbackName := "pause_unregistered_customer_service_original_admission_" + uuid.NewString()
				if first == "claim" {
					table := "media_assets"
					if receiptAdmission {
						table = "media_upload_sessions"
					} else if claimKind == "derivative" {
						table = "media_asset_derivatives"
					}
					require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == table {
							pauseOnce.Do(func() {
								pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, releaseFirst)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Create().Remove(callbackName) })
				} else {
					require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == "media_upload_sessions" || tx.Statement.Table == "uploads" {
							pauseOnce.Do(func() {
								pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, releaseFirst)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Query().Remove(callbackName) })
				}
				handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
				claimResult, cleanupResult := make(chan error, 1), make(chan error, 1)
				if first == "claim" {
					go func() { claimResult <- register() }()
				} else {
					go func() { cleanupResult <- handler.Handle(ctx, event) }()
				}
				pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
				if first == "claim" {
					go func() { cleanupResult <- handler.Handle(ctx, event) }()
				} else {
					go func() { claimResult <- register() }()
				}
				waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
				release()
				claimError := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, claimResult)
				cleanupError := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult)
				if first == "claim" {
					require.NoError(t, claimError)
					require.ErrorIs(t, cleanupError, ErrMediaAssetInUse, "recheck the claim committed while cleanup waited")
					require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
				} else {
					require.NoError(t, cleanupError)
					require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
					if receiptAdmission {
						require.NoError(t, claimError, "a preallocated receipt may commit before its bytes exist")
						_, err := provider.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, bytes.NewReader([]byte("new-receipt-bytes")), key)
						require.NoError(t, err)
						require.ErrorIs(t, handler.Handle(ctx, event), ErrMediaAssetInUse, "the original event's late-write reconciliation must retain bytes protected by the new receipt")
						require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
					} else {
						require.ErrorIs(t, claimError, ErrMediaAssetURLUnavailable, "media registration must revalidate bytes after acquiring the lock")
					}
				}
				var receiptCleanupEvent *outbox.Event
				switch claimKind {
				case "receipt", "whole-asset-receipt":
					ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, receipt.ID)
					current, err := env.mediaService.repo.FindCustomerServiceMediaUploadSession(ctx, receipt.ID)
					require.NoError(t, err)
					retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
					retired, err := retention.retireExpiredUnboundCustomerServiceUploadSession(ctx, *current, time.Now().UTC())
					require.NoError(t, err)
					require.True(t, retired)
					expected, err := newObjectStorageCleanupEvent(objectCleanupResourceCustomerServiceUnregisteredUpload, receipt.ID,
						outbox.AggregateTypeCustomerServiceConversation, fmt.Sprint(env.conversation.ID), []string{key})
					require.NoError(t, err)
					receiptCleanupEvent, err = repository.NewOutboxRepository(db).FindEventByKey(expected.EventKey)
					require.NoError(t, err)
				case "derivative":
					require.NoError(t, db.Unscoped().Where("url = ?", derivative.URL).Delete(&media.MediaAssetDerivative{}).Error)
				default:
					if first == "claim" {
						require.NoError(t, repository.NewMediaRepository(db).HardDeleteAsset(claim.ID))
					}
				}
				require.NoError(t, handler.Handle(ctx, event))
				require.NoError(t, handler.Handle(ctx, event))
				if receiptCleanupEvent != nil {
					require.NoError(t, handler.Handle(ctx, *receiptCleanupEvent))
					require.NoError(t, handler.Handle(ctx, *receiptCleanupEvent))
				}
				require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			})
		}
	}
}

func TestPostgreSQLCustomerServiceUnregisteredOriginalCleanupAndReceiptAdmissionReleaseLocksOnCancellationOrRollback(t *testing.T) {
	for _, failure := range []string{"canceled-receipt", "canceled-cleanup", "rolled-back-receipt"} {
		t.Run(failure, func(t *testing.T) {
			env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			db := env.db.WithContext(ctx)
			key, event := createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t, env)
			receipt := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
				BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
				ObjectKey: key, Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
			acquired, releaseFirst := make(chan int, 1), make(chan struct{})
			var releaseOnce, pauseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
			t.Cleanup(release)
			callbackName := "pause_customer_service_receipt_cancellation_or_rollback_" + uuid.NewString()
			rollback := errors.New("injected customer-service upload receipt rollback")
			if failure == "canceled-receipt" {
				require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_upload_sessions" {
						pauseOnce.Do(func() {
							pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, releaseFirst)
						})
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Query().Remove(callbackName) })
			} else {
				require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_upload_sessions" {
						pauseOnce.Do(func() {
							pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, releaseFirst)
							if failure == "rolled-back-receipt" {
								tx.AddError(rollback)
							}
						})
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Create().Remove(callbackName) })
			}
			handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
			claimResult, cleanupResult := make(chan error, 1), make(chan error, 1)
			waitingContext, cancelWaiting := context.WithCancel(ctx)
			defer cancelWaiting()
			if failure == "canceled-receipt" {
				go func() { cleanupResult <- handler.Handle(ctx, event) }()
			} else {
				go func() { claimResult <- env.mediaService.repo.CreateCustomerServicePendingUploadSession(ctx, &receipt) }()
			}
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
			if failure == "canceled-receipt" {
				go func() {
					claimResult <- env.mediaService.repo.CreateCustomerServicePendingUploadSession(waitingContext, &receipt)
				}()
			} else {
				go func() { cleanupResult <- handler.Handle(waitingContext, event) }()
			}
			waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
			if failure == "canceled-receipt" {
				cancelWaiting()
				require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, claimResult), context.Canceled)
			} else if failure == "canceled-cleanup" {
				cancelWaiting()
				require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult), context.Canceled)
				require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			}
			release()
			if failure == "canceled-cleanup" {
				require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, claimResult))
				require.ErrorIs(t, handler.Handle(ctx, event), ErrMediaAssetInUse)
				require.NoError(t, db.Delete(&receipt).Error)
			} else {
				if failure == "rolled-back-receipt" {
					require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, claimResult), rollback)
				}
				require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult))
			}
			require.NoError(t, handler.Handle(ctx, event))
			require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			var claims int64
			require.NoError(t, db.Model(&media.MediaUploadSession{}).Where("object_key = ?", key).Count(&claims).Error)
			require.Zero(t, claims)
		})
	}
}

func TestPostgreSQLCustomerServiceUploadReceiptAdmissionRejectsOuterStaleSnapshotTransactions(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
		ObjectKey: storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/stale-snapshot.png", Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
	err := env.db.Transaction(func(tx *gorm.DB) error {
		return env.mediaService.repo.WithTx(tx).CreateCustomerServicePendingUploadSession(t.Context(), &session)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	require.ErrorContains(t, err, "requires read committed transaction isolation")
	var count int64
	require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", session.ID).Count(&count).Error)
	require.Zero(t, count)
}

type customerServiceOriginalCleanupStorageThatBlocksDeletionUntilReleased struct {
	storage.StorageService
	entered chan struct{}
	release <-chan struct{}
}

func (s *customerServiceOriginalCleanupStorageThatBlocksDeletionUntilReleased) Delete(ctx context.Context, key string) error {
	close(s.entered)
	<-s.release
	// Model an in-flight provider that cannot stop a physical delete on cancel.
	return s.StorageService.Delete(context.WithoutCancel(ctx), key)
}

func (s *customerServiceOriginalCleanupStorageThatBlocksDeletionUntilReleased) ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx context.Context, assetID uint) ([]string, error) {
	return s.StorageService.(storage.CustomerServiceAttachmentDerivativeNamespaceCleanupLister).ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx, assetID)
}

func TestPostgreSQLCustomerServiceUnregisteredOriginalCleanupKeepsObjectLockUntilCanceledProviderDeleteReturns(t *testing.T) {
	for _, cleanupKind := range []string{"unregistered-private-original", "registered-private-asset", "failed-shared-original"} {
		t.Run(cleanupKind, func(t *testing.T) {
			verifyPostgreSQLMediaCleanupKeepsLocksUntilCanceledProviderDeleteReturns(t, cleanupKind)
		})
	}
}

func verifyPostgreSQLMediaCleanupKeepsLocksUntilCanceledProviderDeleteReturns(t *testing.T, cleanupKind string) {
	t.Helper()
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	key, event := createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t, env)
	if cleanupKind == "registered-private-asset" {
		event = createRetiredRegisteredCustomerServicePrivateOriginalAndDurableCleanupEventForTest(t, env, key)
	} else if cleanupKind == "failed-shared-original" {
		key = "catalog/canceled-provider-delete.png"
		writeMediaCleanupOwnershipConflictTestObjects(t, env.publicRoot, key)
		queued, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaUpload, uuid.NewString(), "test", "test", []string{key})
		require.NoError(t, err)
		require.NoError(t, repository.NewOutboxRepository(env.db).CreateEvent(queued))
		event = *queued
	}
	entered, releaseDelete := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseDelete) }) }
	t.Cleanup(release)
	env.mediaService.storage = &customerServiceOriginalCleanupStorageThatBlocksDeletionUntilReleased{StorageService: env.mediaService.storage, entered: entered, release: releaseDelete}
	cleanupContext, cancelCleanup := context.WithCancel(ctx)
	defer cancelCleanup()
	cleanupResult, claimResult := make(chan error, 1), make(chan error, 1)
	acquired := make(chan int, 1)
	var captureOnce sync.Once
	callbackName := "capture_customer_service_original_cleanup_transaction_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "media_upload_sessions" || tx.Statement.Table == "uploads" {
			captureOnce.Do(func() {
				var pid int
				if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
					tx.AddError(err)
				} else {
					acquired <- pid
				}
			})
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callbackName) })
	go func() {
		cleanupResult <- NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle(cleanupContext, event)
	}()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("cleanup did not enter provider deletion")
	}
	receipt := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
		ObjectKey: key, Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
	if cleanupKind == "failed-shared-original" {
		writer := NewMediaService(repository.NewMediaRepository(env.db.WithContext(ctx)), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
		go func() {
			claimResult <- writer.repo.CreateAsset(&media.MediaAsset{Filename: "after-canceled-delete.png", URL: writer.storage.GetURL(key) + "?new-claim=1", StorageKey: key})
		}()
	} else {
		go func() { claimResult <- env.mediaService.repo.CreateCustomerServicePendingUploadSession(ctx, &receipt) }()
	}
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, pid, "pg_advisory_xact_lock")
	cancelCleanup()
	claimReturnedEarly := false
	select {
	case <-claimResult:
		claimReturnedEarly = true
		t.Error("receipt admission finished while the canceled provider delete still owned the object")
	case <-time.After(200 * time.Millisecond):
	}
	release()
	require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult), context.Canceled)
	if !claimReturnedEarly {
		claimError := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, claimResult)
		if cleanupKind == "failed-shared-original" {
			require.ErrorIs(t, claimError, ErrMediaAssetURLUnavailable)
		} else {
			require.NoError(t, claimError)
		}
	}
	root := env.privateRoot
	if cleanupKind == "failed-shared-original" {
		root = env.publicRoot
	}
	require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(key)))
	// A new receipt can now write; the original event still respects its claim.
	env.mediaService.storage = env.storage.StorageService
	if cleanupKind == "failed-shared-original" {
		require.NoError(t, NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle(ctx, event))
	} else {
		_, err := env.mediaService.storage.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, bytes.NewReader([]byte("new-receipt-owned-bytes")), key)
		require.NoError(t, err)
		require.ErrorIs(t, NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle(ctx, event), ErrMediaAssetInUse)
		require.FileExists(t, filepath.Join(root, filepath.FromSlash(key)))
	}
}
