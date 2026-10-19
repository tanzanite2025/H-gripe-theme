package service

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLUGCShowcasePrecommittedWriteEnvironment(t *testing.T) ugcShowcaseDurableCompensationRegressionEnvironment {
	t.Helper()
	env := createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&order.Order{}))
	configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
	return env
}

func prepareUGCShowcasePrecommittedOperationForConcurrencyTest(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment, operation string) (func(context.Context) error, *ugcshowcase.UGCShowcase) {
	t.Helper()
	if operation == "upload" {
		first := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "first.webp")
		second := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "second.webp")
		return func(ctx context.Context) error {
			_, err := env.service.UploadPhotos(ctx, 7, 91, []*multipart.FileHeader{first, second}, nil)
			return err
		}, nil
	}
	keys := []string{env.key, "showcase/pending/historical/second.webp"}
	for _, key := range keys {
		writeUGCShowcaseRegressionObject(t, env, key)
	}
	owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, keys...)
	return func(ctx context.Context) error { return env.service.Approve(ctx, owner.ID) }, owner
}

func TestPostgreSQLUGCShowcasePrecommittedReceiptIsVisibleBeforeIOAndCanceledProviderRetainsNamespaceAndReceiptLocks(t *testing.T) {
	for _, operation := range []string{"upload", "publish"} {
		t.Run(operation, func(t *testing.T) {
			env := createIsolatedPostgreSQLUGCShowcasePrecommittedWriteEnvironment(t)
			run, owner := prepareUGCShowcasePrecommittedOperationForConcurrencyTest(t, env, operation)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			entered, releaseChannel := make(chan struct{}), make(chan struct{})
			var releaseOnce, captureOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
			defer release()
			pidChannel := make(chan int, 1)
			callback := "capture_ugc_precommitted_writer_backend_" + uuid.NewString()
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
			pause := func() { close(entered); <-releaseChannel }
			if operation == "upload" {
				env.provider.afterUpload = pause
			} else {
				env.provider.beforeCopy = pause
			}
			result := make(chan error, 1)
			go func() { result <- run(ctx) }()
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
			receipt, payload := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			require.Len(t, payload.ObjectKeys, 2)
			claimed, err := repository.NewOutboxRepository(env.db).ClaimReadyEvents(time.Now().Add(6*time.Minute), "ugc-precommitted-visible", 100, time.Minute)
			require.NoError(t, err)
			require.Empty(t, claimed)
			cancel()
			cleanup := make(chan error, 1)
			go func() {
				cleanup <- NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(t.Context(), receipt)
			}()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, t.Context(), env.db, pid)
			release()
			select {
			case err := <-result:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(10 * time.Second):
				t.Fatal("canceled operation stuck")
			}
			select {
			case err := <-cleanup:
				require.NoError(t, err)
			case <-time.After(10 * time.Second):
				t.Fatal("cleanup stuck after provider returned")
			}
			for _, key := range payload.ObjectKeys {
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
			if owner != nil {
				current, err := env.service.Get(owner.ID)
				require.NoError(t, err)
				require.Equal(t, ugcshowcase.StatusPending, current.Status)
				requireUGCShowcaseRegressionObjectExists(t, env, env.key, true)
			}
		})
	}
}

