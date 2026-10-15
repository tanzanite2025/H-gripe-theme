package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func configureCustomerServiceManagedAttachmentRetirementTestService(t *testing.T, env customerServicePrivateAttachmentAccessTestEnvironment) *CustomerServiceRetentionService {
	t.Helper()
	if !env.db.Migrator().HasTable(&audit.AuditLog{}) {
		require.NoError(t, env.db.AutoMigrate(&audit.AuditLog{}))
	}
	retention := NewCustomerServiceRetentionService(env.access.tickets.ticketRepo, nil, nil, nil, nil, NewAuditService(repository.NewAuditRepository(env.db)))
	retention.ConfigureMediaService(env.mediaService)
	retention.ConfigureAttachmentStorage(env.mediaService.storage)
	retention.ConfigureCleanupOutbox(repository.NewOutboxRepository(env.db))
	return retention
}

func ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t *testing.T, env customerServicePrivateAttachmentAccessTestEnvironment, receiptID string) {
	t.Helper()
	require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", receiptID).
		Updates(map[string]interface{}{"created_at": time.Now().Add(-48 * time.Hour), "expires_at": time.Now().Add(-24 * time.Hour)}).Error)
}

type customerServiceAttachmentCleanupStorageWithInjectedDeleteFailure struct {
	storage.StorageService
	fail bool
}

func (s *customerServiceAttachmentCleanupStorageWithInjectedDeleteFailure) Delete(ctx context.Context, key string) error {
	if s.fail {
		return errors.New("injected private object cleanup failure")
	}
	return s.StorageService.Delete(ctx, key)
}

func (s *customerServiceAttachmentCleanupStorageWithInjectedDeleteFailure) ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx context.Context, assetID uint) ([]string, error) {
	return s.StorageService.(storage.CustomerServiceAttachmentDerivativeNamespaceCleanupLister).ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx, assetID)
}

func TestCustomerServiceExpiredPendingUploadRetiresAtomicallyThenReplaysOriginalDerivativeAndOrphanCleanup(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
	orphanKey := fmt.Sprintf("%s/%d/thumbnail/v1/orphaned-unregistered.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
	_, err := env.mediaService.storage.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("orphan")), orphanKey)
	require.NoError(t, err)
	identity := media.MediaUploadActorIdentity{ActorType: asset.UploaderActorType, VisitorSessionHash: asset.UploaderVisitorSessionHash}
	beforeUsage, err := env.mediaService.repo.AssetStorageUsageByUploadActor(identity)
	require.NoError(t, err)
	result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	require.Zero(t, result.Errors)
	_, err = env.mediaService.repo.FindCustomerServiceMediaUploadSession(t.Context(), asset.PendingCustomerServiceUploadSessionID)
	require.Error(t, err)
	tombstone, err := env.mediaService.repo.FindAssetByIDUnscoped(asset.ID)
	require.NoError(t, err)
	require.True(t, tombstone.DeletedAt.Valid)
	usageBeforeFilesDeleted, err := env.mediaService.repo.AssetStorageUsageByUploadActor(identity)
	require.NoError(t, err)
	require.Equal(t, beforeUsage, usageBeforeFilesDeleted)
	var event outbox.Event
	require.NoError(t, env.db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, strconv.FormatUint(uint64(asset.ID), 10)).First(&event).Error)
	store := &customerServiceAttachmentCleanupStorageWithInjectedDeleteFailure{StorageService: env.mediaService.storage, fail: true}
	env.mediaService.storage = store
	handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
	require.Error(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey)))
	require.NoError(t, err)
	store.fail = false
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	for _, key := range append([]string{asset.StorageKey, orphanKey}, asset.Derivatives[0].StorageKey, asset.Derivatives[1].StorageKey, asset.Derivatives[2].StorageKey) {
		_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(key)))
		require.True(t, os.IsNotExist(err), key)
	}
	// The neighboring asset namespace and original remain untouched.
	_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(env.asset.StorageKey)))
	require.NoError(t, err)
	var derivatives int64
	require.NoError(t, env.db.Unscoped().Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ?", asset.ID).Count(&derivatives).Error)
	require.Zero(t, derivatives)
	afterUsage, err := env.mediaService.repo.AssetStorageUsageByUploadActor(identity)
	require.NoError(t, err)
	require.Equal(t, beforeUsage-asset.Size, afterUsage)
	result, err = retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Zero(t, result.Scanned)
}

