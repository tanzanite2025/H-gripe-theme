package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t *testing.T, namespace string) dedicatedImmediateOwnerUploadReceiptRegressionEnvironment {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, base.db.AutoMigrate(&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &setting.Setting{}))
	// Users are already migrated by the base environment. Avoid traversing and
	// altering their associations when creating the missing profile owner table.
	require.NoError(t, base.db.Migrator().CreateTable(&user.AgentProfile{}))
	return configureDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, base.db, base.storage.StorageService, namespace)
}

func TestPostgreSQLDedicatedImmediateOwnerUploadReceiptCommitsBeforeIOAndKeepsNamespaceUntilCanceledProviderReturns(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createIsolatedPostgreSQLDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			entered, releaseChannel := make(chan struct{}), make(chan struct{})
			var releaseOnce, captureOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
			defer release()
			pidChannel := make(chan int, 1)
			callback := "capture_preallocated_upload_namespace_holder_" + uuid.NewString()
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
			env.provider.afterWrite = func() { close(entered); <-releaseChannel }
			uploadResult := make(chan error, 1)
			go func() { uploadResult <- env.upload(ctx) }()
			select {
			case <-entered:
			case <-time.After(10 * time.Second):
				t.Fatal("provider did not start after receipt commitment")
			}
			var pid int
			select {
			case pid = <-pidChannel:
			case <-time.After(time.Second):
				t.Fatal("writer namespace holder was not captured")
			}
			// This independent connection sees the committed receipt while its row
			// and namespace remain locked by the ongoing synchronous provider IO.
			var receipt outbox.Event
			require.NoError(t, env.db.Where("event_key LIKE ?", repository.DedicatedImmediateOwnerUploadReceiptEventKeyPrefix+"%").First(&receipt).Error)
			require.Contains(t, string(receipt.Payload), env.provider.writtenKeys[0])
			cancel()
			cleanupResult := make(chan error, 1)
			go func() { cleanupResult <- env.handle(t.Context(), receipt) }()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, env.db, pid)
			select {
			case err := <-cleanupResult:
				t.Fatalf("cleanup escaped canceled writer's namespace lock: %v", err)
			default:
			}
			release()
			select {
			case err := <-uploadResult:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(10 * time.Second):
				t.Fatal("canceled upload did not finish")
			}
			select {
			case err := <-cleanupResult:
				require.NoError(t, err)
			case <-time.After(10 * time.Second):
				t.Fatal("cleanup did not resume after provider returned")
			}
			require.Empty(t, env.currentKey())
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.provider.writtenKeys[0])
			var retained outbox.Event
			require.NoError(t, env.db.First(&retained, receipt.ID).Error)
			require.Equal(t, receipt.EventKey, retained.EventKey)
		})
	}
}

