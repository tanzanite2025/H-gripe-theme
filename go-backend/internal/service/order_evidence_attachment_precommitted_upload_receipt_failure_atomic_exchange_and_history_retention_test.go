package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func loadOrderEvidencePrecommittedUploadReceiptAndPayload(t *testing.T, db *gorm.DB) (outbox.Event, outbox.ObjectStorageCleanupPayload) {
	t.Helper()
	var event outbox.Event
	require.NoError(t, db.Where("event_key LIKE ?", repository.OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix+"%").First(&event).Error)
	var payload outbox.ObjectStorageCleanupPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, objectCleanupResourceFailedOrderEvidenceUpload, payload.ResourceType)
	require.Len(t, payload.ObjectKeys, 1)
	require.NoError(t, validateOrderEvidenceFailedUploadAttemptNamespace(payload.ResourceID))
	require.True(t, strings.HasPrefix(payload.ObjectKeys[0], payload.ResourceID+"/"))
	return event, payload
}

func processOrderEvidencePrecommittedUploadReceiptsWithExistingWorker(t *testing.T, env orderEvidenceCompensationRegressionEnvironment, now time.Time) OutboxDispatchResult {
	t.Helper()
	handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
	handler.ConfigureOrderEvidenceAttachmentService(env.service)
	worker := NewOutboxService(repository.NewOutboxRepository(env.db))
	worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, handler.Handle)
	result, err := worker.ProcessPending(t.Context(), now, 100)
	require.NoError(t, err)
	return result
}

func TestOrderEvidencePrecommittedUploadRetainsExactReceiptAcrossProviderIntegrityAndAtomicExchangeFailures(t *testing.T) {
	for _, scenario := range []string{"written without response", "empty success", "foreign response", "wrong exact key", "wrong size", "wrong digest", "missing bytes", "canceled after write", "attachment rollback", "receipt exchange rollback"} {
		t.Run(scenario, func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch scenario {
			case "written without response":
				env.provider.uploadError = errors.New("provider lost response after writing")
				env.provider.preallocatedResponse = func(string) string { return "" }
			case "empty success":
				env.provider.preallocatedResponse = func(string) string { return "" }
			case "foreign response":
				env.provider.preallocatedResponse = func(string) string { return "https://foreign.example/proof.png" }
			case "wrong exact key":
				env.provider.preallocatedResponse = func(reference string) string { return reference + ".other" }
			case "wrong size":
				env.provider.preallocatedReader = func(reader io.Reader) io.Reader { return io.MultiReader(reader, strings.NewReader("extra")) }
			case "wrong digest":
				env.provider.preallocatedReader = func(reader io.Reader) io.Reader {
					data, err := io.ReadAll(reader)
					require.NoError(t, err)
					data[len(data)-1] ^= 1
					return bytes.NewReader(data)
				}
			case "missing bytes":
				env.provider.skipPreallocatedWrite = true
				env.provider.preallocatedResponse = func(string) string { return env.provider.preallocatedKeys[0] }
			case "canceled after write":
				env.provider.afterPreallocatedWrite = func(context.Context, string) { cancel() }
			case "attachment rollback":
				callback := "reject_evidence_precommitted_owner_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "order_evidence_attachments" {
						tx.AddError(errors.New("owner rollback"))
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			case "receipt exchange rollback":
				callback := "reject_evidence_precommitted_exchange_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("receipt exchange rollback"))
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
			}
			attachment, err := env.service.Upload(ctx, env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
			require.Error(t, err)
			require.Nil(t, attachment)
			require.Len(t, env.provider.preallocatedKeys, 1)
			receipt, payload := loadOrderEvidencePrecommittedUploadReceiptAndPayload(t, env.db)
			require.Equal(t, env.provider.preallocatedKeys, payload.ObjectKeys)
			require.Equal(t, outbox.EventStatusPending, receipt.Status)
			require.Zero(t, env.provider.deleteCalls)
			var count int64
			require.NoError(t, env.db.Unscoped().Model(&orderevidence.OrderEvidenceAttachment{}).Count(&count).Error)
			require.Zero(t, count)
			require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
			require.EqualValues(t, 1, count)
			assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], scenario != "missing bytes")
			result := processOrderEvidencePrecommittedUploadReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute))
			require.Equal(t, 1, result.Processed)
			assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], false)
			var original outbox.Event
			require.NoError(t, env.db.First(&original, receipt.ID).Error)
			require.Equal(t, receipt.EventKey, original.EventKey)
			require.JSONEq(t, string(receipt.Payload), string(original.Payload))
			require.Equal(t, outbox.EventStatusProcessed, original.Status)
		})
	}
}