func TestPostgreSQLUGCShowcasePrecommittedWaitingWriterRejectsClaimedCompletedRetriedExpiredMissingOrChangedReceipts(t *testing.T) {
	for _, operation := range []string{"upload", "publish"} {
		for _, fence := range []string{"claimed", "completed", "manually retried", "expired", "missing", "changed payload"} {
			t.Run(operation+"/"+fence, func(t *testing.T) {
				env := createIsolatedPostgreSQLUGCShowcasePrecommittedWriteEnvironment(t)
				run, _ := prepareUGCShowcasePrecommittedOperationForConcurrencyTest(t, env, operation)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				locked, releaseChannel := make(chan struct{}), make(chan struct{})
				var once sync.Once
				release := func() { once.Do(func() { close(releaseChannel) }) }
				defer release()
				blocker := make(chan error, 1)
				go func() {
					blocker <- env.service.repo.WithLockedUGCShowcaseLifecycleUntilProviderOperationReturns(ctx, func(*repository.UGCShowcaseRepository, *gorm.DB) error {
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
				result := make(chan error, 1)
				go func() { result <- run(ctx) }()
				var receipt outbox.Event
				require.Eventually(t, func() bool {
					return env.db.Where("event_key LIKE ?", repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix+"%").First(&receipt).Error == nil
				}, 5*time.Second, 20*time.Millisecond)
				repo := repository.NewOutboxRepository(env.db)
				switch fence {
				case "claimed", "completed", "manually retried":
					claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "ugc-precommitted-fence", 100, time.Minute)
					require.NoError(t, err)
					require.Len(t, claimed, 1)
					if fence == "completed" {
						require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "ugc-precommitted-fence", time.Now()))
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
					require.ErrorIs(t, err, repository.ErrUGCShowcaseObjectWriteReceiptNoLongerWritable)
				case <-ctx.Done():
					t.Fatal("fenced writer stuck")
				}
				require.Empty(t, env.provider.preallocatedPendingKeys)
				require.Empty(t, env.provider.copiedKeys)
			})
		}
	}
}

func TestPostgreSQLUGCShowcasePrecommittedPublicationRejectsChangedOwnerImagesOrStatusBeforeCopying(t *testing.T) {
	for _, mutation := range []string{"images", "status"} {
		t.Run(mutation, func(t *testing.T) {
			env := createIsolatedPostgreSQLUGCShowcasePrecommittedWriteEnvironment(t)
			run, owner := prepareUGCShowcasePrecommittedOperationForConcurrencyTest(t, env, "publish")
			other := "showcase/pending/historical/replaced.webp"
			writeUGCShowcaseRegressionObject(t, env, other)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			locked, releaseChannel := make(chan struct{}), make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(releaseChannel) }) }
			defer release()
			blocker := make(chan error, 1)
			go func() {
				blocker <- env.service.repo.WithLockedUGCShowcaseLifecycleUntilProviderOperationReturns(ctx, func(repo *repository.UGCShowcaseRepository, _ *gorm.DB) error {
					close(locked)
					select {
					case <-releaseChannel:
					case <-ctx.Done():
						return ctx.Err()
					}
					if mutation == "status" {
						_, err := repo.UpdatePendingStatus(owner.ID, ugcshowcase.StatusRejected, "rejected during publication planning")
						return err
					}
					encoded, _ := json.Marshal([]string{other})
					_, err := repo.UpdateImagesByStatus(owner.ID, ugcshowcase.StatusPending, encoded)
					return err
				})
			}()
			select {
			case <-locked:
			case <-ctx.Done():
				t.Fatal("owner blocker missing")
			}
			result := make(chan error, 1)
			go func() { result <- run(ctx) }()
			require.Eventually(t, func() bool {
				var count int64
				return env.db.Model(&outbox.Event{}).Where("event_key LIKE ?", repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix+"%").Count(&count).Error == nil && count == 1
			}, 5*time.Second, 20*time.Millisecond)
			release()
			require.NoError(t, <-blocker)
			require.ErrorIs(t, <-result, ErrShowcaseInvalidTransition)
			require.Empty(t, env.provider.copiedKeys)
			_, payload := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			for _, key := range payload.ObjectKeys {
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
		})
	}
}

func TestPostgreSQLUGCShowcasePrecommittedReceiptExchangeFailureRollsBackOwnerAndSourceCleanupTogether(t *testing.T) {
	for _, operation := range []string{"upload", "publish"} {
		t.Run(operation, func(t *testing.T) {
			env := createIsolatedPostgreSQLUGCShowcasePrecommittedWriteEnvironment(t)
			run, owner := prepareUGCShowcasePrecommittedOperationForConcurrencyTest(t, env, operation)
			failure := errors.New("reject final receipt exchange")
			callback := "reject_postgresql_ugc_precommitted_receipt_exchange_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" {
					tx.AddError(failure)
				}
			}))
			t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
			require.ErrorIs(t, run(t.Context()), failure)
			var events []outbox.Event
			require.NoError(t, env.db.Find(&events).Error)
			require.Len(t, events, 1)
			_, payload := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			require.Len(t, payload.ObjectKeys, 2)
			if owner == nil {
				var count int64
				require.NoError(t, env.db.Model(&ugcshowcase.UGCShowcase{}).Count(&count).Error)
				require.Zero(t, count)
			} else {
				current, err := env.service.Get(owner.ID)
				require.NoError(t, err)
				require.Equal(t, ugcshowcase.StatusPending, current.Status)
				require.JSONEq(t, string(owner.Images), string(current.Images))
			}
			result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute))
			require.Equal(t, 1, result.Processed)
		})
	}
}
