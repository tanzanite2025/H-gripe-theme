package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
)

func syntheticCustomerServiceAgentVideoMP4HeaderBytes() []byte {
	return []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm', 'm', 'p', '4', '2'}
}

type customerServiceAgentPrivateVideoPreallocatedWriteObserver struct {
	storage.StorageService
	beforeWrite func(string)
	writeCount  int
}

func (s *customerServiceAgentPrivateVideoPreallocatedWriteObserver) VerifyStoredObjectExists(ctx context.Context, key string) error {
	return s.StorageService.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(ctx, key)
}

func (s *customerServiceAgentPrivateVideoPreallocatedWriteObserver) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.writeCount++
	if s.beforeWrite != nil {
		s.beforeWrite(key)
	}
	return s.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, reader, key)
}

func TestCustomerServiceAgentPrivateVideoPersistsReceiptBeforeWriteAndBindsWithoutPublicURLsOrImageDerivatives(t *testing.T) {
	env, uploads, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	payload := syntheticCustomerServiceAgentVideoMP4HeaderBytes()
	store := &customerServiceAgentPrivateVideoPreallocatedWriteObserver{StorageService: env.mediaService.storage}
	store.beforeWrite = func(key string) {
		var receipt media.MediaUploadSession
		require.NoError(t, env.db.First(&receipt, "object_key = ?", key).Error)
		require.Equal(t, "uploading", receipt.Status)
		require.Nil(t, receipt.MediaAssetID)
		require.Equal(t, int64(len(payload)), *receipt.ExpectedUploadBytes)
		require.Equal(t, env.conversation.ID, receipt.BusinessContextID)
		require.Equal(t, media.MediaUploadActorAdministrator, receipt.UploadActorType)
		require.Equal(t, env.conversation.AssignedTo, receipt.UploadActorAccountID)
		_, err := os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(key)))
		require.True(t, os.IsNotExist(err), "receipt must commit before the first provider write")
	}
	env.mediaService.storage = store
	asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false,
		multipartFileHeader(t, "agent-private-video.mp4", "image/png", payload))
	require.NoError(t, err)
	require.Equal(t, 1, store.writeCount)
	require.Equal(t, "video", asset.MediaType)
	require.Equal(t, "video/mp4", asset.MimeType)
	require.Equal(t, "private", asset.Visibility)
	require.Zero(t, asset.Width)
	require.Zero(t, asset.Height)
	require.Empty(t, asset.Derivatives)
	require.Empty(t, asset.AccessURL)
	require.Equal(t, fmt.Sprintf("%x", sha256.Sum256(payload)), asset.ContentSHA256)
	require.Equal(t, media.MediaLibraryPartitionCustomerService, asset.LibraryPartition)
	require.Equal(t, media.CustomerServiceMediaPurposeConversationAttachment, asset.CustomerServiceMediaPurpose)
	stored, err := os.ReadFile(filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey)))
	require.NoError(t, err)
	require.Equal(t, payload, stored)
	_, err = os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(asset.StorageKey)))
	require.True(t, os.IsNotExist(err))
	// Restore full local read capabilities after observing the write boundary.
	env.mediaService.storage = env.storage.StorageService
	allowed, err := NewPublicUploadAccessService(env.mediaService, nil).CanServePublicUpload(t.Context(), asset.StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
	var receipt media.MediaUploadSession
	require.NoError(t, env.db.First(&receipt, "id = ?", asset.PendingCustomerServiceUploadSessionID).Error)
	require.Equal(t, "ready", receipt.Status)
	require.Equal(t, asset.ID, *receipt.MediaAssetID)
	preview, err := env.access.OpenAgentCustomerServicePendingUploadPreview(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false, receipt.ID, "")
	require.NoError(t, err)
	previewBytes, err := io.ReadAll(preview.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, preview.ReadCloser.Close())
	require.Equal(t, payload, previewBytes)
	_, err = env.access.OpenPublicCustomerServicePendingUploadPreview(t.Context(), *env.conversation.ConversationID, env.owner, receipt.ID, "")
	require.Error(t, err)
	require.ErrorIs(t, env.mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	message := ticket.TicketMessage{TicketID: env.conversation.ID, Content: "Private video", MessageType: "video",
		RequestedMediaAttachmentBindings: []ticket.CustomerServiceMessageMediaAttachmentBindingRequest{{UploadSessionID: receipt.ID}}}
	require.NoError(t, env.access.tickets.AddCustomerServiceAgentMessage(&message, env.conversation.AssignedTo, false))
	require.Equal(t, ticket.TicketMessageMediaAssetBindings, message.AttachmentReferenceFormat)
	require.Empty(t, message.Attachments)
	require.NoError(t, env.db.First(&receipt, "id = ?", receipt.ID).Error)
	require.Equal(t, "bound", receipt.Status)
	_, err = env.access.OpenAgentCustomerServicePendingUploadPreview(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false, receipt.ID, "")
	require.Error(t, err)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, message.ID, 0, "")
	require.NoError(t, err)
	boundBytes, err := io.ReadAll(file.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	require.Equal(t, payload, boundBytes)
	references, err := env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, references, 2)
	require.ErrorIs(t, env.mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
}

func TestCustomerServiceAgentPrivateVideoRejectsInvalidUploadsBeforeReceiptOrObjectCreation(t *testing.T) {
	for _, scenario := range []string{"foreign support", "zero manager identity", "quota exceeded", "unsupported extension", "image disguised as video", "oversized video", "visitor video", "member video"} {
		t.Run(scenario, func(t *testing.T) {
			env, uploads, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			store := &customerServiceAgentPrivateVideoPreallocatedWriteObserver{StorageService: env.mediaService.storage}
			env.mediaService.storage = store
			payload := syntheticCustomerServiceAgentVideoMP4HeaderBytes()
			file := multipartFileHeader(t, "video.mp4", "video/mp4", payload)
			agentID, canViewAll := env.conversation.AssignedTo, false
			switch scenario {
			case "foreign support":
				agentID++
			case "zero manager identity":
				agentID, canViewAll = 0, true
			case "quota exceeded":
				env.mediaService.accountStorageQuotaBytes = int64(len(payload) - 1)
			case "unsupported extension":
				file.Filename = "video.exe"
			case "image disguised as video":
				file = multipartFileHeader(t, "video.mp4", "video/mp4", testOpaquePNG(t, 32, 16))
			case "oversized video":
				file.Size = upload.ProductVideoRule.MaxSize + 1
			}
			var beforeReceipts, beforeAssets int64
			require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Count(&beforeReceipts).Error)
			require.NoError(t, env.db.Model(&media.MediaAsset{}).Count(&beforeAssets).Error)
			var asset *media.MediaAsset
			var err error
			if scenario == "visitor video" || scenario == "member video" {
				owner := env.owner
				if scenario == "member video" {
					memberID := env.conversation.AssignedTo
					require.NoError(t, env.db.Model(&env.conversation).Updates(map[string]any{"customer_user_id": memberID, "visitor_session_hash": ""}).Error)
					owner = CustomerServiceOwner{UserID: &memberID}
				}
				asset, err = uploads.UploadVerifiedCustomerServiceConversationAttachment(t.Context(), *env.conversation.ConversationID, owner, file)
			} else {
				asset, err = uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(t.Context(), env.conversation.ID, agentID, canViewAll, file)
			}
			require.Error(t, err)
			require.Nil(t, asset)
			require.Zero(t, store.writeCount)
			var afterReceipts, afterAssets int64
			require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Count(&afterReceipts).Error)
			require.NoError(t, env.db.Model(&media.MediaAsset{}).Count(&afterAssets).Error)
			require.Equal(t, beforeReceipts, afterReceipts)
			require.Equal(t, beforeAssets, afterAssets)
		})
	}
}

