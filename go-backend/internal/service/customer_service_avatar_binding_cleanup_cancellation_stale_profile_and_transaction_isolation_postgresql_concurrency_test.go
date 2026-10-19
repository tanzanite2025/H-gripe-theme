package service

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLCustomerServiceAvatarLifecycleRegressionEnvironment(t *testing.T) (customerServiceAvatarDurableCompensationRegressionEnvironment, outbox.Event) {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	// The base environment has already migrated users. Create only the missing
	// owner table instead of traversing and migrating those associations again.
	require.NoError(t, base.db.Migrator().CreateTable(&user.AgentProfile{}))
	env := configureCustomerServiceAvatarDurableCompensationRegressionEnvironment(t, base.db, base.storage.StorageService, base.publicRoot)
	return env, createCustomerServiceAvatarDurableCleanupEventForRegression(t, env)
}

func captureCustomerServiceAvatarCleanupNamespaceLockBackendPID(t *testing.T, env customerServiceAvatarDurableCompensationRegressionEnvironment) <-chan int {
	t.Helper()
	acquired := make(chan int, 1)
	var once sync.Once
	callback := "capture_avatar_cleanup_namespace_lock_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "customer_service_agent_profiles" {
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

func TestPostgreSQLCustomerServiceAvatarBindingAndCleanupSerializeInBothOrdersAndReleaseRolledBackOwnership(t *testing.T) {
	for _, first := range []string{"binding", "cleanup", "rolled back binding"} {
		t.Run(first, func(t *testing.T) {
			env, event := createIsolatedPostgreSQLCustomerServiceAvatarLifecycleRegressionEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			locked, releaseFirst := make(chan int, 1), make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(releaseFirst) }) }
			t.Cleanup(release)
			bindingResult, cleanupResult := make(chan error, 1), make(chan error, 1)
			rollback := errors.New("avatar owner deliberately rolled back")
			bind := func() {
				_, err := env.service.userRepo.WithContext(ctx).ReplaceCustomerServiceAgentAvatar(*env.profile.UserID, env.provider.uploadReference, func(tx *gorm.DB, _ uint, _ string) error {
					if first == "cleanup" {
						return nil
					}
					pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseFirst)
					if first == "rolled back binding" {
						return rollback
					}
					return nil
				})
				bindingResult <- err
			}
			cleanup := func() { cleanupResult <- env.handler.Handle(ctx, event) }
			var pid int
			if first == "cleanup" {
				captured := captureCustomerServiceAvatarCleanupNamespaceLockBackendPID(t, env)
				env.provider.beforeDelete = func() {
					select {
					case <-releaseFirst:
					case <-ctx.Done():
					}
				}
				go cleanup()
				pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, captured)
				go bind()
			} else {
				go bind()
				pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, locked)
				go cleanup()
			}
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, env.db, pid)
			release()
			bindingError := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, bindingResult)
			cleanupError := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult)
			if first == "binding" {
				require.NoError(t, bindingError)
				require.ErrorIs(t, cleanupError, ErrMediaAssetInUse)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				require.NoError(t, env.service.Remove(ctx, *env.profile.UserID))
				require.NoError(t, env.handler.Handle(ctx, event))
			} else {
				require.NoError(t, cleanupError)
				if first == "cleanup" {
					require.ErrorIs(t, bindingError, ErrMediaAssetURLUnavailable)
				} else {
					require.ErrorIs(t, bindingError, rollback)
				}
			}
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
			var saved user.AgentProfile
			require.NoError(t, env.db.First(&saved, env.profile.ID).Error)
			require.Empty(t, saved.Avatar)
		})
	}
}

func TestPostgreSQLCustomerServiceAvatarCleanupKeepsNamespaceLockUntilCanceledProviderDeletionReturns(t *testing.T) {
	env, event := createIsolatedPostgreSQLCustomerServiceAvatarLifecycleRegressionEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	cleanupContext, cancelCleanup := context.WithCancel(ctx)
	defer cancelCleanup()
	entered, releaseDelete := make(chan struct{}), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseDelete) }) }
	t.Cleanup(release)
	env.handler.storage = &dedicatedPublicCleanupStorageThatDeletesOnlyAfterCancellationAndExplicitRelease{StorageService: env.provider.StorageService, entered: entered, release: releaseDelete}
	captured := captureCustomerServiceAvatarCleanupNamespaceLockBackendPID(t, env)
	cleanupResult, bindingResult := make(chan error, 1), make(chan error, 1)
	go func() { cleanupResult <- env.handler.Handle(cleanupContext, event) }()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, captured)
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("avatar cleanup never reached provider deletion")
	}
	go func() {
		_, err := env.service.userRepo.WithContext(ctx).ReplaceCustomerServiceAgentAvatar(*env.profile.UserID, env.provider.uploadReference, nil)
		bindingResult <- err
	}()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, env.db, pid)
	cancelCleanup()
	returnedEarly := false
	select {
	case <-bindingResult:
		returnedEarly = true
		t.Error("avatar binding finished while the canceled provider could still delete its object")
	case <-time.After(200 * time.Millisecond):
	}
	release()
	require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult), context.Canceled)
	if !returnedEarly {
		require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, bindingResult), ErrMediaAssetURLUnavailable)
	}
	requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
	env.handler.storage = env.provider
	require.NoError(t, env.handler.Handle(ctx, event))
}

