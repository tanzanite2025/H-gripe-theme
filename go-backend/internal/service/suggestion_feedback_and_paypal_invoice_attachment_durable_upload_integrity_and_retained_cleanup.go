package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"reflect"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

type SuggestionAndPayPalAttachmentUploadService struct {
	outbox  *repository.OutboxRepository
	storage storage.StorageService
	siteURL string
}

func NewSuggestionAndPayPalAttachmentUploadService(events *repository.OutboxRepository, provider storage.StorageService, siteURL string) *SuggestionAndPayPalAttachmentUploadService {
	return &SuggestionAndPayPalAttachmentUploadService{outbox: events, storage: provider, siteURL: siteURL}
}

func (s *SuggestionAndPayPalAttachmentUploadService) uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(ctx context.Context, namespace string, ownerID uint, filename string, payloadBytes []byte) (string, error) {
	if s == nil || s.outbox == nil || s.storage == nil {
		return "", ErrObjectStorageCleanupUnavailable
	}
	writer, supported := s.storage.(storage.PreallocatedSuggestionOrPayPalAttachmentWriter)
	_, readable := s.storage.(storage.ObjectOpener)
	if !supported || !readable {
		return "", ErrObjectStorageCleanupUnavailable
	}
	mimeType := upload.DetectContentTypeFromHeaderIncludingVideoFormats(payloadBytes)
	keyFilename := filename
	if namespace == storage.SuggestionFeedbackAttachmentNamespace {
		switch mimeType {
		case "image/png":
			keyFilename = "attachment.png"
		case "image/jpeg":
			keyFilename = "attachment.jpg"
		case "image/webp":
			keyFilename = "attachment.webp"
		case "image/gif":
			keyFilename = "attachment.gif"
		}
	}
	key, err := storage.AllocateSuggestionOrPayPalAttachmentObjectKey(namespace, ownerID, keyFilename)
	if err != nil {
		return "", err
	}
	_, _, id, _ := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
	receipt, err := newObjectStorageCleanupEvent(outbox.SuggestionAndPayPalAttachmentCleanupResourceType, id, "suggestion_paypal_attachment", id, []string{key})
	if err != nil {
		return "", err
	}
	receipt.EventKey = outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix + id
	var payload outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload
	if err := json.Unmarshal(receipt.Payload, &payload.ObjectStorageCleanupPayload); err != nil {
		return "", err
	}
	payload.OwnerID, payload.Name, payload.Size = ownerID, path.Base(filename), int64(len(payloadBytes))
	payload.MimeType = mimeType
	payload.ContentSHA256 = fmt.Sprintf("%x", sha256.Sum256(payloadBytes))
	url := s.storage.GetURL(key)
	integrity := &MediaService{storage: s.storage}
	if err := integrity.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(url, key); err != nil {
		return "", err
	}
	if namespace == storage.PayPalDisputeInvoiceAttachmentNamespace && !paypalEvidenceDocumentURLUsable(url) {
		return "", fmt.Errorf("invoice storage cannot represent a public HTTPS object URL")
	}
	err = s.outbox.RunSuggestionOrPayPalAttachmentUploadWithPrecommittedReceipt(ctx, receipt, &payload, func(tx *gorm.DB) error {
		returned, err := writer.WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(tx.Statement.Context, bytes.NewReader(payloadBytes), key)
		if err != nil {
			return err
		}
		if err := integrity.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(returned, key); err != nil {
			return err
		}
		return s.verifySuggestionOrPayPalAttachmentReceiptBytes(tx.Statement.Context, &payload)
	})
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *SuggestionAndPayPalAttachmentUploadService) verifySuggestionOrPayPalAttachmentReceiptBytes(ctx context.Context, payload *outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload) error {
	return (&MediaService{storage: s.storage}).verifyStoredMediaObjectSourceSizeDigestAndMimeType(ctx, &media.MediaAsset{StorageKey: payload.ObjectKeys[0], Size: payload.Size, MimeType: payload.MimeType, ContentSHA256: payload.ContentSHA256})
}

func (s *SuggestionAndPayPalAttachmentUploadService) uploadSuggestionAttachmentFileWithDurableOwnerReceipt(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error) {
	if err := upload.ValidateSpecFile(file, string(upload.SpecSuggestionAttachment)); err != nil {
		return "", err
	}
	reader, err := file.Open()
	if err != nil {
		return "", err
	}
	payload, readErr := io.ReadAll(io.LimitReader(reader, (5<<20)+1))
	if err := errors.Join(readErr, reader.Close()); err != nil {
		return "", err
	}
	if int64(len(payload)) != file.Size {
		return "", fmt.Errorf("feedback attachment source size mismatch")
	}
	return s.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(ctx, storage.SuggestionFeedbackAttachmentNamespace, userID, file.Filename, payload)
}

func (s *SuggestionAndPayPalAttachmentUploadService) cleanupSuggestionOrPayPalAttachmentFromOriginalReceipt(ctx context.Context, event outbox.Event) error {
	if s == nil || s.outbox == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	var supplied outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload
	if json.Unmarshal(event.Payload, &supplied) != nil || len(supplied.ObjectKeys) != 1 {
		return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	key := supplied.ObjectKeys[0]
	return s.outbox.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(ctx, func(tx *gorm.DB) error {
		stored, payload, err := s.outbox.WithTx(tx).FindSuggestionOrPayPalAttachmentReceiptForUpdate(key)
		if err != nil {
			return err
		}
		if stored.ID != event.ID || !reflect.DeepEqual(payload, &supplied) ||
			(stored.Status != outbox.EventStatusProcessing && !originalObjectStorageCleanupReceiptHasMatchingReconciliationLease(*stored, event)) {
			return repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable
		}
		references := repository.NewMediaRepository(tx)
		references.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: s.storage, siteURL: s.siteURL})
		claims, err := references.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
		if err != nil {
			return err
		}
		if len(claims) > 0 {
			return &MediaAssetInUseError{References: claims}
		}
		return s.storage.Delete(tx.Statement.Context, key)
	})
}
