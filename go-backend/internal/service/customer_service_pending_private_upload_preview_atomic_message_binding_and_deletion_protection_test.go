package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func createCustomerServicePendingPrivateUploadBindingTestEnvironment(t *testing.T) (customerServicePrivateAttachmentAccessTestEnvironment, *CustomerServiceConversationAttachmentUploadService, *media.MediaAsset) {
	t.Helper()
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&media.MediaUploadSession{}, &ticket.TicketMessageMediaAttachment{}))
	env.mediaService.storage = env.storage.StorageService
	env.mediaService.ConfigureCustomerServicePrivateAttachmentUploads(true)
	env.mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
	uploads := NewCustomerServiceConversationAttachmentUploadService(env.access.tickets, env.mediaService)
	asset, err := uploads.UploadVerifiedCustomerServiceConversationAttachment(t.Context(), *env.conversation.ConversationID, env.owner,
		multipartFileHeader(t, "pending-customer-photo.png", "image/png", testOpaquePNG(t, 64, 32)))
	require.NoError(t, err)
	require.NotEmpty(t, asset.PendingCustomerServiceUploadSessionID)
	return env, uploads, asset
}

func TestCustomerServicePendingPrivateUploadPreviewBecomesAtomicAssetBindingWithoutDuplicatingStoredURLs(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	var session media.MediaUploadSession
	require.NoError(t, env.db.First(&session, "id = ?", asset.PendingCustomerServiceUploadSessionID).Error)
	require.Equal(t, "ready", session.Status)
	require.Equal(t, asset.StorageKey, session.ObjectKey)
	require.Equal(t, asset.ID, *session.MediaAssetID)
	require.Equal(t, "private", asset.Visibility)
	require.Len(t, asset.Derivatives, 3)
	_, err := os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(asset.StorageKey)))
	require.True(t, os.IsNotExist(err))
	file, err := env.access.OpenPublicCustomerServicePendingUploadPreview(t.Context(), *env.conversation.ConversationID, env.owner, session.ID, "thumbnail")
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	beforeReferences, err := env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, beforeReferences, 1)
	require.Equal(t, "media_upload_session", beforeReferences[0].ResourceType)
	require.ErrorIs(t, env.mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)

	previewURL := env.access.PublicCustomerServicePendingUploadPreviewURL(*env.conversation.ConversationID, session.ID)
	bindings, err := env.access.PrepareCustomerServiceMediaAttachmentBindingRequests([]string{previewURL, previewURL}, nil, *env.conversation.ConversationID, 0, 1)
	require.NoError(t, err)
	require.Len(t, bindings, 1)
	_, message, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Customer image", 0, "image",
		`{"url":"private-preview-url","title":"Customer photo"}`, "", bindings...)
	require.NoError(t, err)
	require.Equal(t, ticket.TicketMessageMediaAssetBindings, message.AttachmentReferenceFormat)
	require.Empty(t, message.Attachments)
	require.NotContains(t, message.Metadata, "private-preview-url")
	require.Contains(t, message.Metadata, "Customer photo")
	require.NoError(t, env.db.First(&session, "id = ?", session.ID).Error)
	require.Equal(t, "bound", session.Status)
	file, err = env.access.OpenPublicCustomerServicePendingUploadPreview(t.Context(), *env.conversation.ConversationID, env.owner, session.ID, "")
	require.ErrorIs(t, err, ErrCustomerServiceMessageAttachmentUnavailable)
	require.Nil(t, file)
	for _, preset := range []string{"", "thumbnail", "card", "large"} {
		file, err = env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, message.ID, 0, preset)
		require.NoError(t, err)
		content, readErr := io.ReadAll(file.ReadCloser)
		require.NoError(t, readErr)
		require.NotEmpty(t, content)
		require.NoError(t, file.ReadCloser.Close())
	}
	// A stale compatibility array cannot replace the authoritative binding.
	require.NoError(t, env.db.Model(message).Update("attachments", env.message.Attachments).Error)
	message.Attachments = env.message.Attachments
	urls, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(*message, *env.conversation.ConversationID)
	require.Len(t, descriptions, 1)
	require.Equal(t, asset.ID, descriptions[0].AssetID)
	require.Equal(t, []string{descriptions[0].AccessURL}, urls)
	require.NotContains(t, urls[0], asset.StorageKey)
	afterReferences, err := env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, afterReferences, 2)
	require.Equal(t, "ticket_message", afterReferences[0].ResourceType)
	require.Equal(t, message.ID, afterReferences[0].ResourceID)
	require.Equal(t, "media_upload_session", afterReferences[1].ResourceType)
	require.Equal(t, "retained_upload_receipt", afterReferences[1].UsageState)
	require.Equal(t, asset.ID, afterReferences[1].ResourceID)
	require.ErrorIs(t, env.mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	held, err := env.access.tickets.ticketRepo.HasCustomerServiceConversationManagedAttachmentLifecycleHold(env.conversation.ID)
	require.NoError(t, err)
	require.True(t, held)
}