func TestPostgreSQLCustomerServiceAvatarCanceledCleanupWaiterDoesNotDeleteAndReleasesItsTransaction(t *testing.T) {
	env, event := createIsolatedPostgreSQLCustomerServiceAvatarLifecycleRegressionEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	locked, releaseBinding := make(chan int, 1), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseBinding) }) }
	t.Cleanup(release)
	bindingResult := make(chan error, 1)
	go func() {
		_, err := env.service.userRepo.WithContext(ctx).ReplaceCustomerServiceAgentAvatar(*env.profile.UserID, env.provider.uploadReference, func(tx *gorm.DB, _ uint, _ string) error {
			pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseBinding)
			return nil
		})
		bindingResult <- err
	}()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, locked)
	waitContext, cancelWait := context.WithCancel(ctx)
	defer cancelWait()
	cleanupResult := make(chan error, 1)
	go func() { cleanupResult <- env.handler.Handle(waitContext, event) }()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, env.db, pid)
	cancelWait()
	require.Error(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult))
	require.Empty(t, env.provider.deletedKeys)
	release()
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, bindingResult))
	require.ErrorIs(t, env.handler.Handle(ctx, event), ErrMediaAssetInUse)
	require.NoError(t, env.service.Remove(ctx, *env.profile.UserID))
	require.NoError(t, env.handler.Handle(ctx, event))
	requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
}

func TestPostgreSQLCustomerServiceAvatarStaleProfileMetadataSaveDuringCleanupCannotRestoreRemovedAvatar(t *testing.T) {
	env, event := createIsolatedPostgreSQLCustomerServiceAvatarLifecycleRegressionEnvironment(t)
	require.NoError(t, env.db.Model(env.profile).Update("avatar", env.provider.uploadReference).Error)
	var stale user.AgentProfile
	require.NoError(t, env.db.First(&stale, env.profile.ID).Error)
	require.NoError(t, env.service.Remove(t.Context(), *env.profile.UserID))
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	entered, releaseDelete := make(chan struct{}), make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseDelete) }) }
	t.Cleanup(release)
	env.provider.beforeDelete = func() {
		close(entered)
		select {
		case <-releaseDelete:
		case <-ctx.Done():
		}
	}
	cleanupResult := make(chan error, 1)
	go func() { cleanupResult <- env.handler.Handle(ctx, event) }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("avatar cleanup did not reach its provider")
	}
	stale.Name = "Updated while cleanup runs"
	require.NoError(t, env.service.userRepo.WithContext(ctx).UpdateCustomerServiceAgentProfile(&stale))
	release()
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult))
	var saved user.AgentProfile
	require.NoError(t, env.db.First(&saved, stale.ID).Error)
	require.Equal(t, stale.Name, saved.Name)
	require.Empty(t, saved.Avatar)
	requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
}

func TestPostgreSQLCustomerServiceAvatarBindingRejectsOuterStaleSnapshotAndCleanupRejectsExternalTransactionLifetime(t *testing.T) {
	for _, operation := range []string{"binding in repeatable read", "cleanup in outer transaction"} {
		t.Run(operation, func(t *testing.T) {
			env, _ := createIsolatedPostgreSQLCustomerServiceAvatarLifecycleRegressionEnvironment(t)
			err := env.db.Transaction(func(tx *gorm.DB) error {
				repo := repository.NewUserRepository(tx)
				if operation == "cleanup in outer transaction" {
					return repo.WithLockedCustomerServiceAvatarNamespaceForPhysicalCleanup(t.Context(), func(*repository.MediaRepository) error {
						t.Error("outer cleanup reached provider callback")
						return nil
					})
				}
				NewCustomerServiceAvatarService(repo, env.provider, repository.NewOutboxRepository(env.db))
				_, err := repo.WithContext(t.Context()).ReplaceCustomerServiceAgentAvatar(*env.profile.UserID, env.provider.uploadReference, nil)
				return err
			}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
			if operation == "cleanup in outer transaction" {
				require.ErrorContains(t, err, "requires its own transaction lifetime")
			} else {
				require.ErrorContains(t, err, "requires read committed transaction isolation")
			}
			require.Empty(t, env.provider.deletedKeys)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
			_, err = env.service.userRepo.WithContext(t.Context()).ReplaceCustomerServiceAgentAvatar(*env.profile.UserID, env.provider.uploadReference, nil)
			require.NoError(t, err)
		})
	}
}
