package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOriginalStorageCleanupReconciliationRemovesObjectsArrivingAfterRepeatedSuccessfulAbsenceChecks(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			env.provider.writeError = errors.New("provider response lost")
			require.Error(t, env.upload(t.Context()))
			key := env.provider.writtenKeys[0]
			worker := NewOutboxService(repository.NewOutboxRepository(env.db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, env.handle)
			worker.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, env.handle)
			result, err := worker.ProcessPending(t.Context(), time.Now().Add(6*time.Minute), 100)
			require.NoError(t, err)
			require.Equal(t, 1, result.Processed)
			for index := 0; index < 3; index++ {
				check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(time.Duration(index+1)*24*time.Hour), 100)
				require.NoError(t, err)
				require.Equal(t, 1, check.Reconciled)
			}
			// Simulate arbitrarily delayed remote completion after several absences.
			_, err = env.provider.StorageService.(storage.PreallocatedDedicatedImmediateOwnerObjectWriter).WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("late remote bytes")), key, storage.DedicatedImmediateOwnerUploadCacheControl)
			require.NoError(t, err)
			// A fresh service instance proves the schedule is persisted in SQL.
			restarted := NewOutboxService(repository.NewOutboxRepository(env.db))
			restarted.RegisterHandler(outbox.EventTypeObjectStorageCleanup, env.handle)
			restarted.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, env.handle)
			check, err := restarted.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(4*24*time.Hour), 100)
			require.NoError(t, err)
			require.Equal(t, 1, check.Reconciled)
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, key)
			var saved outbox.Event
			require.NoError(t, env.db.First(&saved).Error)
			require.Equal(t, "processed", saved.Status)
			require.Equal(t, 1, saved.Attempts)
			require.NotNil(t, saved.LastAttemptAt)
			require.NotNil(t, saved.ReconcileAfter)
			require.Empty(t, env.currentKey())
		})
	}
}

func TestOriginalStorageCleanupReconciliationRetainsOriginalTaskWhileHistoricalContentReferencesExist(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, "visual-showcase"} {
		t.Run(namespace, func(t *testing.T) {
			db, _, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
			root := t.TempDir()
			provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://dedicated.example.test/assets"})
			require.NoError(t, err)
			handler.siteLogo.storage, handler.siteFavicon.storage, handler.homeVisual.storage = provider, provider, provider
			worker := NewOutboxService(repository.NewOutboxRepository(db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, handler.Handle)
			_, err = worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 10)
			require.NoError(t, err)
			objectPath := filepath.Join(root, filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o700))
			require.NoError(t, os.WriteFile(objectPath, []byte("late bytes"), 0o600))
			detach := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, db, key, "product_description")
			check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(2*time.Hour), 10)
			require.NoError(t, err)
			require.Equal(t, 1, check.Deferred)
			require.NoError(t, provider.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
			var retained outbox.Event
			require.NoError(t, db.First(&retained, event.ID).Error)
			require.Equal(t, "processed", retained.Status)
			require.NotEmpty(t, retained.StorageCleanupReconciliationError)
			detach()
			check, err = worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), retained.ReconcileAfter.Add(time.Second), 10)
			require.NoError(t, err)
			require.Equal(t, 1, check.Reconciled)
			require.Error(t, provider.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
		})
	}
}