func TestCustomerServiceExpiredFailedAndAbandonedUnregisteredUploadsKeepExactCleanupKeysAndProtectNewClaims(t *testing.T) {
	for _, status := range []string{"failed", "uploading"} {
		t.Run(status, func(t *testing.T) {
			env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
			key, err := storage.AllocatePrivateObjectKeyForPendingUpload("failed.png", storage.CustomerServiceConversationAttachmentOriginalsPrefix)
			require.NoError(t, err)
			_, err = env.mediaService.storage.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("partial")), key)
			require.NoError(t, err)
			session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
				BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
				ObjectKey: key, Status: status, CreatedAt: time.Now().Add(-48 * time.Hour), ExpiresAt: time.Now().Add(-24 * time.Hour)}
			require.NoError(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), &session))
			result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Retired)
			var event outbox.Event
			require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
			handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
			// A delayed event must not delete an object newly claimed by a receipt.
			claim := session
			claim.ID, claim.Status, claim.ExpiresAt = uuid.NewString(), "uploading", time.Now().Add(time.Hour)
			require.NoError(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), &claim))
			require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
			require.NoError(t, env.db.Delete(&claim).Error)
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event))
			_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestCustomerServiceUploadExpiryWaitsForBoundedOriginalWritesAndPrivateDerivativeLease(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", asset.PendingCustomerServiceUploadSessionID).Update("expires_at", time.Now().Add(-time.Minute)).Error)
	result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Zero(t, result.Scanned, "recent writers must never be selected")
	ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
	require.NoError(t, env.db.Model(asset).Updates(map[string]interface{}{"customer_service_derivative_write_token": uuid.NewString(), "customer_service_derivative_write_expires_at": time.Now().Add(time.Minute)}).Error)
	result, err = retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Blocked)
	require.Zero(t, result.Retired)
	require.NoError(t, env.db.Model(asset).Update("customer_service_derivative_write_expires_at", time.Now().Add(-time.Minute)).Error)
	result, err = retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
}

