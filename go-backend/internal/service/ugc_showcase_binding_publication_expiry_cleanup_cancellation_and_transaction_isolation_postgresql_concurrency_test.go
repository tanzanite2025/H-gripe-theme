package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/ugcshowcase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t *testing.T) ugcShowcaseDurableCompensationRegressionEnvironment {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, base.db.Migrator().CreateTable(&ugcshowcase.UGCShowcase{}))
	return configureUGCShowcaseDurableCompensationRegressionEnvironment(t, base.db, base.storage.StorageService, base.publicRoot, base.privateRoot)
}

func captureUGCShowcaseNamespaceLockHolderPostgreSQLBackendPID(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment) <-chan int {
	t.Helper()
	acquired := make(chan int, 1)
	var once sync.Once
	callback := "capture_ugc_namespace_holder_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "showcases" {
			if _, active := tx.Statement.ConnPool.(gorm.TxCommitter); !active {
				return
			}
			once.Do(func() {
				var pid int
				if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
					tx.AddError(err)
				} else {
					acquired <- pid
				}
			})
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callback) })
	return acquired
}

func TestPostgreSQLUGCShowcasePendingAndApprovedBindingSerializeWithCleanupInBothOrdersAndRollback(t *testing.T) {
	for _, namespace := range []string{showcasePendingUploadPrefix, showcaseApprovedUploadPrefix} {
		for _, first := range []string{"binding", "cleanup", "rolled back binding"} {
			t.Run(namespace+"/"+first, func(t *testing.T) {
				env := createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t)
				key := namespace + "/concurrent/ride.webp"
				writeUGCShowcaseRegressionObject(t, env, key)
				encoded, err := json.Marshal([]string{key})
				require.NoError(t, err)
				candidate := &ugcshowcase.UGCShowcase{UserID: 7, Status: ugcshowcase.StatusPending, Images: encoded}
				if namespace == showcaseApprovedUploadPrefix {
					candidate = createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, "showcase/pending/old-source.webp")
				}
				event := createUGCShowcaseRegressionCleanupEvent(t, env, 0, key, true)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				locked, releaseFirst := make(chan int, 1), make(chan struct{})
				var once sync.Once
				release := func() { once.Do(func() { close(releaseFirst) }) }
				t.Cleanup(release)
				rollback := errors.New("UGC binding deliberately rolled back")
				if first != "cleanup" {
					callback := "pause_ugc_binding_" + uuid.NewString()
					pause := func(tx *gorm.DB) {
						if tx.Statement.Table == "showcases" {
							pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseFirst)
							if first == "rolled back binding" {
								tx.AddError(rollback)
							}
						}
					}
					if namespace == showcasePendingUploadPrefix {
						require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, pause))
						t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
					} else {
						require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callback, pause))
						t.Cleanup(func() { _ = env.db.Callback().Update().Remove(callback) })
					}
				}
				bindingResult, cleanupResult := make(chan error, 1), make(chan error, 1)
				bind := func() {
					var err error
					if namespace == showcasePendingUploadPrefix {
						err = env.service.repo.WithContext(ctx).Create(candidate)
					} else {
						_, err = env.service.repo.WithContext(ctx).UpdatePendingImagesAndStatus(candidate.ID, encoded, ugcshowcase.StatusApproved, "")
					}
					bindingResult <- err
				}
				cleanup := func() {
					cleanupResult <- NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(ctx, event)
				}
				var pid int
				if first == "cleanup" {
					captured := captureUGCShowcaseNamespaceLockHolderPostgreSQLBackendPID(t, env)
					env.provider.beforeDelete = func() {
						select {
						case <-releaseFirst:
						case <-ctx.Done():
						}
					}
					go cleanup()
					pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, captured)
					go bind()
				} else {
					go bind()
					pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, locked)
					go cleanup()
				}
				requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
				release()
				bindingErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, bindingResult)
				cleanupErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult)
				if first == "binding" {
					require.NoError(t, bindingErr)
					require.ErrorIs(t, cleanupErr, ErrMediaAssetInUse)
					requireUGCShowcaseRegressionObjectExists(t, env, key, true)
				} else {
					require.NoError(t, cleanupErr)
					if first == "cleanup" {
						require.ErrorIs(t, bindingErr, ErrMediaAssetURLUnavailable)
					} else {
						require.ErrorIs(t, bindingErr, rollback)
					}
					requireUGCShowcaseRegressionObjectExists(t, env, key, false)
				}
			})
		}
	}
}

