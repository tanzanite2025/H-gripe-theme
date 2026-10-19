package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type visualShowcaseDraftUploadControlledStorage struct {
	storage.StorageService
	writtenKeys             []string
	writeError              error
	returnReferenceOverride *string
	replacementBytes        []byte
	afterWrite              func()
}

func (s *visualShowcaseDraftUploadControlledStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *visualShowcaseDraftUploadControlledStorage) WriteVisualShowcaseDraftObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.writtenKeys = append(s.writtenKeys, key)
	if s.replacementBytes != nil {
		reader = bytes.NewReader(s.replacementBytes)
	}
	reference, err := s.StorageService.(storage.PreallocatedVisualShowcaseDraftObjectWriter).WriteVisualShowcaseDraftObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return reference, err
	}
	if s.afterWrite != nil {
		s.afterWrite()
	}
	if s.returnReferenceOverride != nil {
		reference = *s.returnReferenceOverride
	}
	return reference, s.writeError
}

func configureVisualShowcaseDraftUploadControlledStorage(t *testing.T, db *gorm.DB, provider storage.StorageService) (*HomeVisualTileService, *visualShowcaseDraftUploadControlledStorage) {
	t.Helper()
	controlled := &visualShowcaseDraftUploadControlledStorage{StorageService: provider}
	svc := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), controlled, "https://shop.example.test")
	svc.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return svc, controlled
}

func loadVisualShowcaseDraftUploadReceiptAndPayload(t *testing.T, db *gorm.DB, key string) (outbox.Event, homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload) {
	t.Helper()
	id, valid := storage.VisualShowcaseDraftUploadReceiptIDFromObjectKey(key)
	require.True(t, valid)
	var event outbox.Event
	require.NoError(t, db.Where("event_key = ?", homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix+id).First(&event).Error)
	var payload homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	return event, payload
}

func processVisualShowcaseDraftCleanupEventsAtTime(t *testing.T, db *gorm.DB, svc *HomeVisualTileService, now time.Time) OutboxDispatchResult {
	t.Helper()
	worker := NewOutboxService(repository.NewOutboxRepository(db))
	worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, NewObjectStorageCleanupOutboxHandler(nil, nil, svc, nil).Handle)
	result, err := worker.ProcessPending(t.Context(), now, 100)
	require.NoError(t, err)
	return result
}

func TestVisualShowcaseDraftUploadReadyReceiptHasExactIntegrityAndTwentyFourHourSaveWindowWithoutPublicAccess(t *testing.T) {
	db, svc := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	started := time.Now().UTC()
	file := multipartFileHeader(t, "showcase.png", "image/png", testOpaquePNG(t, 600, 600))
	upload, err := svc.UploadAdminImage(t.Context(), "home-hero", "en", file)
	require.NoError(t, err)
	require.Equal(t, upload.ImageURL, upload.ThumbnailURL)
	require.NotNil(t, upload.DraftExpiresAt)
	require.WithinRange(t, *upload.DraftExpiresAt, started.Add(24*time.Hour), time.Now().Add(24*time.Hour))
	event, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, upload.StorageKey)
	require.Equal(t, "ready", payload.UploadState)
	require.Equal(t, []string{upload.StorageKey}, payload.ObjectKeys)
	require.Equal(t, file.Size, payload.Size)
	digest, err := contentSHA256FromMultipartFile(file)
	require.NoError(t, err)
	require.Equal(t, digest, payload.ContentSHA256)
	require.False(t, payload.RequestedAt.IsZero())
	require.True(t, event.AvailableAt.Equal(*upload.DraftExpiresAt))
	require.WithinRange(t, payload.WriteDeadline, started, time.Now().Add(5*time.Minute))
	allowed, err := svc.CanServePublicVisualShowcaseImage(t.Context(), upload.StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
	result := processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, started.Add(23*time.Hour))
	require.Zero(t, result.Processed)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, upload.StorageKey, true)
	result = processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, upload.DraftExpiresAt.Add(time.Second))
	require.Equal(t, 1, result.Processed)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, upload.StorageKey, false)
}

