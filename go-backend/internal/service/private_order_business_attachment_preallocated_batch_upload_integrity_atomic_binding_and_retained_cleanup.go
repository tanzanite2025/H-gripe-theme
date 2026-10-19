package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PrivateOrderBusinessAttachmentUploadFiles struct {
	Images []*multipart.FileHeader
	Video  *multipart.FileHeader
}

type PrivateOrderBusinessAttachmentUploadService struct {
	outbox  *repository.OutboxRepository
	storage storage.StorageService
	siteURL string
}

type preparedPrivateOrderBusinessAttachment struct {
	file                                   *multipart.FileHeader
	key, reference, mimeType, digest, kind string
	size                                   int64
}

func NewPrivateOrderBusinessAttachmentUploadService(events *repository.OutboxRepository, provider storage.StorageService, siteURL string) *PrivateOrderBusinessAttachmentUploadService {
	return &PrivateOrderBusinessAttachmentUploadService{outbox: events, storage: provider, siteURL: siteURL}
}

func (s *PrivateOrderBusinessAttachmentUploadService) uploadPrivateOrderBusinessAttachmentBatchAndBindOwner(ctx context.Context, prefix string, files PrivateOrderBusinessAttachmentUploadFiles, bindOwner func(*gorm.DB, []preparedPrivateOrderBusinessAttachment) error) error {
	if s == nil || s.outbox == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	writer, writable := s.storage.(storage.PreallocatedPrivateObjectWriter)
	_, readable := s.storage.(storage.ObjectOpener)
	if !writable || !readable {
		return ErrObjectStorageCleanupUnavailable
	}
	if err := upload.ValidateSpecFiles(files.Images, string(upload.SpecWarrantyEvidence)); err != nil {
		return err
	}
	all := append([]*multipart.FileHeader{}, files.Images...)
	if files.Video != nil {
		if err := upload.ValidateFile(files.Video, upload.WarrantyVideoRule); err != nil {
			return err
		}
		all = append(all, files.Video)
	}
	if err := upload.ValidateTotalSize(all, 80<<20); err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("private attachment upload requires files")
	}
	prefix += "/owned-upload-" + uuid.NewString()
	prepared := make([]preparedPrivateOrderBusinessAttachment, 0, len(all))
	keys := make([]string, 0, len(all))
	for index, file := range all {
		key, err := storage.AllocatePrivateObjectKeyForPendingUpload(file.Filename, prefix)
		if err != nil {
			return err
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		header := make([]byte, 512)
		n, headerErr := io.ReadFull(reader, header)
		if headerErr != nil && !errors.Is(headerErr, io.EOF) && !errors.Is(headerErr, io.ErrUnexpectedEOF) {
			_ = reader.Close()
			return headerErr
		}
		digest := sha256.New()
		_, _ = digest.Write(header[:n])
		size, readErr := io.Copy(digest, io.LimitReader(reader, file.Size+1))
		if err := errors.Join(readErr, reader.Close()); err != nil {
			return err
		}
		size += int64(n)
		if size != file.Size {
			return fmt.Errorf("private attachment source size mismatch")
		}
		kind := aftersales.AttachmentKindImage
		if index >= len(files.Images) {
			kind = aftersales.AttachmentKindVideo
		}
		prepared = append(prepared, preparedPrivateOrderBusinessAttachment{file: file, key: key, reference: s.storage.GetURL(key), mimeType: upload.DetectContentTypeFromHeaderIncludingVideoFormats(header[:n]), digest: fmt.Sprintf("%x", digest.Sum(nil)), size: size, kind: kind})
		keys = append(keys, key)
	}
	if err := repository.ValidatePrivateOrderBusinessAttachmentUploadAttempt(prefix, keys); err != nil {
		return err
	}
	expectedBytes := make(map[string]int64, len(prepared))
	for _, object := range prepared {
		expectedBytes[object.key] = object.size
	}
	receipt, err := newObjectStorageCleanupEvent(repository.PrivateOrderBusinessAttachmentFailedUploadResourceType, prefix, "private_order_business_attachment", uuid.NewString(), keys, expectedBytes)
	if err != nil {
		return err
	}
	receipt.EventKey = repository.PrivateOrderBusinessAttachmentReceiptEventKeyPrefix + uuid.NewString()
	return s.outbox.RunPrivateOrderBusinessAttachmentBatchWithPrecommittedReceiptAndAtomicOwnerExchange(ctx, receipt, func(tx *gorm.DB) error {
		// Tentative ownership and authorization are checked before sending bytes.
		// The rows, challenge consumption and receipt exchange commit together.
		if err := bindOwner(tx, prepared); err != nil {
			return err
		}
		integrity := &MediaService{storage: s.storage}
		for _, object := range prepared {
			reader, err := object.file.Open()
			if err != nil {
				return err
			}
			reference, writeErr := writer.WritePrivateObjectAtPreallocatedKey(tx.Statement.Context, reader, object.key)
			if err := errors.Join(writeErr, reader.Close()); err != nil {
				return err
			}
			if err := integrity.validateMediaProviderReturnedCanonicalURLAndExactObjectKey(reference, object.key); err != nil {
				return err
			}
			if err := integrity.verifyStoredMediaObjectSourceSizeDigestAndMimeType(tx.Statement.Context, &media.MediaAsset{StorageKey: object.key, Size: object.size, MimeType: object.mimeType, ContentSHA256: object.digest}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PrivateOrderBusinessAttachmentUploadService) cleanupPrivateOrderBusinessAttachmentFailedBatchWithRetainedReferences(ctx context.Context, prefix string, keys []string) error {
	if s == nil || s.outbox == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if err := repository.ValidatePrivateOrderBusinessAttachmentUploadAttempt(prefix, keys); err != nil {
		return err
	}
	return s.outbox.WithPrivateOrderBusinessAttachmentCleanupLockUntilProviderReturns(ctx, func(tx *gorm.DB) error {
		references := repository.NewMediaRepository(tx)
		references.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: s.storage, siteURL: s.siteURL})
		// Check the entire batch before deleting any object. Keep the same event
		// retryable while active or retained historical ownership still exists.
		for _, key := range keys {
			claims, err := references.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
			if err != nil {
				return err
			}
			if len(claims) > 0 {
				return &MediaAssetInUseError{References: claims}
			}
		}
		for _, key := range keys {
			if err := s.storage.Delete(ctx, key); err != nil {
				return err
			}
		}
		return nil
	})
}

func privateOrderBusinessAttachmentURLs(prepared []preparedPrivateOrderBusinessAttachment) ([]string, string) {
	images := make([]string, 0, len(prepared))
	video := ""
	for _, object := range prepared {
		if object.kind == aftersales.AttachmentKindVideo {
			video = object.reference
		} else {
			images = append(images, object.reference)
		}
	}
	return images, video
}

func privateOrderBusinessAfterSalesAttachments(prepared []preparedPrivateOrderBusinessAttachment) []aftersales.AfterSalesCaseAttachment {
	attachments := make([]aftersales.AfterSalesCaseAttachment, 0, len(prepared))
	for _, object := range prepared {
		filename := filepath.Base(strings.ReplaceAll(strings.TrimSpace(object.file.Filename), "\\", "/"))
		if len(filename) > 255 {
			filename = filename[len(filename)-255:]
		}
		attachments = append(attachments, aftersales.AfterSalesCaseAttachment{Kind: object.kind, StorageURL: object.reference, Filename: filename, ContentType: object.mimeType, SizeBytes: object.size})
	}
	return attachments
}
