package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *OrderEvidenceAttachmentService) uploadOrderEvidenceAttachmentWithPrecommittedExactKeyAndAtomicOwnerExchange(ctx context.Context, orderID uint, plannedItem *orderevidence.OrderEvidenceItem, file *multipart.FileHeader, uploadedBy uint, mimeType, digest string, sizeBytes int64) (*orderevidence.OrderEvidenceAttachment, error) {
	writer, ok := s.storage.(storage.PreallocatedPrivateObjectWriter)
	if !ok {
		return nil, fmt.Errorf("%w: preallocated private evidence storage is required", ErrOrderEvidenceAttachmentUnavailable)
	}
	prefix := fmt.Sprintf("order-evidence/%d/%d/owned-upload-%s", orderID, plannedItem.ID, uuid.NewString())
	key, err := storage.AllocatePrivateObjectKeyForPendingUpload(file.Filename, prefix)
	if err != nil {
		return nil, err
	}
	attemptID := uuid.NewString()
	receipt, err := newObjectStorageCleanupEvent(objectCleanupResourceFailedOrderEvidenceUpload, prefix, "order_evidence", attemptID, []string{key}, map[string]int64{key: sizeBytes})
	if err != nil {
		return nil, err
	}
	receipt.EventKey = repository.OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix + attemptID
	input := OrderEvidenceAttachmentReferenceInput{EvidenceItemID: plannedItem.ID, StorageKey: key, OriginalFilename: file.Filename, MimeType: mimeType, SizeBytes: sizeBytes, SHA256: digest, UploadedBy: uploadedBy}
	var attachment *orderevidence.OrderEvidenceAttachment
	err = s.outbox.RunOrderEvidenceAttachmentObjectWritesWithPrecommittedCleanupReceiptAndAtomicOwnerExchange(ctx, receipt, func(tx *gorm.DB) error {
		repo := s.evidenceRepo.WithTx(tx)
		// Package locking, revisions and attachment mutations share this namespace.
		// Recheck the planned item and order before sending any bytes.
		currentItem, currentPackage, err := repo.FindItemAndPackageByIDForOrder(plannedItem.ID, orderID)
		if repository.IsRecordNotFound(err) {
			return ErrOrderEvidenceAttachmentItemNotFound
		}
		if err != nil {
			return err
		}
		if currentItem.PackageID != plannedItem.PackageID || currentItem.ItemType != plannedItem.ItemType {
			return ErrOrderEvidenceAttachmentOwnership
		}
		if err := currentPackage.EnsureMutable(); err != nil {
			return err
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		reference, writeErr := writer.WritePrivateObjectAtPreallocatedKey(tx.Statement.Context, reader, key)
		if err := errors.Join(writeErr, reader.Close()); err != nil {
			return err
		}
		returnedKey, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, prefix, reference)
		if err != nil || returnedKey != key {
			return ErrOrderEvidenceAttachmentOwnership
		}
		// Registration verifies exact item ownership, source bytes, size, SHA256 and
		// MIME; its owner write shares the final receipt exchange transaction.
		attachment, err = s.registerOrderEvidenceReferenceAfterLifecycleLock(tx.Statement.Context, repo, input)
		return err
	})
	if err != nil {
		return nil, err
	}
	return attachment, nil
}