func TestCustomerServicePendingUploadBindingsRejectForeignExpiredReusedAndInvalidAssetsWithoutMessageOrOutboxWrites(t *testing.T) {
	for _, scenario := range []string{"other conversation", "expired", "failed", "already bound", "inactive asset", "deleted asset", "session key mismatch", "session owner mismatch", "bare private asset ID", "duplicate receipt", "missing receipt"} {
		t.Run(scenario, func(t *testing.T) {
			env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			request := ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID}
			requests := []ticket.CustomerServiceMessageMediaAttachmentBindingRequest{request}
			switch scenario {
			case "other conversation":
				require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", request.UploadSessionID).Update("business_context_id", env.conversation.ID+100).Error)
			case "expired":
				require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", request.UploadSessionID).Update("expires_at", time.Now().Add(-time.Minute)).Error)
			case "failed", "already bound":
				status := "failed"
				if scenario == "already bound" {
					status = "bound"
				}
				require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", request.UploadSessionID).Update("status", status).Error)
			case "inactive asset":
				require.NoError(t, env.db.Model(asset).Update("status", "inactive").Error)
			case "deleted asset":
				require.NoError(t, env.db.Delete(asset).Error)
			case "session key mismatch":
				require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", request.UploadSessionID).Update("object_key", "customer-service/attachments/originals/foreign.png").Error)
			case "session owner mismatch":
				require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", request.UploadSessionID).Update("upload_actor_visitor_session_hash", strings.Repeat("b", 64)).Error)
			case "bare private asset ID":
				requests = []ticket.CustomerServiceMessageMediaAttachmentBindingRequest{{MediaAssetID: asset.ID}}
			case "duplicate receipt":
				requests = append(requests, request)
			case "missing receipt":
				requests[0].UploadSessionID = uuid.NewString()
			}
			var beforeMessages, beforeEvents int64
			require.NoError(t, env.db.Model(&ticket.TicketMessage{}).Count(&beforeMessages).Error)
			require.NoError(t, env.db.Model(&outbox.Event{}).Count(&beforeEvents).Error)
			_, message, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Photo", 0, "image", "", "", requests...)
			require.ErrorIs(t, err, repository.ErrCustomerServiceAttachmentBindingUnavailable)
			require.Nil(t, message)
			var afterMessages, afterEvents, bindingCount int64
			require.NoError(t, env.db.Model(&ticket.TicketMessage{}).Count(&afterMessages).Error)
			require.NoError(t, env.db.Model(&outbox.Event{}).Count(&afterEvents).Error)
			require.NoError(t, env.db.Model(&ticket.TicketMessageMediaAttachment{}).Count(&bindingCount).Error)
			require.Equal(t, beforeMessages, afterMessages)
			require.Equal(t, beforeEvents, afterEvents)
			require.Zero(t, bindingCount)
		})
	}
}

func TestCustomerServicePendingUploadBindingAndReadyStateRollBackWhenRealtimeOutboxFails(t *testing.T) {
	env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	require.NoError(t, env.db.Migrator().DropTable(&outbox.Event{}))
	_, message, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Photo", 0, "image", "", "",
		ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID})
	require.Error(t, err)
	require.Nil(t, message)
	var session media.MediaUploadSession
	require.NoError(t, env.db.First(&session, "id = ?", asset.PendingCustomerServiceUploadSessionID).Error)
	require.Equal(t, "ready", session.Status)
	var count int64
	require.NoError(t, env.db.Model(&ticket.TicketMessageMediaAttachment{}).Count(&count).Error)
	require.Zero(t, count)
	file, err := env.access.OpenPublicCustomerServicePendingUploadPreview(t.Context(), *env.conversation.ConversationID, env.owner, session.ID, "")
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
}