func TestPostgreSQLDedicatedImmediateOwnerCleanupClaimFencesWriterWaitingForNamespace(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		for _, workerCompletesFirst := range []bool{false, true} {
			t.Run(namespace+"/"+map[bool]string{false: "claimed cleanup", true: "completed cleanup"}[workerCompletesFirst], func(t *testing.T) {
				env := createIsolatedPostgreSQLDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				// An independent cleanup transaction holds the namespace before the
				// upload starts, but never locks or claims its new receipt row.
				locked, releaseChannel := make(chan struct{}), make(chan struct{})
				var releaseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
				defer release()
				blockerResult := make(chan error, 1)

				go func() {
					operation := func() error {
						close(locked)
						select {
						case <-releaseChannel:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}
					switch namespace {
					case SiteLogoStoragePrefix:
						blockerResult <- repository.NewSiteLogoRepository(env.db).WithLockedCurrentSiteLogoForObjectCleanup(ctx, func(*sitelogodomain.Asset, *repository.MediaRepository) error { return operation() })
					case SiteFaviconStoragePrefix:
						blockerResult <- repository.NewSiteFaviconRepository(env.db).WithLockedCurrentSiteFaviconForObjectCleanup(ctx, func(*sitefavicondomain.Asset, *repository.MediaRepository) error { return operation() })
					default:
						blockerResult <- repository.NewUserRepository(env.db).WithLockedCustomerServiceAvatarNamespaceForPhysicalCleanup(ctx, func(*repository.MediaRepository) error { return operation() })
					}
				}()
				select {
				case <-locked:
				case <-ctx.Done():
					t.Fatal("namespace blocker did not start")
				}
				uploadResult := make(chan error, 1)
				go func() { uploadResult <- env.upload(ctx) }()
				var receipt outbox.Event
				require.Eventually(t, func() bool {
					return env.db.Where("event_key LIKE ?", repository.DedicatedImmediateOwnerUploadReceiptEventKeyPrefix+"%").First(&receipt).Error == nil
				}, 5*time.Second, 20*time.Millisecond)
				repo := repository.NewOutboxRepository(env.db)
				claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "preallocated-cleanup-worker", 10, time.Minute)
				require.NoError(t, err)
				require.Len(t, claimed, 1)
				if workerCompletesFirst {
					// Deleting a not-yet-created object is harmless. Completion must
					// still fence the writer after this namespace is released.
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "preallocated-cleanup-worker", time.Now()))
				}
				release()
				require.NoError(t, <-blockerResult)
				select {
				case err := <-uploadResult:
					require.ErrorIs(t, err, repository.ErrDedicatedImmediateOwnerUploadReceiptNoLongerWritable)
				case <-ctx.Done():
					t.Fatal("fenced writer did not finish")
				}
				require.Empty(t, env.provider.writtenKeys)
				require.Empty(t, env.currentKey())
				if !workerCompletesFirst {
					require.NoError(t, env.handle(ctx, claimed[0]))
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "preallocated-cleanup-worker", time.Now()))
				}
			})
		}
	}
}

func TestPostgreSQLDedicatedImmediateOwnerUploadReceiptRollbackKeepsOriginalTaskAndSuccessfulExchangeKeepsOldCleanup(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createIsolatedPostgreSQLDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			require.NoError(t, env.upload(t.Context()))
			firstKey := env.currentKey()
			require.NoError(t, env.upload(t.Context()))
			require.NotEqual(t, firstKey, env.currentKey())
			var oldCleanup []outbox.Event
			require.NoError(t, env.db.Find(&oldCleanup).Error)
			require.Len(t, oldCleanup, 1)
			require.Contains(t, string(oldCleanup[0].Payload), firstKey)
			failure := errors.New("provider lost response after write")
			env.provider.writeError = failure
			empty := ""
			env.provider.returnReferenceOverride = &empty
			require.ErrorIs(t, env.upload(t.Context()), failure)
			var all []outbox.Event
			require.NoError(t, env.db.Order("id ASC").Find(&all).Error)
			require.Len(t, all, 2)
			require.Equal(t, oldCleanup[0].EventKey, all[0].EventKey)
			require.Contains(t, string(all[1].Payload), env.provider.writtenKeys[2])
			require.NotEqual(t, env.provider.writtenKeys[2], env.currentKey())
		})
	}
}

func TestPostgreSQLDedicatedImmediateOwnerReceiptExchangeFailureRollsBackOwnerAndPreservesExactCleanupTask(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createIsolatedPostgreSQLDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			require.NoError(t, env.upload(t.Context()))
			previousKey := env.currentKey()
			failure := errors.New("receipt exchange database failure after owner mutation")
			callback := "reject_postgresql_dedicated_receipt_exchange_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" {
					tx.AddError(failure)
				}
			}))
			t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
			require.ErrorIs(t, env.upload(t.Context()), failure)
			require.Equal(t, previousKey, env.currentKey())
			var events []outbox.Event
			require.NoError(t, env.db.Find(&events).Error)
			require.Len(t, events, 1, "old owner cleanup must roll back with the rejected owner mutation")
			require.Contains(t, events[0].EventKey, repository.DedicatedImmediateOwnerUploadReceiptEventKeyPrefix)
			require.Contains(t, string(events[0].Payload), env.provider.writtenKeys[1])
			worker := NewOutboxService(repository.NewOutboxRepository(env.db))
			worker.RegisterHandler(events[0].EventType, env.handle)
			result, err := worker.ProcessPending(t.Context(), time.Now().Add(6*time.Minute), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Processed)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, previousKey)
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.provider.writtenKeys[1])
		})
	}
}