func TestVisualShowcaseDraftFailedUploadKeepsExactPrecommittedReceiptForExistingCleanupWorker(t *testing.T) {
	for _, failure := range []string{"lost response", "error with URL", "empty success", "foreign URL", "wrong key", "wrong digest", "wrong size", "cancellation", "ready transition failure"} {
		t.Run(failure, func(t *testing.T) {
			db, base := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
			svc, provider := configureVisualShowcaseDraftUploadControlledStorage(t, db, base.storage)
			data := testOpaquePNG(t, 600, 600)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			override := ""
			removeFailureInjection := func() {}
			switch failure {
			case "lost response":
				provider.writeError = errors.New("provider response lost")
				provider.returnReferenceOverride = &override
			case "error with URL":
				provider.writeError = errors.New("provider failed after writing")
			case "empty success":
				provider.returnReferenceOverride = &override
			case "foreign URL":
				override = "https://foreign.example.test/uploads/visual-showcase/other.png"
				provider.returnReferenceOverride = &override
			case "wrong key":
				override = provider.GetURL("visual-showcase/home-hero/en/unrelated.png")
				provider.returnReferenceOverride = &override
			case "wrong digest":
				provider.replacementBytes = bytes.Repeat([]byte{'x'}, len(data))
			case "wrong size":
				provider.replacementBytes = []byte("short")
			case "cancellation":
				provider.afterWrite = cancel
			case "ready transition failure":
				callback := "reject_visual_draft_ready_transition_" + uuid.NewString()
				require.NoError(t, db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("ready transition failed"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Update().Remove(callback) })
				removeFailureInjection = func() { _ = db.Callback().Update().Remove(callback) }
			}
			upload, err := svc.UploadAdminImage(ctx, "home-hero", "en", multipartFileHeader(t, "showcase.png", "image/png", data))
			require.Error(t, err)
			require.Nil(t, upload)
			require.Len(t, provider.writtenKeys, 1)
			event, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, provider.writtenKeys[0])
			require.Equal(t, "uploading", payload.UploadState)
			require.Nil(t, payload.DraftExpiresAt)
			require.Equal(t, outbox.EventStatusPending, event.Status)
			require.Equal(t, []string{provider.writtenKeys[0]}, payload.ObjectKeys)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, provider.writtenKeys[0], true)
			// The failure injection must not affect the cleanup worker's status updates.
			removeFailureInjection()
			result := processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, time.Now().Add(6*time.Minute))
			require.Equal(t, 1, result.Processed)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, provider.writtenKeys[0], false)
		})
	}
}

func TestVisualShowcaseDraftReceiptPersistenceFailureAndUnsupportedWriterNeverStartObjectIO(t *testing.T) {
	for _, failure := range []string{"receipt insert", "missing outbox", "unsupported writer", "nested transaction"} {
		t.Run(failure, func(t *testing.T) {
			db, base := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
			svc, provider := configureVisualShowcaseDraftUploadControlledStorage(t, db, base.storage)
			switch failure {
			case "receipt insert":
				callback := "reject_visual_draft_receipt_insert_" + uuid.NewString()
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("receipt unavailable"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Create().Remove(callback) })
			case "missing outbox":
				svc.ConfigureObjectCleanupOutbox(nil)
			case "unsupported writer":
				svc.storage = struct{ storage.StorageService }{provider}
			case "nested transaction":
				tx := db.Begin()
				require.NoError(t, tx.Error)
				defer tx.Rollback()
				svc.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(tx))
			}
			_, err := svc.UploadAdminImage(t.Context(), "home-hero", "en", multipartFileHeader(t, "showcase.png", "image/png", testOpaquePNG(t, 600, 600)))
			require.Error(t, err)
			require.Empty(t, provider.writtenKeys)
		})
	}
}

func TestVisualShowcaseDraftOriginalAndIndependentThumbnailBindAtomicallyAndSurviveExpiry(t *testing.T) {
	for _, mode := range []string{"single card", "whole group", "same object across slots"} {
		t.Run(mode, func(t *testing.T) {
			db, svc := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
			original := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
			thumbnail := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 2)
			original.ThumbnailURL = thumbnail.ImageURL
			if mode == "single card" {
				_, err := svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", original)
				require.NoError(t, err)
			} else {
				items := []HomeVisualTileInput{original}
				if mode == "same object across slots" {
					second := original
					second.DesktopOrder = 2
					items = append(items, second)
				}
				_, err := svc.ReplaceAdminItems(t.Context(), "home-hero", "en", items)
				require.NoError(t, err)
			}
			var count int64
			require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
			require.Zero(t, count)
			result := processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, time.Now().Add(48*time.Hour))
			require.Zero(t, result.Processed)
			for _, key := range []string{original.StorageKey, thumbnail.StorageKey} {
				requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, key, true)
				allowed, err := svc.CanServePublicVisualShowcaseImage(t.Context(), key)
				require.NoError(t, err)
				require.True(t, allowed)
			}
			_, err := svc.ReplaceAdminItems(t.Context(), "home-hero", "en", nil)
			require.NoError(t, err)
			result = processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, time.Now().Add(time.Minute))
			require.Equal(t, 1, result.Processed)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, original.StorageKey, false)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, thumbnail.StorageKey, false)
		})
	}
}