func TestCustomerServiceAgentDraftPreviewRequiresUploaderAndCurrentAssignmentBeforeMessageBinding(t *testing.T) {
	env, uploads, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false,
		multipartFileHeader(t, "agent-photo.png", "image/png", testOpaquePNG(t, 32, 16)))
	require.NoError(t, err)
	file, err := env.access.OpenAgentCustomerServicePendingUploadPreview(t.Context(), env.conversation.ID, env.conversation.AssignedTo, false, asset.PendingCustomerServiceUploadSessionID, "")
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	file, err = env.access.OpenPublicCustomerServicePendingUploadPreview(t.Context(), *env.conversation.ConversationID, env.owner, asset.PendingCustomerServiceUploadSessionID, "")
	require.Error(t, err)
	require.Nil(t, file)
	agentID := env.conversation.AssignedTo
	require.NoError(t, env.db.Model(&env.conversation).Update("assigned_to", agentID+100).Error)
	file, err = env.access.OpenAgentCustomerServicePendingUploadPreview(t.Context(), env.conversation.ID, agentID, false, asset.PendingCustomerServiceUploadSessionID, "")
	require.ErrorIs(t, err, ErrCustomerServiceAgentAccessDenied)
	require.Nil(t, file)
	file, err = env.access.OpenAgentCustomerServicePendingUploadPreview(t.Context(), env.conversation.ID, agentID+100, false, asset.PendingCustomerServiceUploadSessionID, "")
	require.ErrorIs(t, err, ErrCustomerServiceMessageAttachmentUnavailable)
	require.Nil(t, file)
}

type customerServicePendingUploadStorageThatWritesObjectThenFails struct{ storage.StorageService }

func (s *customerServicePendingUploadStorageThatWritesObjectThenFails) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	_, err := s.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return "", err
	}
	return "", errors.New("simulated storage response failure after write")
}

func TestCustomerServiceFailedUploadKeepsItsPreallocatedObjectKeyAndOwnerForRecovery(t *testing.T) {
	env, uploads, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	env.mediaService.storage = &customerServicePendingUploadStorageThatWritesObjectThenFails{StorageService: env.storage.StorageService}
	var beforeAssets int64
	require.NoError(t, env.db.Model(&media.MediaAsset{}).Count(&beforeAssets).Error)
	asset, err := uploads.UploadVerifiedCustomerServiceConversationAttachment(t.Context(), *env.conversation.ConversationID, env.owner,
		multipartFileHeader(t, "failed-photo.png", "image/png", testOpaquePNG(t, 32, 16)))
	require.Error(t, err)
	require.Nil(t, asset)
	var failed media.MediaUploadSession
	require.NoError(t, env.db.First(&failed, "status = ?", "failed").Error)
	require.Nil(t, failed.MediaAssetID)
	require.Equal(t, env.owner.VisitorSessionHash, failed.UploadActorVisitorSessionHash)
	require.Equal(t, env.conversation.ID, failed.BusinessContextID)
	_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(failed.ObjectKey)))
	require.NoError(t, err)
	var afterAssets int64
	require.NoError(t, env.db.Model(&media.MediaAsset{}).Count(&afterAssets).Error)
	require.Equal(t, beforeAssets, afterAssets)
}

func TestCustomerServicePublicAssetReuseAddsOnlyMessageBindingsAndKeepsLegacyURLMessagesReadable(t *testing.T) {
	env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	asset, err := env.mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "public-product-photo.png", "image/png", testOpaquePNG(t, 32, 16)), MediaType: "image",
		UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: env.conversation.AssignedTo},
	})
	require.NoError(t, err)
	var beforeAssets, beforeDerivatives int64
	require.NoError(t, env.db.Model(&media.MediaAsset{}).Count(&beforeAssets).Error)
	require.NoError(t, env.db.Model(&media.MediaAssetDerivative{}).Count(&beforeDerivatives).Error)
	for range 2 {
		_, message, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Product image", 0, "image", "", "",
			ticket.CustomerServiceMessageMediaAttachmentBindingRequest{MediaAssetID: asset.ID})
		require.NoError(t, err)
		require.Empty(t, message.Attachments)
		file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, message.ID, 0, "")
		require.NoError(t, err)
		require.NoError(t, file.ReadCloser.Close())
		urls, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(*message, *env.conversation.ConversationID)
		require.Equal(t, []string{env.mediaService.CanonicalPublicMediaURL(asset.URL)}, urls)
		require.Len(t, descriptions, 1)
		require.Equal(t, asset.ID, descriptions[0].AssetID)
	}
	var afterAssets, afterDerivatives int64
	require.NoError(t, env.db.Model(&media.MediaAsset{}).Count(&afterAssets).Error)
	require.NoError(t, env.db.Model(&media.MediaAssetDerivative{}).Count(&afterDerivatives).Error)
	require.Equal(t, beforeAssets, afterAssets)
	require.Equal(t, beforeDerivatives, afterDerivatives)
	references, err := env.mediaService.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, references, 2)
	require.ErrorIs(t, env.mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	legacyAttachments, err := json.Marshal([]string{asset.URL})
	require.NoError(t, err)
	_, legacy, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Old client", 0, "image", "", string(legacyAttachments))
	require.NoError(t, err)
	require.Equal(t, ticket.TicketMessageLegacyURLAttachments, legacy.AttachmentReferenceFormat)
	require.Equal(t, string(legacyAttachments), legacy.Attachments)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, legacy.ID, 0, "")
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
}

