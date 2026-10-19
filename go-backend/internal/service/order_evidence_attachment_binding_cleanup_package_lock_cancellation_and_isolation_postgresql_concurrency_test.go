package service

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t *testing.T) orderEvidenceCompensationRegressionEnvironment {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, base.db.AutoMigrate(&order.Order{}, &order.OrderItem{}, &orderevidence.OrderEvidenceSnapshot{}, &orderevidence.OrderEvidencePackage{}, &orderevidence.OrderEvidenceItem{}, &orderevidence.OrderEvidenceAttachment{}))
	return configureOrderEvidenceCompensationRegressionEnvironment(t, base.db)
}

func captureOrderEvidenceCleanupNamespaceHolderBackendPID(t *testing.T, env orderEvidenceCompensationRegressionEnvironment) <-chan int {
	t.Helper()
	captured := make(chan int, 1)
	var once sync.Once
	callback := "capture_order_evidence_cleanup_pid_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "order_evidence_attachments" {
			once.Do(func() {
				var pid int
				err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error
				if err != nil {
					tx.AddError(err)
				} else {
					captured <- pid
				}
			})
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callback) })
	return captured
}

func TestPostgreSQLOrderEvidenceAttachmentBindingAndCleanupSerializeInBothOrdersAndRollback(t *testing.T) {
	for _, first := range []string{"binding", "cleanup", "rolled back binding"} {
		t.Run(first, func(t *testing.T) {
			env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
			prefix, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			acquired, release := make(chan int, 1), make(chan struct{})
			var once sync.Once
			releaseFirst := func() { once.Do(func() { close(release) }) }
			t.Cleanup(releaseFirst)
			rollback := errors.New("evidence registration deliberately rolled back")
			if first != "cleanup" {
				callback := "pause_order_evidence_registration_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "order_evidence_attachments" {
						pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, release)
						if first == "rolled back binding" {
							tx.AddError(rollback)
						}
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			}
			bindingResult, cleanupResult := make(chan error, 1), make(chan error, 1)
			bind := func() {
				_, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
				bindingResult <- err
			}
			cleanup := func() {
				cleanupResult <- env.service.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(ctx, prefix, []string{input.StorageKey})
			}
			var pid int
			if first == "cleanup" {
				captured := captureOrderEvidenceCleanupNamespaceHolderBackendPID(t, env)
				env.provider.beforeDelete = func(context.Context) {
					select {
					case <-release:
					case <-ctx.Done():
					}
				}
				go cleanup()
				pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, captured)
				go bind()
			} else {
				go bind()
				pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
				go cleanup()
			}
			waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, pid, "pg_advisory_xact_lock")
			releaseFirst()
			bindingErr, cleanupErr := <-bindingResult, <-cleanupResult
			switch first {
			case "binding":
				require.NoError(t, bindingErr)
				require.ErrorIs(t, cleanupErr, ErrMediaAssetInUse)
			case "cleanup":
				require.Error(t, bindingErr)
				require.NoError(t, cleanupErr)
			case "rolled back binding":
				require.ErrorIs(t, bindingErr, rollback)
				require.NoError(t, cleanupErr)
			}
			assertOrderEvidenceRegressionObjectExists(t, env, input.StorageKey, first == "binding")
		})
	}
}

func TestPostgreSQLOrderEvidenceCanceledCleanupHoldsNamespaceUntilProviderReturns(t *testing.T) {
	env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
	prefix, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	release, providerObservedCancellation, providerEntered := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	releaseProvider := func() { once.Do(func() { close(release) }) }
	t.Cleanup(releaseProvider)
	captured := captureOrderEvidenceCleanupNamespaceHolderBackendPID(t, env)
	env.provider.beforeDelete = func(providerContext context.Context) {
		close(providerEntered)
		<-providerContext.Done()
		close(providerObservedCancellation)
		<-release
	}
	cleanupResult, bindingResult := make(chan error, 1), make(chan error, 1)
	go func() {
		cleanupResult <- env.service.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(ctx, prefix, []string{input.StorageKey})
	}()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t.Context(), t, captured)
	select {
	case <-providerEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("provider did not enter Delete")
	}
	// The scan precedes Delete, so wait until the provider receives cancellation.
	cancel()
	select {
	case <-providerObservedCancellation:
	case <-time.After(5 * time.Second):
		t.Fatal("provider did not enter Delete")
	}
	go func() {
		_, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
		bindingResult <- err
	}()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, pid, "pg_advisory_xact_lock")
	releaseProvider()
	require.ErrorIs(t, <-cleanupResult, context.Canceled)
	require.Error(t, <-bindingResult)
	assertOrderEvidenceRegressionObjectExists(t, env, input.StorageKey, false)
}

