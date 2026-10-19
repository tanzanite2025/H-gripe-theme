package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

const objectCleanupResourceFailedOrderEvidenceUpload = "order_evidence_failed_upload"

func (s *OrderEvidenceAttachmentService) ConfigureObjectCleanupOutbox(repo *repository.OutboxRepository) {
	if s != nil {
		s.outbox = repo
	}
}

func (s *OrderEvidenceAttachmentService) validateOrderEvidenceAttachmentStoredIdentityAndBytesAfterLifecycleLock(ctx context.Context, attachment *orderevidence.OrderEvidenceAttachment) error {
	if s.storage == nil {
		return ErrOrderEvidenceAttachmentUnavailable
	}
	key, err := s.storage.ObjectKey(attachment.StorageKey)
	if err != nil || key != attachment.StorageKey || !storage.IsPrivateObjectKey(key) {
		return ErrOrderEvidenceAttachmentOwnership
	}
	opener, ok := s.storage.(storage.ObjectOpener)
	if !ok {
		return ErrOrderEvidenceAttachmentUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		return fmt.Errorf("verify evidence attachment bytes: %w", err)
	}
	if object == nil || object.ReadCloser == nil {
		return ErrOrderEvidenceAttachmentUnavailable
	}
	defer func() { _ = object.ReadCloser.Close() }()
	// Read at most the claimed size plus one, and derive MIME from bytes rather
	// than trusting provider headers or caller metadata.
	if attachment.SizeBytes <= 0 || attachment.SizeBytes > 20<<20 {
		return errors.New("evidence attachment size is invalid")
	}
	prefix := make([]byte, 512)
	count, readErr := io.ReadFull(io.LimitReader(object.ReadCloser, attachment.SizeBytes+1), prefix)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return readErr
	}
	prefix = prefix[:count]
	hasher := sha256.New()
	_, _ = hasher.Write(prefix)
	remainder, err := io.Copy(hasher, io.LimitReader(object.ReadCloser, attachment.SizeBytes+1-int64(count)))
	if err != nil {
		return err
	}
	if int64(count)+remainder != attachment.SizeBytes || !strings.EqualFold(fmt.Sprintf("%x", hasher.Sum(nil)), attachment.SHA256) ||
		!strings.EqualFold(http.DetectContentType(prefix), attachment.MimeType) {
		return errors.New("evidence attachment bytes do not match size, SHA256, or MIME")
	}
	return ctx.Err()
}

func validateOrderEvidenceFailedUploadAttemptNamespace(prefix string) error {
	parts := strings.Split(prefix, "/")
	if len(parts) != 4 || parts[0] != "order-evidence" {
		return ErrOrderEvidenceAttachmentOwnership
	}
	for _, value := range parts[1:3] {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil || id == 0 || strconv.FormatUint(id, 10) != value {
			return ErrOrderEvidenceAttachmentOwnership
		}
	}
	attempt := strings.TrimPrefix(parts[3], "owned-upload-")
	parsed, err := uuid.Parse(attempt)
	if err != nil || parsed.String() != attempt || parts[3] != "owned-upload-"+attempt {
		return ErrOrderEvidenceAttachmentOwnership
	}
	return nil
}

func (s *OrderEvidenceAttachmentService) persistFailedOrderEvidenceUploadBeforeOwnerCheckedCompensation(ctx context.Context, prefix, reference string) error {
	if s == nil || s.outbox == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if err := validateOrderEvidenceFailedUploadAttemptNamespace(prefix); err != nil {
		return err
	}
	key, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, prefix, reference)
	if err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceFailedOrderEvidenceUpload, prefix, "order_evidence", uuid.NewString(), []string{key})
	if err != nil {
		return err
	}
	if err := s.outbox.WithContext(cleanupContext).CreateEvent(event); err != nil {
		return fmt.Errorf("%w: persist evidence upload compensation: %w", ErrObjectStorageCleanupUnavailable, err)
	}
	_ = s.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(cleanupContext, prefix, []string{key})
	return nil
}

func (s *OrderEvidenceAttachmentService) cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(ctx context.Context, prefix string, keys []string) error {
	if s == nil || s.evidenceRepo == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if err := validateOrderEvidenceFailedUploadAttemptNamespace(prefix); err != nil {
		return err
	}
	for _, key := range keys {
		resolved, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, prefix, key)
		if err != nil || resolved != key {
			return ErrOrderEvidenceAttachmentOwnership
		}
	}
	return s.evidenceRepo.WithLockedOrderEvidenceAttachmentCleanupUntilProviderReturns(ctx, func(_ *repository.OrderEvidenceRepository, references *repository.MediaRepository) error {
		references.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: s.storage})
		for _, key := range keys {
			claims, err := references.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
			if err != nil {
				return err
			}
			if len(claims) > 0 {
				return &MediaAssetInUseError{References: claims}
			}
			if err := s.storage.Delete(ctx, key); err != nil {
				return fmt.Errorf("clean never-registered evidence upload: %w", err)
			}
		}
		return nil
	})
}