func TestPostgreSQLUGCShowcasePublicationAndRejectionOrExpirySerializeAndRecheckCurrentStatus(t *testing.T) {
	for _, competitor := range []string{"reject", "expire"} {
		for _, first := range []string{"publish", "reject"} {
			t.Run(competitor+"/"+first+" first", func(t *testing.T) {
				env := createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t)
				writeUGCShowcaseRegressionObject(t, env, env.key)
				owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				releaseFirst := make(chan struct{})
				var once sync.Once
				release := func() { once.Do(func() { close(releaseFirst) }) }
				t.Cleanup(release)
				publicationResult, competitorResult := make(chan error, 1), make(chan error, 1)
				publish := func() { publicationResult <- env.service.Approve(ctx, owner.ID) }
				compete := func() {
					if competitor == "reject" {
						competitorResult <- env.service.Reject(ctx, owner.ID, "reject")
					} else {
						_, err := env.service.CleanupExpiredPendingImages(ctx, time.Now(), 24*time.Hour, 10)
						competitorResult <- err
					}
				}
				var pid int
				if first == "publish" {
					captured := captureUGCShowcaseNamespaceLockHolderPostgreSQLBackendPID(t, env)
					env.provider.beforeCopy = func() {
						select {
						case <-releaseFirst:
						case <-ctx.Done():
						}
					}
					go publish()
					pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, captured)
					go compete()
				} else {
					locked := make(chan int, 1)
					callback := "pause_ugc_moderation_" + uuid.NewString()
					var pauseOnce sync.Once
					require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "showcases" {
							pauseOnce.Do(func() {
								pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseFirst)
							})
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Update().Remove(callback) })
					go compete()
					pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, locked)
					go publish()
				}
				requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
				release()
				publicationErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, publicationResult)
				competitorErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, competitorResult)
				if first == "publish" {
					require.NoError(t, publicationErr)
					if competitor == "reject" {
						require.ErrorIs(t, competitorErr, ErrShowcaseInvalidTransition)
					} else {
						require.NoError(t, competitorErr)
					}
				} else {
					require.ErrorIs(t, publicationErr, ErrShowcaseInvalidTransition)
					require.NoError(t, competitorErr)
					require.Empty(t, env.provider.copiedKeys)
				}
			})
		}
	}
}

func TestPostgreSQLUGCShowcaseCanceledProviderDeletionRetainsLockUntilReturnAndRejectsMissingRebinding(t *testing.T) {
	env := createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t)
	writeUGCShowcaseRegressionObject(t, env, env.key)
	event := createUGCShowcaseRegressionCleanupEvent(t, env, 0, env.key, true)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	cleanupContext, cancelCleanup := context.WithCancel(ctx)
	defer cancelCleanup()
	entered, releaseDelete := make(chan struct{}), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseDelete) }) }
	t.Cleanup(release)
	cleanupService := *env.service
	cleanupService.storage = &dedicatedPublicCleanupStorageThatDeletesOnlyAfterCancellationAndExplicitRelease{StorageService: env.provider.StorageService, entered: entered, release: releaseDelete}
	captured := captureUGCShowcaseNamespaceLockHolderPostgreSQLBackendPID(t, env)
	cleanupResult, bindingResult := make(chan error, 1), make(chan error, 1)
	go func() {
		cleanupResult <- NewObjectStorageCleanupOutboxHandler(nil, nil, nil, &cleanupService).Handle(cleanupContext, event)
	}()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, captured)
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("UGC cleanup did not enter provider")
	}
	encoded, _ := json.Marshal([]string{env.key})
	go func() {
		bindingResult <- env.service.repo.WithContext(ctx).Create(&ugcshowcase.UGCShowcase{UserID: 7, Status: ugcshowcase.StatusPending, Images: encoded})
	}()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
	cancelCleanup()
	returnedEarly := false
	select {
	case <-bindingResult:
		returnedEarly = true
		t.Error("UGC binding entered while canceled provider could still delete")
	case <-time.After(200 * time.Millisecond):
	}
	release()
	require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult), context.Canceled)
	if !returnedEarly {
		require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, bindingResult), ErrMediaAssetURLUnavailable)
	}
	requireUGCShowcaseRegressionObjectExists(t, env, env.key, false)
	_, err := env.service.cleanupUGCShowcaseObjectsWithRetainedReferencesAndLockedRejectedOwner(ctx, 0, []string{env.key}, true)
	require.NoError(t, err)
}