func TestOriginalStorageCleanupReconciliationUsesExactReceiptAfterRegistrationRemovalAndProtectsNewOwners(t *testing.T) {
	for _, retainedBy := range []string{"none", "new shared registration", "historical body", "dedicated same name", "missing exact receipt keys"} {
		t.Run(retainedBy, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{})
			root := t.TempDir()
			svc := newMediaDeleteTestService(t, db, root)
			asset, err := svc.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "retired.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
			require.NoError(t, err)
			require.NoError(t, svc.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
			worker := NewOutboxService(repository.NewOutboxRepository(db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, NewObjectStorageCleanupOutboxHandler(svc, nil, nil, nil).Handle)
			_, err = worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 100)
			require.NoError(t, err)
			key := asset.StorageKey
			if retainedBy == "dedicated same name" || retainedBy == "missing exact receipt keys" {
				if retainedBy == "dedicated same name" {
					key = SiteLogoStoragePrefix + "/same-name.png"
				}
				// A stale generic original never gains dedicated ownership.
				var event outbox.Event
				require.NoError(t, db.First(&event).Error)
				var payload outbox.ObjectStorageCleanupPayload
				require.NoError(t, json.Unmarshal(event.Payload, &payload))
				payload.ObjectKeys = []string{key}
				if retainedBy == "missing exact receipt keys" {
					payload.ObjectKeys = nil
				}
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				require.NoError(t, db.Model(&event).Update("payload", encoded).Error)
			}
			objectPath := filepath.Join(root, filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o700))
			require.NoError(t, os.WriteFile(objectPath, []byte("late remote bytes"), 0o600))
			if retainedBy == "new shared registration" {
				require.NoError(t, db.Create(&media.MediaAsset{Filename: "retained.png", URL: svc.storage.GetURL(key), StorageKey: key}).Error)
			}
			if retainedBy == "historical body" {
				item := product.Product{SKU: "history", Slug: "history", Name: "History", Description: `<img src="/uploads/` + key + `">`}
				require.NoError(t, db.Create(&item).Error)
				require.NoError(t, db.Delete(&item).Error)
			}
			check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(2*time.Hour), 100)
			require.NoError(t, err)
			_, objectErr := os.Stat(objectPath)
			if retainedBy == "none" {
				require.Equal(t, 1, check.Reconciled)
				require.ErrorIs(t, objectErr, os.ErrNotExist)
			} else {
				require.Equal(t, 1, check.Deferred)
				require.NoError(t, objectErr)
			}
		})
	}
}

func TestOriginalStorageCleanupReconciliationPreservesUnknownOutcomeAndPersistsFailuresWithoutReopeningUpload(t *testing.T) {
	db, worker := newTestOutboxService(t)
	now := time.Now().UTC()
	event := outbox.Event{EventKey: "unknown-delete", EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "media_asset", AggregateID: "1", Status: "unknown", Attempts: 3, LastAttemptAt: &now, LastError: "original unknown outcome"}
	require.NoError(t, db.Create(&event).Error)
	check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), now.Add(time.Minute), 10)
	require.NoError(t, err)
	require.Equal(t, 1, check.Deferred, "result: %+v, original: %+v", check, event)
	var saved outbox.Event
	require.NoError(t, db.First(&saved, event.ID).Error)
	require.Contains(t, saved.StorageCleanupReconciliationError, "no cleanup handler")
	worker.RegisterHandler(event.EventType, func(context.Context, outbox.Event) error { return nil })
	check, err = worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), saved.ReconcileAfter.Add(time.Second), 10)
	require.NoError(t, err)
	require.Equal(t, 1, check.Reconciled)
	require.NoError(t, db.First(&saved, event.ID).Error)
	require.Equal(t, "unknown", saved.Status)
	require.Equal(t, 3, saved.Attempts)
	require.Equal(t, event.LastError, saved.LastError)
	require.Empty(t, saved.StorageCleanupReconciliationError)
	require.NotNil(t, saved.ReconcileAfter)
}

