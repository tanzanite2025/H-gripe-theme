package service

import (
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"mime/multipart"
	"strings"
)

// Preserves historical returned-identity upload compensation coverage.
func (s *OrderEvidenceAttachmentService) uploadLegacyOrderEvidenceAttachmentWithReturnedIdentityForRegression(
	ctx context.Context,
	orderID uint,
	itemID uint,
	file *multipart.FileHeader,
	uploadedBy uint,
) (*orderevidence.OrderEvidenceAttachment, error) {
	if s == nil || s.txManager == nil || s.evidenceRepo == nil || s.storage == nil || s.outbox == nil {
		return nil, ErrOrderEvidenceAttachmentUnavailable
	}
	if orderID == 0 || itemID == 0 || uploadedBy == 0 {
		return nil, errors.New("order id, evidence item id, and uploader are required")
	}
	item, pkg, err := s.evidenceRepo.FindItemAndPackageByIDForOrder(itemID, orderID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidenceAttachmentItemNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := pkg.EnsureMutable(); err != nil {
		return nil, err
	}
	validationRule := upload.WarrantyImageRule.FileRule
	if item.ItemType == orderevidence.EvidenceItemTypeSignedPOD {
		// Keep legacy signed-POD images valid, while accepting carrier-issued
		// PDFs that can be forwarded to PayPal as official documents.
		contentType, detectErr := upload.DetectContentType(file)
		if detectErr != nil {
			return nil, detectErr
		}
		if strings.EqualFold(strings.TrimSpace(contentType), "application/pdf") {
			validationRule = upload.OrderEvidencePODDocumentRule
		}
	}
	if err := upload.ValidateFile(file, validationRule); err != nil {
		return nil, err
	}

	sha256Value, sizeBytes, err := hashMultipartFile(file)
	if err != nil {
		return nil, fmt.Errorf("hash order evidence attachment: %w", err)
	}
	mimeType, err := upload.DetectContentType(file)
	if err != nil {
		return nil, err
	}

	// A fresh attempt namespace grants no authority over pre-upgrade objects,
	// whose deleted reference history cannot be reconstructed.
	prefix := fmt.Sprintf("order-evidence/%d/%d/owned-upload-%s", orderID, itemID, uuid.NewString())
	privateUploader, ok := s.storage.(storage.PrivateObjectUploader)
	if !ok {
		return nil, fmt.Errorf("%w: private evidence storage is required", ErrOrderEvidenceAttachmentUnavailable)
	}
	referenceURL, uploadErr := privateUploader.UploadWithPrefixPrivate(ctx, file, prefix)
	if uploadErr != nil {
		if referenceURL == "" {
			return nil, fmt.Errorf("upload order evidence attachment: %w", uploadErr)
		}
		return nil, errors.Join(fmt.Errorf("upload order evidence attachment: %w", uploadErr), s.persistFailedOrderEvidenceUploadBeforeOwnerCheckedCompensation(ctx, prefix, referenceURL))
	}

	storageKey, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, prefix, referenceURL)
	if err != nil {
		// No trustworthy identity means no authority to delete another object.
		return nil, fmt.Errorf("resolve order evidence attachment key: %w", err)
	}
	attachmentInput := OrderEvidenceAttachmentReferenceInput{
		EvidenceItemID:   item.ID,
		StorageKey:       storageKey,
		OriginalFilename: file.Filename,
		MimeType:         mimeType,
		SizeBytes:        sizeBytes,
		SHA256:           sha256Value,
		UploadedBy:       uploadedBy,
	}

	var attachment *orderevidence.OrderEvidenceAttachment
	err = s.evidenceRepo.WithLockedOrderEvidenceAttachmentLifecycleTransaction(ctx, func(lockedContext context.Context, repo *repository.OrderEvidenceRepository) error {
		var registerErr error
		attachment, registerErr = s.registerOrderEvidenceReferenceAfterLifecycleLock(lockedContext, repo, attachmentInput)
		return registerErr
	})
	if err != nil {
		return nil, errors.Join(err, s.persistFailedOrderEvidenceUploadBeforeOwnerCheckedCompensation(ctx, prefix, referenceURL))
	}
	return attachment, nil
}