func TestPostgreSQLUGCShowcaseCanceledPublicationCopyRetainsNamespaceUntilProviderReturns(t *testing.T) {
	env := createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t)
	writeUGCShowcaseRegressionObject(t, env, env.key)
	owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	publicationContext, cancelPublication := context.WithCancel(ctx)
	defer cancelPublication()
	entered, releaseCopy := make(chan struct{}), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseCopy) }) }
	t.Cleanup(release)
	env.provider.beforeCopy = func() { close(entered); <-releaseCopy }
	captured := captureUGCShowcaseNamespaceLockHolderPostgreSQLBackendPID(t, env)
	publicationResult, bindingResult := make(chan error, 1), make(chan error, 1)
	go func() { publicationResult <- env.service.Approve(publicationContext, owner.ID) }()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, captured)
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("UGC publication did not enter copy")
	}
	encoded, _ := json.Marshal([]string{env.key})
	go func() {
		bindingResult <- env.service.repo.WithContext(ctx).Create(&ugcshowcase.UGCShowcase{UserID: 7, Status: ugcshowcase.StatusPending, Images: encoded})
	}()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
	cancelPublication()
	returnedEarly := false
	select {
	case <-bindingResult:
		returnedEarly = true
		t.Error("UGC namespace released before canceled copy returned")
	case <-time.After(200 * time.Millisecond):
	}
	release()
	require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, publicationResult), context.Canceled)
	if !returnedEarly {
		require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, bindingResult))
	}
	requireUGCShowcaseRegressionObjectExists(t, env, env.key, true)
}

func TestPostgreSQLUGCShowcaseLifecycleRejectsStaleOuterSnapshotsAndExternallyOwnedProviderTransactions(t *testing.T) {
	for _, operation := range []string{"binding", "cleanup", "publication"} {
		t.Run(operation, func(t *testing.T) {
			env := createIsolatedPostgreSQLUGCShowcaseLifecycleRegressionEnvironment(t)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
			err := env.db.Transaction(func(tx *gorm.DB) error {
				repo := env.service.repo.WithTx(tx)
				if operation == "binding" {
					encoded, _ := json.Marshal([]string{env.key})
					return repo.Create(&ugcshowcase.UGCShowcase{UserID: 7, Status: ugcshowcase.StatusPending, Images: encoded})
				}
				service := *env.service
				service.repo = repo
				if operation == "publication" {
					return service.Approve(t.Context(), owner.ID)
				}
				_, err := service.cleanupUGCShowcaseObjectsWithRetainedReferencesAndLockedRejectedOwner(t.Context(), 0, []string{env.key}, true)
				return err
			}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
			if operation == "binding" {
				require.ErrorContains(t, err, "requires read committed transaction isolation")
			} else {
				require.ErrorContains(t, err, "requires its own transaction lifetime")
			}
			require.Empty(t, env.provider.deletedKeys)
			require.Empty(t, env.provider.copiedKeys)
		})
	}
}