func TestCustomerServicePendingAndBoundPrivateReadsRejectRegisteredAssetUploaderIdentityMismatch(t *testing.T) {
	for _, state := range []string{"pending", "bound"} {
		t.Run(state, func(t *testing.T) {
			env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			var message *ticket.TicketMessage
			if state == "bound" {
				var err error
				_, message, err = env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Photo", 0, "image", "", "",
					ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID})
				require.NoError(t, err)
			}
			require.NoError(t, env.db.Model(asset).Update("uploader_visitor_session_hash", strings.Repeat("b", 64)).Error)
			var file *MediaAssetFile
			var err error
			if state == "pending" {
				file, err = env.access.OpenPublicCustomerServicePendingUploadPreview(t.Context(), *env.conversation.ConversationID, env.owner, asset.PendingCustomerServiceUploadSessionID, "")
			} else {
				file, err = env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, message.ID, 0, "")
				urls, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(*message, *env.conversation.ConversationID)
				require.Equal(t, []string{""}, urls)
				require.Empty(t, descriptions)
			}
			require.ErrorIs(t, err, ErrCustomerServiceMessageAttachmentUnavailable)
			require.Nil(t, file)
		})
	}
}

func TestCustomerServiceManagedAttachmentRetentionHoldPreservesBoundPendingAndFailedUploadRecords(t *testing.T) {
	for _, state := range []string{"bound", "expired pending", "failed without asset"} {
		t.Run(state, func(t *testing.T) {
			env, _, asset := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			if state == "bound" {
				_, _, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Photo", 0, "image", "", "",
					ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID})
				require.NoError(t, err)
			} else {
				updates := map[string]interface{}{"expires_at": time.Now().Add(-time.Hour)}
				if state == "failed without asset" {
					updates["status"] = "failed"
					updates["media_asset_id"] = nil
				}
				require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", asset.PendingCustomerServiceUploadSessionID).Updates(updates).Error)
			}
			deletedAt := time.Now().Add(-90 * 24 * time.Hour)
			require.NoError(t, env.db.Model(&env.conversation).Updates(map[string]interface{}{"status": "closed", "deleted_at": deletedAt}).Error)
			retention := NewCustomerServiceRetentionService(env.access.tickets.ticketRepo, nil, nil, nil, nil, NewAuditService(repository.NewAuditRepository(env.db)))
			eligibility, err := retention.Evaluate([]uint{env.conversation.ID})
			require.NoError(t, err)
			require.Len(t, eligibility, 1)
			require.False(t, eligibility[0].Eligible)
			require.Equal(t, "attachment retention hold", eligibility[0].Reason)
			_, err = retention.Purge([]uint{env.conversation.ID}, env.conversation.AssignedTo, "verify managed attachment protection")
			require.ErrorIs(t, err, ErrCustomerServiceRetentionIneligible)
			require.ErrorIs(t, env.access.tickets.ticketRepo.PurgeCustomerServiceConversation(env.conversation.ID), repository.ErrCustomerServiceAttachmentBindingUnavailable)
			var conversationCount, sessionCount int64
			require.NoError(t, env.db.Unscoped().Model(&ticket.Ticket{}).Where("id = ?", env.conversation.ID).Count(&conversationCount).Error)
			require.NoError(t, env.db.Model(&media.MediaUploadSession{}).Where("id = ?", asset.PendingCustomerServiceUploadSessionID).Count(&sessionCount).Error)
			require.EqualValues(t, 1, conversationCount)
			require.EqualValues(t, 1, sessionCount)
			_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(asset.StorageKey)))
			require.NoError(t, err)
		})
	}
}