func TestPostgreSQLOrderEvidencePackageLockRejectsConcurrentRegistrationAndDetachment(t *testing.T) {
	for _, mutation := range []string{"register", "detach"} {
		t.Run(mutation, func(t *testing.T) {
			env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
			_, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
			attachment, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
			require.NoError(t, err)
			// Ready fixture: unrelated operational checks are explicitly waived.
			require.NoError(t, env.db.Exec("UPDATE order_evidence_items SET status='waived', captured_at=NOW(), captured_by=7, data_json='{}' WHERE package_id=? AND id<>?", env.pkg.ID, env.item.ID).Error)
			_, err = NewOrderEvidenceService().UpdateItem(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, OrderEvidenceItemUpdateInput{ItemID: env.item.ID, Status: orderevidence.EvidenceItemStatusComplete, DataJSON: []byte(`{"serial":"proof"}`), CapturedBy: 7})
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			acquired, release := make(chan int, 1), make(chan struct{})
			var once sync.Once
			releaseLock := func() { once.Do(func() { close(release) }) }
			t.Cleanup(releaseLock)
			callback := "pause_order_evidence_package_lock_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == "order_evidence_packages" {
					pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, release)
				}
			}))
			t.Cleanup(func() { _ = env.db.Callback().Update().Remove(callback) })
			lockResult, mutationResult := make(chan error, 1), make(chan error, 1)
			go func() {
				_, _, err := env.service.evidenceRepo.LockPackage(env.pkg.ID, time.Now().UTC())
				lockResult <- err
			}()
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
			go func() {
				if mutation == "register" {
					_, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo}, input)
					mutationResult <- err
				} else {
					mutationResult <- env.service.Delete(ctx, env.item.OrderID, env.item.ID, attachment.ID)
				}
			}()
			waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, env.db, pid, "pg_advisory_xact_lock")
			releaseLock()
			require.NoError(t, <-lockResult)
			require.ErrorIs(t, <-mutationResult, orderevidence.ErrOrderEvidencePackageLocked)
			assertOrderEvidenceRegressionObjectExists(t, env, input.StorageKey, true)
		})
	}
}

func TestPostgreSQLOrderEvidenceAttachmentRejectsStaleTransactionIsolation(t *testing.T) {
	for _, isolation := range []sql.IsolationLevel{sql.LevelRepeatableRead, sql.LevelSerializable} {
		t.Run(isolation.String(), func(t *testing.T) {
			env := createIsolatedPostgreSQLOrderEvidenceCompensationEnvironment(t)
			prefix, input := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
			err := env.db.Transaction(func(tx *gorm.DB) error {
				_, err := env.service.RegisterReference(repository.TxRepositories{OrderEvidence: env.service.evidenceRepo.WithTx(tx)}, input)
				return err
			}, &sql.TxOptions{Isolation: isolation})
			require.ErrorContains(t, err, "read committed")
			err = env.db.Transaction(func(tx *gorm.DB) error {
				nested := *env.service
				nested.evidenceRepo = env.service.evidenceRepo.WithTx(tx)
				return nested.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(t.Context(), prefix, []string{input.StorageKey})
			})
			require.ErrorContains(t, err, "own transaction lifetime")
			require.Zero(t, env.provider.deleteCalls)
		})
	}
}