func TestOrderEvidencePrecommittedUploadGuardsRejectBeforeObjectIO(t *testing.T) {
	for _, scenario := range []string{"missing outbox", "receipt persistence failure", "unsupported writer", "canceled request", "locked package", "superseded package", "wrong order", "outer owner transaction", "outer receipt transaction"} {
		t.Run(scenario, func(t *testing.T) {
			env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			orderID := env.item.OrderID
			switch scenario {
			case "missing outbox":
				env.service.outbox = nil
			case "receipt persistence failure":
				require.NoError(t, env.db.Migrator().DropTable(&outbox.Event{}))
			case "unsupported writer":
				env.service.storage = struct{ storage.StorageService }{env.provider}
			case "canceled request":
				cancel()
			case "locked package":
				require.NoError(t, env.db.Model(env.pkg).Update("status", orderevidence.PackageStatusLocked).Error)
			case "superseded package":
				require.NoError(t, env.db.Model(env.pkg).Update("status", orderevidence.PackageStatusSuperseded).Error)
			case "wrong order":
				orderID++
			}
			file := newEvidencePNGFileHeader(t, "proof.png")
			var err error
			if strings.HasPrefix(scenario, "outer") {
				err = env.db.Transaction(func(tx *gorm.DB) error {
					nested := *env.service
					if scenario == "outer owner transaction" {
						nested.evidenceRepo = nested.evidenceRepo.WithTx(tx)
					} else {
						nested.outbox = nested.outbox.WithTx(tx)
					}
					_, err := nested.Upload(ctx, orderID, env.item.ID, file, 7)
					return err
				})
			} else {
				_, err = env.service.Upload(ctx, orderID, env.item.ID, file, 7)
			}
			require.Error(t, err)
			require.Empty(t, env.provider.preallocatedKeys)
			require.Zero(t, env.provider.deleteCalls)
			if scenario != "receipt persistence failure" {
				var count int64
				require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
				require.Zero(t, count)
			}
		})
	}
}

func TestOrderEvidencePrecommittedSuccessfulUploadConsumesOnlyItsReceiptAndRetainsImmutableDetachedOwner(t *testing.T) {
	env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
	oldPrefix, oldInput := writeOrderEvidenceCompensationCandidateAndReferenceInput(t, env)
	oldEvent, err := newObjectStorageCleanupEvent(objectCleanupResourceFailedOrderEvidenceUpload, oldPrefix, "order_evidence", uuid.NewString(), []string{oldInput.StorageKey})
	require.NoError(t, err)
	require.NoError(t, repository.NewOutboxRepository(env.db).CreateEvent(oldEvent))
	var committedReceipt outbox.Event
	callback := "capture_evidence_successful_receipt_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
		if event, ok := tx.Statement.Dest.(*outbox.Event); ok && strings.HasPrefix(event.EventKey, repository.OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix) {
			committedReceipt = *event
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
	env.provider.beforePreallocatedWrite = func(ctx context.Context, key string) {
		receipt := committedReceipt
		require.NotZero(t, receipt.ID)
		var payload outbox.ObjectStorageCleanupPayload
		require.NoError(t, json.Unmarshal(receipt.Payload, &payload))
		require.Equal(t, []string{key}, payload.ObjectKeys)
		require.True(t, receipt.AvailableAt.After(time.Now()))
		require.True(t, receipt.AvailableAt.Before(time.Now().Add(5*time.Minute+time.Second)))
	}
	attachment, err := env.service.Upload(t.Context(), env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
	require.NoError(t, err)
	require.True(t, storage.IsPrivateObjectKey(attachment.StorageKey))
	var remaining []outbox.Event
	require.NoError(t, env.db.Find(&remaining).Error)
	require.Len(t, remaining, 1)
	require.Equal(t, oldEvent.ID, remaining[0].ID)
	require.JSONEq(t, string(oldEvent.Payload), string(remaining[0].Payload))
	require.NoError(t, env.service.Delete(t.Context(), env.item.OrderID, env.item.ID, attachment.ID))
	handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
	handler.ConfigureOrderEvidenceAttachmentService(env.service)
	require.ErrorIs(t, handler.Handle(t.Context(), committedReceipt), ErrMediaAssetInUse)
	assertOrderEvidenceRegressionObjectExists(t, env, attachment.StorageKey, true)
}

func TestOrderEvidencePrecommittedReceiptReplaysSameTaskAfterSnapshotLastDetachment(t *testing.T) {
	env := configureOrderEvidenceCompensationRegressionEnvironment(t, newOrderEvidenceServiceTestDB(t))
	env.provider.uploadError = errors.New("upload response lost")
	_, err := env.service.Upload(t.Context(), env.item.OrderID, env.item.ID, newEvidencePNGFileHeader(t, "proof.png"), 7)
	require.Error(t, err)
	original, payload := loadOrderEvidencePrecommittedUploadReceiptAndPayload(t, env.db)
	snapshot, err := json.Marshal(map[string]string{"storage_key": payload.ObjectKeys[0]})
	require.NoError(t, err)
	require.NoError(t, env.db.Exec("UPDATE order_evidence_snapshots SET snapshot_data = ?", string(snapshot)).Error)
	result := processOrderEvidencePrecommittedUploadReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute))
	require.Equal(t, 1, result.Failed)
	require.Zero(t, env.provider.deleteCalls)
	require.NoError(t, env.db.Exec("UPDATE order_evidence_snapshots SET snapshot_data = '{}' ").Error)
	result = processOrderEvidencePrecommittedUploadReceiptsWithExistingWorker(t, env, time.Now().Add(24*time.Hour))
	require.Equal(t, 1, result.Processed)
	var final outbox.Event
	require.NoError(t, env.db.First(&final, original.ID).Error)
	require.Equal(t, original.EventKey, final.EventKey)
	require.JSONEq(t, string(original.Payload), string(final.Payload))
	assertOrderEvidenceRegressionObjectExists(t, env, payload.ObjectKeys[0], false)
}
