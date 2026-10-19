package service

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

const (
	AfterSalesEvidenceUploadPrefix = "after-sales/pending"
)

type AfterSalesAttachmentFile struct {
	ReadCloser io.ReadCloser
	Filename   string
	MimeType   string
	Size       int64
}

func (s *AfterSalesService) ConfigureAttachmentStorage(storageService storage.StorageService) {
	if s != nil {
		s.attachmentStorage = storageService
	}
}

func (s *AfterSalesService) OpenAttachmentFile(
	ctx context.Context,
	caseID uint,
	attachmentID uint,
) (*AfterSalesAttachmentFile, error) {
	if s == nil || s.caseRepo == nil {
		return nil, errors.New("after-sales service is not configured")
	}
	if s.attachmentStorage == nil {
		return nil, ErrAfterSalesAttachmentStorageUnavailable
	}
	if caseID == 0 {
		return nil, ErrAfterSalesCaseNotFound
	}
	if attachmentID == 0 {
		return nil, ErrAfterSalesAttachmentNotFound
	}

	attachment, err := s.caseRepo.FindAttachment(caseID, attachmentID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesAttachmentNotFound
		}
		return nil, err
	}

	object, err := openPrivateOrderBusinessAttachmentFromOwningProvider(ctx, s.attachmentStorage, attachment.StorageURL, "after-sales", nil)
	if err != nil {
		return nil, err
	}
	return &AfterSalesAttachmentFile{
		ReadCloser: object.ReadCloser,
		Filename:   afterSalesAttachmentFilename(*attachment, object.Name),
		MimeType:   afterSalesAttachmentMimeType(*attachment, object.MimeType),
		Size:       object.Size,
	}, nil
}

func IsAfterSalesEvidenceStorageKey(key string) bool {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	return ok && strings.HasPrefix(normalizedKey, "after-sales/")
}

func afterSalesAttachmentFilename(
	attachment aftersales.AfterSalesCaseAttachment,
	fallback string,
) string {
	if value := strings.TrimSpace(attachment.Filename); value != "" {
		return filepath.Base(value)
	}
	if value := strings.TrimSpace(fallback); value != "" {
		return filepath.Base(value)
	}
	return "after-sales-attachment"
}

func afterSalesAttachmentMimeType(
	attachment aftersales.AfterSalesCaseAttachment,
	fallback string,
) string {
	if value := strings.TrimSpace(attachment.ContentType); value != "" {
		return value
	}
	if value := strings.TrimSpace(fallback); value != "" {
		return value
	}
	return "application/octet-stream"
}
