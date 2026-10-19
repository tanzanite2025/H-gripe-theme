package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
)

func (h *ObjectStorageCleanupOutboxHandler) ConfigureSuggestionAndPayPalAttachmentUploads(uploads *SuggestionAndPayPalAttachmentUploadService) {
	if h != nil {
		h.suggestionAndPayPalAttachments = uploads
	}
}

func (s *PublicUploadAccessService) ConfigureSuggestionAndPayPalAttachmentUploads(uploads *SuggestionAndPayPalAttachmentUploadService) {
	if s != nil {
		s.suggestionAndPayPalAttachments = uploads
	}
}

// These URLs preserve the existing public attachment contract. An uploading,
// failed, expired or unbound invoice never gains permission through a colliding
// generic media record. Private customer attachments use their separate API.
func (s *SuggestionAndPayPalAttachmentUploadService) canServeOwnedSuggestionOrPayPalAttachment(ctx context.Context, key string) (bool, error) {
	if s == nil || s.outbox == nil || s.storage == nil {
		return false, nil
	}
	namespace, ownerID, _, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
	if !valid {
		return false, nil
	}
	allowed := false
	err := s.outbox.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(ctx, func(tx *gorm.DB) error {
		if namespace == storage.SuggestionFeedbackAttachmentNamespace {
			event, payload, err := s.outbox.WithTx(tx).FindSuggestionOrPayPalAttachmentReceiptForUpdate(key)
			if err == nil {
				allowed = event.Status == outbox.EventStatusPending && event.Attempts == 0 && event.LastAttemptAt == nil && payload.UploadState == "ready" && payload.BindingExpiresAt != nil && time.Now().Before(*payload.BindingExpiresAt) && time.Now().Before(event.AvailableAt)
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			var rows []suggestionfeedback.SuggestionFeedback
			if err := tx.Select("id", "attachments").Where("user_id = ?", ownerID).Find(&rows).Error; err != nil {
				return err
			}
			for _, row := range rows {
				var attachments []suggestionfeedback.Attachment
				if err := json.Unmarshal(row.Attachments, &attachments); err != nil {
					return err
				}
				for _, attachment := range attachments {
					if attachment.URL == s.storage.GetURL(key) {
						allowed = true
						return nil
					}
				}
			}
			return nil
		}
		var snapshots []orderevidence.OrderEvidenceSubmissionSnapshot
		if err := tx.Where("provider = ? AND dispute_id = ?", "paypal", ownerID).Find(&snapshots).Error; err != nil {
			return err
		}
		for _, snapshot := range snapshots {
			if err := snapshot.Validate(); err != nil {
				return err
			}
			var payload paypalDisputeEvidenceSnapshotPayload
			if err := json.Unmarshal(snapshot.SnapshotData, &payload); err != nil {
				return err
			}
			for _, document := range payload.Request.Documents {
				if document.Type == "commercial_invoice" && document.URL == s.storage.GetURL(key) {
					allowed = true
					return nil
				}
			}
		}
		return nil
	})
	return allowed, err
}
