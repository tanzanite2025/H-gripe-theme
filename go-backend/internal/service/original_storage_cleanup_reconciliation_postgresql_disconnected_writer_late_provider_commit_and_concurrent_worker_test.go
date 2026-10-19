package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dedicatedUploadStorageWithRemoteWriteDelayedUntilOriginalCleanupCompletes struct {
	storage.StorageService
	entered chan struct{}
	release chan struct{}
}

func (s *dedicatedUploadStorageWithRemoteWriteDelayedUntilOriginalCleanupCompletes) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *dedicatedUploadStorageWithRemoteWriteDelayedUntilOriginalCleanupCompletes) WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key, cacheControl string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	close(s.entered)
	<-s.release
	// An already accepted remote request can complete after its client's DB
	// connection and context have disappeared. Model that completion explicitly.
	reference, err := s.StorageService.(storage.PreallocatedDedicatedImmediateOwnerObjectWriter).WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(context.WithoutCancel(ctx), bytes.NewReader(data), key, cacheControl)
	return reference, errors.Join(err, errors.New("remote response arrived after database disconnect"))
}

func TestPostgreSQLOriginalStorageCleanupReconciliationRemovesRemoteWriteCommittedAfterWriterDatabaseDisconnectAndSuccessfulCleanup(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createIsolatedPostgreSQLDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			delayed := &dedicatedUploadStorageWithRemoteWriteDelayedUntilOriginalCleanupCompletes{StorageService: env.provider.StorageService, entered: make(chan struct{}), release: make(chan struct{})}
			env.provider.StorageService = delayed
			var releaseOnce, captureOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(delayed.release) }) }
			defer release()
			pidChannel := make(chan int, 1)
			callback := "capture_disconnected_late_remote_upload_transaction_" + uuid.NewString()
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
			uploadFinished := make(chan error, 1)
			go func() { uploadFinished <- env.upload(t.Context()) }()
			select {
			case <-delayed.entered:
			case <-time.After(10 * time.Second):
				t.Fatal("remote upload did not start")
			}
			var pid int
			select {
			case pid = <-pidChannel:
			case <-time.After(5 * time.Second):
				t.Fatal("writer database transaction was not captured")
			}
			var terminated bool
			require.NoError(t, env.db.Raw("SELECT pg_terminate_backend(?)", pid).Scan(&terminated).Error)
			require.True(t, terminated)
			key := env.provider.writtenKeys[0]
			worker := NewOutboxService(repository.NewOutboxRepository(env.db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, env.handle)
			worker.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, env.handle)
			result, err := worker.ProcessPending(t.Context(), time.Now().Add(6*time.Minute), 100)
			require.NoError(t, err)
			require.Equal(t, 1, result.Processed)
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, key)
			release()
			select {
			case err := <-uploadFinished:
				require.Error(t, err)
			case <-time.After(10 * time.Second):
				t.Fatal("disconnected writer did not return")
			}
			require.NoError(t, delayed.StorageService.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
			check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(2*time.Hour), 100)
			require.NoError(t, err)
			require.Equal(t, 1, check.Reconciled)
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, key)
			require.Empty(t, env.currentKey())
		})
	}
}

func TestPostgreSQLOriginalStorageCleanupReconciliationClaimsAreDisjointAndUseDatabaseLeaseClock(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	now := time.Now().UTC()
	for index := range 12 {
		event := outbox.Event{EventKey: uuid.NewString(), EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "test", AggregateID: "1", Status: "processed", UpdatedAt: now.Add(-time.Hour), CreatedAt: now.Add(-time.Hour)}
		require.NoError(t, env.db.Create(&event).Error, index)
	}
	type claimResult struct {
		events []outbox.Event
		err    error
	}
	start := make(chan struct{})
	results := make(chan claimResult, 3)
	for range 3 {
		go func() {
			<-start
			events, err := repository.NewOutboxRepository(env.db).ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now, uuid.NewString(), 4, time.Minute)
			results <- claimResult{events, err}
		}()
	}
	close(start)
	seen := make(map[uint]bool)
	for range 3 {
		result := <-results
		require.NoError(t, result.err)
		require.Len(t, result.events, 4)
		for _, event := range result.events {
			require.False(t, seen[event.ID])
			seen[event.ID] = true
		}
	}
	require.Len(t, seen, 12)
	repo := repository.NewOutboxRepository(env.db)
	claimed, err := repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now.Add(24*time.Hour), "fast-client-clock", 100, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)
	require.NoError(t, env.db.Model(&outbox.Event{}).Where("id IN ?", mapKeysForClaimedOriginalCleanupReceipts(seen)).UpdateColumn("locked_at", gorm.Expr("NOW() - INTERVAL '2 minutes'")).Error)
	claimed, err = repo.ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now, "recovered-worker", 100, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 12)
}

func mapKeysForClaimedOriginalCleanupReceipts(seen map[uint]bool) []uint {
	ids := make([]uint, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

func TestPostgreSQLOriginalStorageCleanupReconciliationMigrationPreservesOriginalAuthorizationAndWriterFence(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	event := outbox.Event{EventKey: "original-migration-proof", EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "media_asset", AggregateID: "1", Payload: []byte(`{"resource_type":"unregistered_media_upload","resource_id":"attempt","object_keys":["exact.png"]}`), Status: "processed", Attempts: 7, LastAttemptAt: ptrTime(time.Now().UTC()), ReconcileAfter: ptrTime(time.Now().Add(time.Hour).UTC())}
	require.NoError(t, env.db.Create(&event).Error)
	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "427_add_original_storage_cleanup_continuous_late_write_reconciliation.up.sql"))
	require.NoError(t, err)
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "427_add_original_storage_cleanup_continuous_late_write_reconciliation.down.sql"))
	require.NoError(t, err)
	for range 2 {
		require.NoError(t, env.db.Exec(string(up)).Error)
	}
	var indexExists bool
	require.NoError(t, env.db.Raw("SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE schemaname=current_schema() AND indexname='idx_outbox_original_storage_cleanup_reconciliation_due')").Scan(&indexExists).Error)
	require.True(t, indexExists)
	require.NoError(t, env.db.Exec(string(down)).Error)
	require.NoError(t, env.db.Exec(string(up)).Error)
	var saved outbox.Event
	require.NoError(t, env.db.First(&saved, event.ID).Error)
	require.JSONEq(t, string(event.Payload), string(saved.Payload))
	require.Equal(t, event.Status, saved.Status)
	require.Equal(t, event.Attempts, saved.Attempts)
	require.WithinDuration(t, *event.LastAttemptAt, *saved.LastAttemptAt, time.Microsecond)
	require.WithinDuration(t, *event.ReconcileAfter, *saved.ReconcileAfter, time.Microsecond)
}