func TestVisualShowcaseDraftBindingFailureRestoresBothReceiptsAndPreviousOwner(t *testing.T) {
	for _, failure := range []string{"thumbnail expired", "owner write", "old owner cleanup"} {
		t.Run(failure, func(t *testing.T) {
			db, svc := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
			old := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
			_, err := svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", old)
			require.NoError(t, err)
			original := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
			thumbnail := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 2)
			// Ensure the valid original receipt is consumed before validating the failing thumbnail.
			if original.StorageKey > thumbnail.StorageKey {
				original, thumbnail = thumbnail, original
				original.DesktopOrder = 1
			}
			original.ThumbnailURL = thumbnail.ImageURL
			if failure == "thumbnail expired" {
				event, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, thumbnail.StorageKey)
				expired := time.Now().Add(-time.Minute)
				payload.DraftExpiresAt = &expired
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				require.NoError(t, db.Model(&event).Update("payload", encoded).Error)
			} else if failure == "owner write" {
				callback := "reject_visual_draft_owner_write_" + uuid.NewString()
				require.NoError(t, db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "visual_showcase_items" {
						tx.AddError(errors.New("owner write failed"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Update().Remove(callback) })
			} else {
				svc.ConfigureObjectCleanupOutbox(nil)
			}
			var before []outbox.Event
			require.NoError(t, db.Order("id").Find(&before).Error)
			require.Len(t, before, 2)
			_, err = svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", original)
			require.Error(t, err)
			var after []outbox.Event
			require.NoError(t, db.Order("id").Find(&after).Error)
			require.Equal(t, before, after)
			items, err := svc.GetAdminItems("home-hero", "en")
			require.NoError(t, err)
			require.Len(t, items, 1)
			require.Equal(t, old.StorageKey, items[0].StorageKey)
			for _, key := range []string{old.StorageKey, original.StorageKey, thumbnail.StorageKey} {
				requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, key, true)
			}
		})
	}
}

func TestVisualShowcaseDraftBindingRejectsExpiredClaimedCompletedMissingAndChangedReceipts(t *testing.T) {
	for _, state := range []string{"expired with postponed retry", "uploading", "claimed", "completed", "ignored", "retried after claim", "missing", "changed object key"} {
		t.Run(state, func(t *testing.T) {
			db, svc := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
			input := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
			event, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, input.StorageKey)
			updates := map[string]interface{}{}
			switch state {
			case "expired with postponed retry":
				expired := time.Now().Add(-time.Minute)
				payload.DraftExpiresAt = &expired
				updates["available_at"] = time.Now().Add(48 * time.Hour)
			case "uploading":
				payload.UploadState = "uploading"
			case "claimed":
				updates["status"], updates["attempts"] = outbox.EventStatusProcessing, 1
			case "completed":
				updates["status"] = outbox.EventStatusProcessed
			case "ignored":
				updates["status"] = outbox.EventStatusIgnored
			case "retried after claim":
				updates["last_attempt_at"] = time.Now().Add(-time.Minute)
			case "missing":
				require.NoError(t, db.Unscoped().Delete(&event).Error)
			case "changed object key":
				payload.ObjectKeys = []string{"visual-showcase/home-hero/en/other.png"}
			}
			if state != "missing" {
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				updates["payload"] = encoded
				require.NoError(t, db.Model(&event).Updates(updates).Error)
			}
			_, err := svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", input)
			require.ErrorIs(t, err, ErrHomeVisualTileImageInvalid)
			items, err := svc.GetAdminItems("home-hero", "en")
			require.NoError(t, err)
			require.Empty(t, items)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, input.StorageKey, true)
		})
	}
}

func TestVisualShowcaseDraftExpiryRetainsHistoricalBusinessReferencesAndRetriesTheSameTaskAfterLastDetachment(t *testing.T) {
	for _, consumer := range []string{"product_deleted_placement", "product_description", "post_content"} {
		t.Run(consumer, func(t *testing.T) {
			db, svc := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
			require.NoError(t, db.AutoMigrate(&product.Product{}, &product.ProductMedia{}, &post.Post{}))
			input := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
			original, payload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, input.StorageKey)
			detach := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, db, input.StorageKey, consumer)
			result := processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, payload.DraftExpiresAt.Add(time.Second))
			require.Equal(t, 1, result.Failed)
			retained, retainedPayload := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, input.StorageKey)
			require.Equal(t, original.ID, retained.ID)
			require.Equal(t, original.EventKey, retained.EventKey)
			require.Equal(t, original.Payload, retained.Payload)
			require.True(t, payload.DraftExpiresAt.Equal(*retainedPayload.DraftExpiresAt))
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, input.StorageKey, true)
			detach()
			result = processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, retained.AvailableAt.Add(time.Second))
			require.Equal(t, 1, result.Processed)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, input.StorageKey, false)
			completed, _ := loadVisualShowcaseDraftUploadReceiptAndPayload(t, db, input.StorageKey)
			require.Equal(t, original.ID, completed.ID)
			require.Equal(t, outbox.EventStatusProcessed, completed.Status)
		})
	}
}

func TestVisualShowcaseDraftPreviouslySavedButOverwrittenOwnerWithoutShowcaseHistoryRequiresNewUpload(t *testing.T) {
	db, svc := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	require.NoError(t, db.AutoMigrate(&post.Post{}))
	old := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
	_, err := svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", old)
	require.NoError(t, err)
	detach := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, db, old.StorageKey, "post_content")
	replacement := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, svc, 1)
	_, err = svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", replacement)
	require.NoError(t, err)
	_, err = svc.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", old)
	require.ErrorIs(t, err, ErrHomeVisualTileImageInvalid)
	result := processVisualShowcaseDraftCleanupEventsAtTime(t, db, svc, time.Now().Add(time.Minute))
	require.Equal(t, 1, result.Failed)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, svc, old.StorageKey, true)
	detach()
}
