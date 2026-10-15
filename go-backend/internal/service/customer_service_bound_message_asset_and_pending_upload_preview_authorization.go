package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
)

func (s *CustomerServiceConversationMessageAttachmentAccessService) resolveBoundCustomerServiceMessageAttachment(ctx context.Context, message ticket.TicketMessage, attachmentIndex int) (*media.MediaAsset, string, error) {
	if attachmentIndex < 0 {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	bindings, err := s.tickets.ticketRepo.FindCustomerServiceMessageMediaAssetBindings(message.ID)
	if err != nil {
		return nil, "", err
	}
	for _, binding := range bindings {
		if binding.AttachmentIndex != attachmentIndex {
			continue
		}
		return s.resolveCustomerServiceMessageMediaAssetBinding(ctx, message, binding)
	}
	return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) resolveCustomerServiceMessageMediaAssetBinding(ctx context.Context, message ticket.TicketMessage, binding ticket.TicketMessageMediaAttachment) (*media.MediaAsset, string, error) {
	if binding.TicketMessageID != message.ID {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	asset, err := s.media.repo.FindAssetByID(binding.MediaAssetID)
	if err != nil {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	asset, key, err := validateRegisteredCustomerServiceAttachmentAsset(asset, asset.StorageKey)
	if err != nil {
		return nil, "", err
	}
	if asset.Visibility == "private" {
		if binding.UploadSessionID == nil {
			return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
		}
		session, sessionErr := s.media.repo.FindCustomerServiceMediaUploadSession(ctx, *binding.UploadSessionID)
		if sessionErr != nil || session.Status != "bound" || session.BusinessContextType != media.MediaUploadSessionCustomerServiceContext ||
			session.BusinessContextID != message.TicketID || !customerServiceUploadSessionMatchesRegisteredAsset(session, asset, key) {
			return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
		}
	}
	return asset, key, nil
}

func customerServiceUploadSessionMatchesRegisteredAsset(session *media.MediaUploadSession, asset *media.MediaAsset, key string) bool {
	identity := media.MediaUploadActorIdentity{ActorType: asset.UploaderActorType, AccountID: asset.UploaderID, VisitorSessionHash: asset.UploaderVisitorSessionHash}
	return identity.IsValidMediaUploadActorIdentity() && session.UploadActorIdentity() == identity &&
		session.MediaAssetID != nil && *session.MediaAssetID == asset.ID && session.ObjectKey == key
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) OpenPublicCustomerServicePendingUploadPreview(
	ctx context.Context, conversationID string, owner CustomerServiceOwner, uploadSessionID, preset string,
) (*MediaAssetFile, error) {
	if s == nil || s.tickets == nil || s.media == nil {
		return nil, ErrMediaStorageUnavailable
	}
	conversation, err := s.tickets.GetPublicCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	identity := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: conversation.VisitorSessionHash}
	if conversation.CustomerUserID != nil {
		identity = media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: *conversation.CustomerUserID}
	}
	return s.openOwnedCustomerServicePendingUploadPreview(ctx, conversation.ID, identity, uploadSessionID, preset)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) OpenAgentCustomerServicePendingUploadPreview(
	ctx context.Context, ticketID, agentUserID uint, canViewAll bool, uploadSessionID, preset string,
) (*MediaAssetFile, error) {
	if s == nil || s.tickets == nil || s.media == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if agentUserID == 0 {
		return nil, ErrCustomerServiceAgentAccessDenied
	}
	if _, err := s.tickets.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll); err != nil {
		return nil, err
	}
	return s.openOwnedCustomerServicePendingUploadPreview(ctx, ticketID, media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: agentUserID}, uploadSessionID, preset)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) openOwnedCustomerServicePendingUploadPreview(
	ctx context.Context, ticketID uint, identity media.MediaUploadActorIdentity, uploadSessionID, preset string,
) (*MediaAssetFile, error) {
	session, err := s.media.repo.FindCustomerServiceMediaUploadSession(ctx, uploadSessionID)
	if err != nil {
		return nil, ErrCustomerServiceMessageAttachmentUnavailable
	}
	if session.BusinessContextType != media.MediaUploadSessionCustomerServiceContext || session.BusinessContextID != ticketID ||
		session.Status != "ready" || session.MediaAssetID == nil || !session.ExpiresAt.After(time.Now().UTC()) ||
		session.UploadActorIdentity() != identity {
		return nil, ErrCustomerServiceMessageAttachmentUnavailable
	}
	asset, err := s.media.repo.FindAssetByID(*session.MediaAssetID)
	if err != nil {
		return nil, ErrCustomerServiceMessageAttachmentUnavailable
	}
	asset, key, err := validateRegisteredCustomerServiceAttachmentAsset(asset, asset.StorageKey)
	if err != nil || !customerServiceUploadSessionMatchesRegisteredAsset(session, asset, key) || asset.Visibility != "private" ||
		!storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
		return nil, ErrCustomerServiceMessageAttachmentUnavailable
	}
	return s.openCustomerServiceAttachmentAssetFile(ctx, asset, key, preset)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) PublicCustomerServicePendingUploadPreviewURL(conversationID, uploadSessionID string) string {
	return fmt.Sprintf("%s/api/v1/customer-service/conversations/%s/attachment-uploads/%s/file", strings.TrimRight(s.media.siteURL, "/"), url.PathEscape(conversationID), url.PathEscape(uploadSessionID))
}

func AgentCustomerServicePendingUploadPreviewURL(ticketID uint, uploadSessionID string) string {
	return fmt.Sprintf("/api/admin/customer-service/conversations/%d/attachment-uploads/%s/file", ticketID, url.PathEscape(uploadSessionID))
}