func TestCustomerServiceManagedConversationPurgeReleasesBindingsAndReceiptsButPreservesPublicProductAssets(t *testing.T) {
	env, _, privateAsset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	_, _, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Private photo", 0, "image", "", "",
		ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: privateAsset.PendingCustomerServiceUploadSessionID})
	require.NoError(t, err)
	publicAsset, err := env.mediaService.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "product.png", "image/png", testOpaquePNG(t, 32, 16)), MediaType: "image",
		Width: 32, Height: 16, UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: env.conversation.AssignedTo}})
	require.NoError(t, err)
	_, _, err = env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Public product photo", 0, "image", "", "",
		ticket.CustomerServiceMessageMediaAttachmentBindingRequest{MediaAssetID: publicAsset.ID})
	require.NoError(t, err)
	require.NoError(t, env.db.AutoMigrate(&product.Product{}, &product.ProductMedia{}))
	item := product.Product{SKU: "retention-shared-product", Slug: "retention-shared-product", Name: "Shared product", PriceMinor: 100}
	require.NoError(t, env.db.Create(&item).Error)
	require.NoError(t, env.db.Create(&product.ProductMedia{ProductID: item.ID, MediaAssetID: &publicAsset.ID, URL: publicAsset.URL, MediaType: "image", Role: "primary"}).Error)
	ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, privateAsset.PendingCustomerServiceUploadSessionID)
	deletedAt := time.Now().Add(-31 * 24 * time.Hour)
	require.NoError(t, env.db.Model(&env.conversation).Updates(map[string]interface{}{"status": "closed", "deleted_at": deletedAt}).Error)
	eligibility, err := retention.Evaluate([]uint{env.conversation.ID})
	require.NoError(t, err)
	require.True(t, eligibility[0].Eligible)
	_, err = retention.Purge([]uint{env.conversation.ID}, env.conversation.AssignedTo, "validate managed lifecycle")
	require.NoError(t, err)
	for _, model := range []interface{}{&ticket.TicketMessage{}, &ticket.TicketMessageMediaAttachment{}, &media.MediaUploadSession{}} {
		var count int64
		require.NoError(t, env.db.Model(model).Count(&count).Error)
		require.Zero(t, count)
	}
	var cleanupEvents []outbox.Event
	require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Find(&cleanupEvents).Error)
	require.Len(t, cleanupEvents, 2, "both private originals have asset cleanup events")
	handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
	for _, event := range cleanupEvents {
		require.NoError(t, handler.Handle(t.Context(), event))
		require.NoError(t, handler.Handle(t.Context(), event))
	}
	publicRecord, err := env.mediaService.repo.FindAssetByID(publicAsset.ID)
	require.NoError(t, err)
	require.False(t, publicRecord.DeletedAt.Valid)
	_, err = os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(publicAsset.StorageKey)))
	require.NoError(t, err)
	for _, derivative := range publicAsset.Derivatives {
		_, err = os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(derivative.StorageKey)))
		require.NoError(t, err)
	}
	var projectionCleanup outbox.Event
	require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeCustomerServiceRetentionCleanup).First(&projectionCleanup).Error)
	require.NoError(t, retention.HandleCustomerServiceRetentionCleanup(t.Context(), projectionCleanup))
	require.NoError(t, retention.HandleCustomerServiceRetentionCleanup(t.Context(), projectionCleanup))
}

func TestCustomerServiceAttachmentExpiryAndConversationPurgeRollBackWhenCleanupOutboxCannotPersist(t *testing.T) {
	for _, operation := range []string{"expiry", "conversation purge"} {
		t.Run(operation, func(t *testing.T) {
			env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
			if operation == "conversation purge" {
				_, _, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Photo", 0, "image", "", "",
					ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID})
				require.NoError(t, err)
				require.NoError(t, env.db.Model(&env.conversation).Update("deleted_at", time.Now().Add(-31*24*time.Hour)).Error)
			}
			ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
			require.NoError(t, env.db.Migrator().DropTable(&outbox.Event{}))
			if operation == "expiry" {
				result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Errors)
			} else {
				_, err := retention.Purge([]uint{env.conversation.ID}, env.conversation.AssignedTo, "outbox failure")
				require.Error(t, err)
				var bindings int64
				require.NoError(t, env.db.Model(&ticket.TicketMessageMediaAttachment{}).Count(&bindings).Error)
				require.EqualValues(t, 1, bindings)
			}
			_, err := env.mediaService.repo.FindCustomerServiceMediaUploadSession(t.Context(), asset.PendingCustomerServiceUploadSessionID)
			require.NoError(t, err)
			current, err := env.mediaService.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			require.False(t, current.DeletedAt.Valid)
			_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey)))
			require.NoError(t, err)
		})
	}
}

