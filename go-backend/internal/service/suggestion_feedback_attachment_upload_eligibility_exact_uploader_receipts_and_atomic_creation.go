package service

import (
	"context"
	"mime/multipart"
	"strings"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"gorm.io/gorm"
)

func (s *SuggestionFeedbackService) ConfigureSuggestionAttachmentDurableUploads(uploads *SuggestionAndPayPalAttachmentUploadService) {
	if s != nil {
		s.attachmentUploads = uploads
	}
}

func (s *SuggestionFeedbackService) UploadSuggestionAttachmentForEligibleUser(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error) {
	eligibility, err := s.GetEligibility(userID, userID != 0)
	if err != nil {
		return "", err
	}
	if !eligibility.CanAttach {
		return "", ErrSuggestionFeedbackAttachmentEligibility
	}
	return s.attachmentUploads.uploadSuggestionAttachmentFileWithDurableOwnerReceipt(ctx, userID, file)
}

func (s *SuggestionFeedbackService) createSuggestionWithExactUploaderReceiptsAndAtomicAttachmentBinding(ctx context.Context, item *suggestionfeedback.SuggestionFeedback, attachments []suggestionfeedback.Attachment) error {
	if len(attachments) == 0 {
		item.Attachments = suggestionfeedback.JSONFromAttachments(nil)
		return s.suggestionRepo.Create(item)
	}
	if len(attachments) > 3 {
		return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	if s.attachmentUploads == nil || s.attachmentUploads.outbox == nil || s.attachmentUploads.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	return s.attachmentUploads.outbox.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(ctx, func(tx *gorm.DB) error {
		eligibilityService := NewSuggestionFeedbackService(repository.NewSuggestionFeedbackRepository(tx))
		eligibility, err := eligibilityService.GetEligibility(item.UserID, item.UserID != 0)
		if err != nil {
			return err
		}
		if !eligibility.CanAttach {
			return ErrSuggestionFeedbackAttachmentEligibility
		}
		item.MemberLevelRequired, item.MemberLevelMet = eligibility.RequiredLevel, true
		bound := make([]suggestionfeedback.Attachment, 0, len(attachments))
		seen := make(map[string]bool)
		for _, attachment := range attachments {
			url := strings.TrimSpace(attachment.URL)
			key, err := s.attachmentUploads.storage.ObjectKey(url)
			namespace, ownerID, _, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
			if err != nil || !valid || namespace != storage.SuggestionFeedbackAttachmentNamespace || ownerID != item.UserID || url != s.attachmentUploads.storage.GetURL(key) || seen[key] {
				return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
			}
			seen[key] = true
			if err := s.attachmentUploads.outbox.WithTx(tx).ConsumeReadySuggestionOrPayPalAttachmentReceiptWithinOwnerTransaction(key, item.UserID, func(payload *outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload) error {
				if err := s.attachmentUploads.verifySuggestionOrPayPalAttachmentReceiptBytes(tx.Statement.Context, payload); err != nil {
					return err
				}
				bound = append(bound, suggestionfeedback.Attachment{Name: payload.Name, Size: payload.Size, URL: s.attachmentUploads.storage.GetURL(key)})
				return nil
			}); err != nil {
				return err
			}
		}
		item.Attachments = suggestionfeedback.JSONFromAttachments(bound)
		return repository.NewSuggestionFeedbackRepository(tx).Create(item)
	})
}