func TestCustomerServiceAgentPrivateVideoFailedWritesAndRegistrationPreserveOriginalReceiptForExpiryCleanup(t *testing.T) {
	for _, scenario := range []string{"lost storage response", "registration failure"} {
		t.Run(scenario, func(t *testing.T) {
			env, uploads, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			if scenario == "lost storage response" {
				env.mediaService.storage = &customerServicePendingUploadStorageThatWritesObjectThenFails{StorageService: env.storage.StorageService}
			} else {
				callback := "reject_private_video_asset_registration"
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_assets" {
						tx.AddError(errors.New("synthetic registration failure"))
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			}
			payload := syntheticCustomerServiceAgentVideoMP4HeaderBytes()
			asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false,
				multipartFileHeader(t, "failed-agent-video.mp4", "video/mp4", payload))
			require.Error(t, err)
			require.Nil(t, asset)
			var receipt media.MediaUploadSession
			require.NoError(t, env.db.First(&receipt, "status = ?", "failed").Error)
			require.Nil(t, receipt.MediaAssetID)
			require.Equal(t, media.MediaUploadActorAdministrator, receipt.UploadActorType)
			require.Equal(t, env.conversation.AssignedTo, receipt.UploadActorAccountID)
			require.Equal(t, int64(len(payload)), *receipt.ExpectedUploadBytes)
			stored, err := os.ReadFile(filepath.Join(env.privateRoot, filepath.FromSlash(receipt.ObjectKey)))
			require.NoError(t, err)
			require.Equal(t, payload, stored)
			env.mediaService.storage = env.storage.StorageService
			ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, receipt.ID)
			retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
			result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Retired)
			var event outbox.Event
			require.NoError(t, env.db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, strconv.FormatUint(uint64(env.conversation.ID), 10)).First(&event).Error)
			handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event))
			_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(receipt.ObjectKey)))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestCustomerServiceAgentPrivateVideoPendingExpiryKeepsQuotaUntilOriginalCleanupSucceeds(t *testing.T) {
	env, uploads, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false,
		multipartFileHeader(t, "expired-agent-video.mp4", "video/mp4", syntheticCustomerServiceAgentVideoMP4HeaderBytes()))
	require.NoError(t, err)
	identity := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: env.conversation.AssignedTo}
	before, err := env.mediaService.repo.AssetStorageUsageByUploadActor(identity)
	require.NoError(t, err)
	ageCustomerServiceUploadSessionBeyondWriteAndReceiptDeadlines(t, env, asset.PendingCustomerServiceUploadSessionID)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	afterRetirement, err := env.mediaService.repo.AssetStorageUsageByUploadActor(identity)
	require.NoError(t, err)
	require.Equal(t, before, afterRetirement)
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
	afterCleanup, err := env.mediaService.repo.AssetStorageUsageByUploadActor(identity)
	require.NoError(t, err)
	require.Equal(t, before-asset.Size, afterCleanup)
}