func TestCustomerServiceManagedConversationPurgePreservesPrivateAssetsReferencedByOtherMessagesOrBusinesses(t *testing.T) {
	for _, survivingReference := range []string{"bound message", "legacy message", "product media"} {
		t.Run(survivingReference, func(t *testing.T) {
			env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
			_, _, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Photo", 0, "image", "", "",
				ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID})
			require.NoError(t, err)
			if survivingReference == "product media" {
				require.NoError(t, env.db.AutoMigrate(&product.Product{}, &product.ProductMedia{}))
				item := product.Product{SKU: "private-reference-safety", Slug: "private-reference-safety", Name: "Reference safety", PriceMinor: 100}
				require.NoError(t, env.db.Create(&item).Error)
				// A pre-existing cross-domain reference must remain a deletion hold,
				// even if that domain cannot serve the private object publicly.
				require.NoError(t, env.db.Create(&product.ProductMedia{ProductID: item.ID, MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image", Role: "primary"}).Error)
			} else {
				conversationID := uuid.NewString()
				otherConversation := ticket.Ticket{TicketNumber: "shared-private-" + uuid.NewString(), ConversationID: &conversationID,
					UserID: env.conversation.UserID, Category: "customer_service", Subject: "Surviving conversation", Status: "in_progress"}
				require.NoError(t, env.db.Create(&otherConversation).Error)
				message := ticket.TicketMessage{TicketID: otherConversation.ID, Content: "Surviving photo", MessageType: "image"}
				if survivingReference == "legacy message" {
					attachmentJSON, err := json.Marshal([]string{asset.URL})
					require.NoError(t, err)
					message.Attachments = string(attachmentJSON)
				} else {
					message.AttachmentReferenceFormat = ticket.TicketMessageMediaAssetBindings
				}
				require.NoError(t, env.db.Create(&message).Error)
				if survivingReference == "bound message" {
					require.NoError(t, env.db.Create(&ticket.TicketMessageMediaAttachment{TicketMessageID: message.ID, MediaAssetID: asset.ID}).Error)
				}
			}
			ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
			require.NoError(t, env.db.Model(&env.conversation).Update("deleted_at", time.Now().Add(-31*24*time.Hour)).Error)
			_, err = retention.Purge([]uint{env.conversation.ID}, env.conversation.AssignedTo, "shared private reference safety")
			require.NoError(t, err)
			current, err := env.mediaService.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			references, err := env.mediaService.repo.FindAssetReferences(current)
			require.NoError(t, err)
			require.Len(t, references, 1)
			var cleanupEvents []outbox.Event
			require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Find(&cleanupEvents).Error)
			handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
			for _, event := range cleanupEvents {
				require.NotEqual(t, strconv.FormatUint(uint64(asset.ID), 10), event.AggregateID)
				require.NoError(t, handler.Handle(t.Context(), event))
			}
			var projectionCleanup outbox.Event
			require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeCustomerServiceRetentionCleanup).First(&projectionCleanup).Error)
			require.NoError(t, retention.HandleCustomerServiceRetentionCleanup(t.Context(), projectionCleanup))
			for _, key := range append([]string{asset.StorageKey}, asset.Derivatives[0].StorageKey, asset.Derivatives[1].StorageKey, asset.Derivatives[2].StorageKey) {
				_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(key)))
				require.NoError(t, err, key)
			}
		})
	}
}

func TestCustomerServiceRetiredAssetCleanupRejectsForeignDerivativeMetadataBeforeDeletingAnyObject(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
	derivative := asset.Derivatives[0]
	require.NoError(t, env.db.Model(&derivative).Update("storage_key", env.asset.Derivatives[0].StorageKey).Error)
	result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	var event outbox.Event
	require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
	handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
	require.ErrorContains(t, handler.Handle(t.Context(), event), "does not belong to its asset")
	for _, key := range append([]string{asset.StorageKey, env.asset.Derivatives[0].StorageKey}, asset.Derivatives[0].StorageKey, asset.Derivatives[1].StorageKey, asset.Derivatives[2].StorageKey) {
		_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(key)))
		require.NoError(t, err, key)
	}
	tombstone, err := env.mediaService.repo.FindAssetByIDUnscoped(asset.ID)
	require.NoError(t, err)
	require.True(t, tombstone.DeletedAt.Valid)
	// Correcting metadata allows the same durable event to finish safely.
	require.NoError(t, env.db.Model(&derivative).Update("storage_key", asset.Derivatives[0].StorageKey).Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(env.asset.Derivatives[0].StorageKey)))
	require.NoError(t, err)
}
