package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

func (s *MediaService) ConfigureCustomerServicePrivateAttachmentUploads(enabled bool) {
	if s != nil {
		s.customerServicePrivateAttachmentUploadsEnabled = enabled
	}
}

func (s *MediaService) CustomerServicePrivateAttachmentUploadsEnabled() bool {
	return s != nil && s.customerServicePrivateAttachmentUploadsEnabled
}

func (s *CustomerServiceConversationAttachmentUploadService) UploadVerifiedAgentCustomerServiceConversationAttachment(
	ctx context.Context, ticketID, agentUserID uint, canViewAll bool, file *multipart.FileHeader,
) (*media.MediaAsset, error) {
	if s == nil || s.media == nil || s.tickets == nil {
		return nil, ErrMediaStorageUnavailable
	}
	conversation, err := s.tickets.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll)
	if err != nil {
		return nil, err
	}
	mediaType, width, height, err := validateAgentCustomerServiceAttachmentForCurrentPrivateUploadRollout(file, s.media.CustomerServicePrivateAttachmentUploadsEnabled())
	if err != nil {
		return nil, err
	}
	identity := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: agentUserID}
	if !s.media.CustomerServicePrivateAttachmentUploadsEnabled() {
		return s.media.UploadAsset(ctx, MediaUploadInput{File: file, MediaType: mediaType, UploadActorIdentity: identity, Width: width, Height: height,
			UploadOrigin: media.MediaUploadOriginCustomerServiceConversation, UploadPurpose: string(upload.SpecCustomerServiceAttachment)})
	}
	return s.uploadPreparedPrivateCustomerServiceConversationAttachment(ctx, conversation, identity, file, width, height, CustomerServiceOwner{}, agentUserID, canViewAll)
}

func (s *CustomerServiceConversationAttachmentUploadService) uploadPreparedPrivateCustomerServiceConversationAttachment(
	ctx context.Context, conversation *ticket.Ticket, identity media.MediaUploadActorIdentity, file *multipart.FileHeader, width, height int,
	owner CustomerServiceOwner, agentUserID uint, canViewAll bool,
) (*media.MediaAsset, error) {
	if _, ok := s.media.storage.(storage.PreallocatedPrivateObjectWriter); !ok {
		return nil, ErrMediaStorageUnavailable
	}
	if err := s.media.ensureMediaUploadActorStorageQuota(identity, file.Size); err != nil {
		return nil, err
	}
	key, err := storage.AllocatePrivateObjectKeyForPendingUpload(file.Filename, storage.CustomerServiceConversationAttachmentOriginalsPrefix)
	if err != nil {
		return nil, err
	}
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: conversation.ID, UploadActorType: identity.ActorType, UploadActorAccountID: identity.AccountID,
		UploadActorVisitorSessionHash: identity.VisitorSessionHash, ObjectKey: key, Status: "uploading", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)}
	err = s.tickets.ticketRepo.WithinTx(func(repo *repository.TicketRepository, tx *gorm.DB) error {
		current, lookupErr := repo.FindTicketByIDForUpdate(conversation.ID)
		if lookupErr != nil {
			return lookupErr
		}
		if agentUserID > 0 {
			if permissionErr := validateAgentCustomerServiceConversation(current, agentUserID, canViewAll); permissionErr != nil {
				return permissionErr
			}
		} else if current.Category != customerServiceTicketCategory || !customerServiceOwnerMatches(current, normalizeCustomerServiceOwner(owner)) {
			return ErrCustomerServiceConversationAccessDenied
		}
		return s.media.repo.WithTx(tx).CreateCustomerServicePendingUploadSession(ctx, &session)
	})
	if err != nil {
		return nil, err
	}
	asset, err := s.media.UploadAsset(ctx, MediaUploadInput{File: file, MediaType: "image", Alt: "Customer service attachment", UploadActorIdentity: identity, Width: width, Height: height,
		UploadOrigin: media.MediaUploadOriginCustomerServiceConversation, UploadPurpose: string(upload.SpecCustomerServiceAttachment),
		PreparedCustomerServiceUploadSessionID: session.ID, PreparedCustomerServiceOriginalObjectKey: key})
	if err != nil {
		// The persisted key survives a canceled request or a failed compensation.
		// A durable session remains available for the later cleanup workflow.
		failureContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.media.repo.MarkCustomerServiceUploadSessionFailed(failureContext, session.ID)
		return nil, err
	}
	asset.PendingCustomerServiceUploadSessionID = session.ID
	return asset, nil
}
