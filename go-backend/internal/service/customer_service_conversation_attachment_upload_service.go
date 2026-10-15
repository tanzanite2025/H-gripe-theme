package service

import (
	"context"
	"mime/multipart"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/upload"
)

// CustomerServiceConversationAttachmentUploadService derives upload identity
// only after verifying ownership of the existing conversation. Public handlers
// cannot supply uploader IDs or visitor hashes through the multipart form.
type CustomerServiceConversationAttachmentUploadService struct {
	tickets *TicketService
	media   *MediaService
}

func NewCustomerServiceConversationAttachmentUploadService(tickets *TicketService, mediaService *MediaService) *CustomerServiceConversationAttachmentUploadService {
	return &CustomerServiceConversationAttachmentUploadService{tickets: tickets, media: mediaService}
}

func (s *CustomerServiceConversationAttachmentUploadService) UploadVerifiedCustomerServiceConversationAttachment(
	ctx context.Context, conversationID string, owner CustomerServiceOwner, file *multipart.FileHeader,
) (*media.MediaAsset, error) {
	if s == nil || s.media == nil || s.tickets == nil {
		return nil, ErrMediaStorageUnavailable
	}
	conversation, err := s.tickets.GetPublicCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	// Persist identity from the authorized business record, rather than copying
	// caller-provided owner values into the media table.
	identity := media.MediaUploadActorIdentity{
		ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: conversation.VisitorSessionHash,
	}
	if conversation.CustomerUserID != nil && *conversation.CustomerUserID > 0 {
		identity = media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: *conversation.CustomerUserID}
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecCustomerServiceAttachment)); err != nil {
		return nil, err
	}
	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return nil, err
	}
	if s.media.CustomerServicePrivateAttachmentUploadsEnabled() {
		return s.uploadPreparedPrivateCustomerServiceConversationAttachment(ctx, conversation, identity, file, width, height, owner, 0, false)
	}
	return s.media.UploadAsset(ctx, MediaUploadInput{
		UploadOrigin: media.MediaUploadOriginCustomerServiceConversation, UploadPurpose: string(upload.SpecCustomerServiceAttachment),
		File: file, MediaType: "image", Alt: "Customer service attachment",
		UploadActorIdentity: identity, Width: width, Height: height,
	})
}