func TestOriginalStorageCleanupReconciliationRevalidatesSuggestionAndPayPalReceiptUnderLifecycleLock(t *testing.T) {
	for _, namespace := range []string{storage.SuggestionFeedbackAttachmentNamespace, storage.PayPalDisputeInvoiceAttachmentNamespace} {
		t.Run(namespace, func(t *testing.T) {
			db, _, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
			data, filename := testOpaquePNG(t, 24, 24), "customer.png"
			if namespace == storage.PayPalDisputeInvoiceAttachmentNamespace {
				data, filename = []byte("%PDF-synthetic-invoice"), "invoice.pdf"
			}
			_, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(t.Context(), namespace, 71, filename, data)
			require.NoError(t, err)
			key := provider.writtenKeys[0]
			event, _ := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, key)
			handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
			handler.ConfigureSuggestionAndPayPalAttachmentUploads(uploads)
			worker := NewOutboxService(repository.NewOutboxRepository(db))
			worker.RegisterHandler(event.EventType, handler.Handle)
			result, err := worker.ProcessPending(t.Context(), event.AvailableAt.Add(time.Second), 100)
			require.NoError(t, err)
			require.Equal(t, 1, result.Processed)
			_, err = provider.StorageService.(storage.PreallocatedSuggestionOrPayPalAttachmentWriter).WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(t.Context(), bytes.NewReader(data), key)
			require.NoError(t, err)
			var saved outbox.Event
			require.NoError(t, db.First(&saved, event.ID).Error)
			require.ErrorIs(t, handler.Handle(t.Context(), saved), repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable)
			check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), event.AvailableAt.Add(time.Hour), 100)
			require.NoError(t, err)
			require.Equal(t, 1, check.Reconciled)
			require.Error(t, provider.StorageService.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
			require.NoError(t, db.First(&saved, event.ID).Error)
			require.Equal(t, "processed", saved.Status)
			require.Equal(t, 1, saved.Attempts)
		})
	}
}

func TestOriginalStorageCleanupReconciliationDiscoversLatePrivateDerivativeAfterParentRegistrationRemovalAndProtectsNewClaim(t *testing.T) {
	for _, retained := range []bool{false, true} {
		t.Run(fmt.Sprint(retained), func(t *testing.T) {
			env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
			ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
			result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Retired)
			worker := NewOutboxService(repository.NewOutboxRepository(env.db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle)
			_, err = worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 100)
			require.NoError(t, err)
			require.ErrorIs(t, env.db.Unscoped().First(&media.MediaAsset{}, asset.ID).Error, gorm.ErrRecordNotFound)
			key := fmt.Sprintf("%s/%d/card/late-after-parent-removal.webp", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
			objectPath := filepath.Join(env.privateRoot, filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o700))
			require.NoError(t, os.WriteFile(objectPath, []byte("late private derivative"), 0o600))
			if retained {
				require.NoError(t, env.db.Create(&media.MediaAsset{Filename: "retained derivative", URL: "/uploads/" + key, StorageKey: key, Visibility: "private"}).Error)
			}
			check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(2*time.Hour), 100)
			require.NoError(t, err)
			_, objectErr := os.Stat(objectPath)
			if retained {
				require.Equal(t, 1, check.Deferred)
				require.NoError(t, objectErr)
			} else {
				require.Equal(t, 1, check.Reconciled)
				require.ErrorIs(t, objectErr, os.ErrNotExist)
			}
		})
	}
}

func TestOriginalStorageCleanupReconciliationNeverEnrollsSuccessfullyBoundDedicatedUploads(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			require.NoError(t, env.upload(t.Context()))
			key := env.provider.writtenKeys[0]
			worker := NewOutboxService(repository.NewOutboxRepository(env.db))
			worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, env.handle)
			worker.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, env.handle)
			check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now().Add(365*24*time.Hour), 100)
			require.NoError(t, err)
			require.Zero(t, check.Claimed)
			require.Equal(t, key, env.currentKey())
			require.NoError(t, env.provider.StorageService.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
		})
	}
}

func TestOriginalStorageCleanupReconciliationHeartbeatsItsIndependentLeaseDuringProviderIO(t *testing.T) {
	db, worker := newTestOutboxService(t)
	worker.ConfigureLockTimeout(150 * time.Millisecond)
	event := outbox.Event{EventKey: "heartbeat-original", EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "media_asset", AggregateID: "1", Status: "processed", Attempts: 1}
	require.NoError(t, db.Create(&event).Error)
	worker.RegisterHandler(event.EventType, func(ctx context.Context, claimed outbox.Event) error {
		require.Eventually(t, func() bool {
			var stored outbox.Event
			if db.First(&stored, event.ID).Error != nil || stored.LockedAt == nil {
				return false
			}
			return stored.LockedBy == claimed.LockedBy && stored.LockedAt.After(*claimed.LockedAt)
		}, time.Second, 5*time.Millisecond)
		return ctx.Err()
	})
	check, err := worker.ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(t.Context(), time.Now(), 100)
	require.NoError(t, err)
	require.Equal(t, 1, check.Reconciled)
}
