package service

import (
	"context"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func (s *PaymentService) ConfigurePayPalInvoiceDurableAttachmentUploads(uploads *SuggestionAndPayPalAttachmentUploadService) {
	if s != nil {
		s.paypalDisputeAttachmentUploads = uploads
	}
}

func (s *PaymentService) configurePayPalInvoiceDurableUploadProvider(provider PayPalDisputeEvidenceDocumentStorage) {
	// The old UploadFromReader-only capability cannot authorize an upload.
	// Use an independent configuration copy so feedback keeps its own provider.
	clonedValue := *s.paypalDisputeAttachmentUploads
	clonedValue.storage, _ = provider.(storage.StorageService)
	s.paypalDisputeAttachmentUploads = &clonedValue
}

func (s *PaymentService) createCanonicalPayPalSubmissionSnapshotAndBindNewInvoiceReceipts(ctx context.Context, candidate *orderevidence.OrderEvidenceSubmissionSnapshot, documents []PayPalDisputeEvidenceDocument) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	uploads := s.paypalDisputeAttachmentUploads
	if uploads == nil || uploads.outbox == nil {
		return s.createOrGetEvidenceSubmissionSnapshot(candidate)
	}
	if s.orderEvidenceSubmissionRepo == nil {
		return nil, ErrOrderEvidenceSubmissionSnapshotStoreUnavailable
	}
	var canonical *orderevidence.OrderEvidenceSubmissionSnapshot
	err := uploads.outbox.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(ctx, func(tx *gorm.DB) error {
		var err error
		ownerRepo := s.orderEvidenceSubmissionRepo.WithTx(tx)
		ownerRepo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(uploads)
		canonical, err = ownerRepo.CreateOrGetLatestWithAtomicPayPalInvoiceReceiptBinding(candidate, func(ownerTx *gorm.DB, keys []string) error {
			if uploads.storage == nil {
				return ErrObjectStorageCleanupUnavailable
			}
			documentsByKey := make(map[string]PayPalDisputeEvidenceDocument)
			for _, document := range documents {
				if document.Type != "commercial_invoice" {
					continue
				}
				key, keyErr := uploads.storage.ObjectKey(document.URL)
				namespace, ownerID, _, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
				if keyErr != nil || !valid || namespace != storage.PayPalDisputeInvoiceAttachmentNamespace || ownerID != candidate.DisputeID || document.URL != uploads.storage.GetURL(key) {
					return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
				}
				documentsByKey[key] = document
			}
			for _, key := range keys {
				document, exists := documentsByKey[key]
				if !exists {
					return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
				}
				if err := uploads.outbox.WithTx(ownerTx).ConsumeReadySuggestionOrPayPalAttachmentReceiptWithinOwnerTransaction(key, candidate.DisputeID, func(payload *outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload) error {
					if payload.ContentSHA256 != document.SHA256 || payload.Name != document.Name {
						return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
					}
					return uploads.verifySuggestionOrPayPalAttachmentReceiptBytes(ownerTx.Statement.Context, payload)
				}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		return canonical.Validate()
	})
	return canonical, err
}
