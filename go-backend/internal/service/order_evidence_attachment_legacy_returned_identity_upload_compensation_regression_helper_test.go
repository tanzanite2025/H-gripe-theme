package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

// Preserves historical returned-identity upload compensation coverage.
func (s *OrderEvidenceAttachmentService) uploadLegacyOrderEvidenceAttachmentWithReturnedIdentityForRegression(
	ctx context.Context,
	orderID uint,
	itemID uint,
	file *multipart.FileHeader,
	uploadedBy uint,
) error {
	if s == nil || s.txManager == nil || s.evidenceRepo == nil || s.storage == nil || s.outbox == nil {
		return ErrOrderEvidenceAttachmentUnavailable
	}
	if orderID == 0 || itemID == 0 || uploadedBy == 0 {
		return errors.New("order id, evidence item id, and uploader are required")
	}
	item, pkg, err := s.evidenceRepo.FindItemAndPackageByIDForOrder(itemID, orderID)
	if repository.IsRecordNotFound(err) {
		return ErrOrderEvidenceAttachmentItemNotFound
	}
	if err != nil {
		return err
	}
	if err := pkg.EnsureMutable(); err != nil {
		return err
	}
	validationRule := upload.WarrantyImageRule.FileRule
	if item.ItemType == orderevidence.EvidenceItemTypeSignedPOD {
		// Keep legacy signed-POD images valid, while accepting carrier-issued
		// PDFs that can be forwarded to PayPal as official documents.
		contentType, detectErr := upload.DetectContentType(file)
		if detectErr != nil {
			return detectErr
		}
		if strings.EqualFold(strings.TrimSpace(contentType), "application/pdf") {
			validationRule = upload.OrderEvidencePODDocumentRule
		}
	}
	if err := upload.ValidateFile(file, validationRule); err != nil {
		return err
	}

	sha256Value, sizeBytes, err := hashMultipartFile(file)
	if err != nil {
		return fmt.Errorf("hash order evidence attachment: %w", err)
	}
	mimeType, err := upload.DetectContentType(file)
	if err != nil {
		return err
	}

	// A fresh attempt namespace grants no authority over pre-upgrade objects,
	// whose deleted reference history cannot be reconstructed.
	prefix := fmt.Sprintf("order-evidence/%d/%d/owned-upload-%s", orderID, itemID, uuid.NewString())
	privateUploader, ok := s.storage.(storage.PrivateObjectUploader)
	if !ok {
		return fmt.Errorf("%w: private evidence storage is required", ErrOrderEvidenceAttachmentUnavailable)
	}
	referenceURL, uploadErr := privateUploader.UploadWithPrefixPrivate(ctx, file, prefix)
	if uploadErr != nil {
		if referenceURL == "" {
			return fmt.Errorf("upload order evidence attachment: %w", uploadErr)
		}
		return errors.Join(fmt.Errorf("upload order evidence attachment: %w", uploadErr), s.persistFailedOrderEvidenceUploadBeforeOwnerCheckedCompensation(ctx, prefix, referenceURL))
	}

	storageKey, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, prefix, referenceURL)
	if err != nil {
		// No trustworthy identity means no authority to delete another object.
		return fmt.Errorf("resolve order evidence attachment key: %w", err)
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

	err = s.evidenceRepo.WithLockedOrderEvidenceAttachmentLifecycleTransaction(ctx, func(lockedContext context.Context, repo *repository.OrderEvidenceRepository) error {
		_, registerErr := s.registerOrderEvidenceReferenceAfterLifecycleLock(lockedContext, repo, attachmentInput)
		return registerErr
	})
	if err != nil {
		return errors.Join(err, s.persistFailedOrderEvidenceUploadBeforeOwnerCheckedCompensation(ctx, prefix, referenceURL))
	}
	return nil
}
