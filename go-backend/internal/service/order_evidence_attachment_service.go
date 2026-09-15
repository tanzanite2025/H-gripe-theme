package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

var (
	ErrOrderEvidenceAttachmentItemNotFound = errors.New("order evidence attachment item not found")
	ErrOrderEvidenceAttachmentOwnership    = errors.New("order evidence attachment does not belong to the evidence order")
	ErrOrderEvidenceAttachmentUnavailable  = errors.New("order evidence attachment service is unavailable")
)

// OrderEvidenceAttachmentService owns the boundary between validated
// multipart uploads and immutable evidence attachment references.
type OrderEvidenceAttachmentService struct {
	txManager    *repository.TxManager
	evidenceRepo *repository.OrderEvidenceRepository
	storage      storage.StorageService
}

type OrderEvidenceAttachmentReferenceInput struct {
	EvidenceItemID   uint
	StorageKey       string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	SHA256           string
	UploadedBy       uint
}

func NewOrderEvidenceAttachmentService() *OrderEvidenceAttachmentService {
	return &OrderEvidenceAttachmentService{}
}

func NewConfiguredOrderEvidenceAttachmentService(
	txManager *repository.TxManager,
	evidenceRepo *repository.OrderEvidenceRepository,
	storageService storage.StorageService,
) *OrderEvidenceAttachmentService {
	service := NewOrderEvidenceAttachmentService()
	service.ConfigureDependencies(txManager, evidenceRepo, storageService)
	return service
}

func (s *OrderEvidenceAttachmentService) ConfigureDependencies(
	txManager *repository.TxManager,
	evidenceRepo *repository.OrderEvidenceRepository,
	storageService storage.StorageService,
) {
	if s == nil {
		return
	}
	s.txManager = txManager
	s.evidenceRepo = evidenceRepo
	s.storage = storageService
}

func (s *OrderEvidenceAttachmentService) Upload(
	ctx context.Context,
	orderID uint,
	itemID uint,
	file *multipart.FileHeader,
	uploadedBy uint,
) (*orderevidence.OrderEvidenceAttachment, error) {
	if s == nil || s.txManager == nil || s.evidenceRepo == nil || s.storage == nil {
		return nil, ErrOrderEvidenceAttachmentUnavailable
	}
	if orderID == 0 || itemID == 0 || uploadedBy == 0 {
		return nil, errors.New("order id, evidence item id, and uploader are required")
	}
	if err := upload.ValidateFile(file, upload.WarrantyImageRule.FileRule); err != nil {
		return nil, err
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

	sha256Value, sizeBytes, err := hashMultipartFile(file)
	if err != nil {
		return nil, fmt.Errorf("hash order evidence attachment: %w", err)
	}
	mimeType, err := upload.DetectContentType(file)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("order-evidence/%d/%d", orderID, itemID)
	privateUploader, ok := s.storage.(storage.PrivateObjectUploader)
	if !ok {
		return nil, fmt.Errorf("%w: private evidence storage is required", ErrOrderEvidenceAttachmentUnavailable)
	}
	referenceURL, err := privateUploader.UploadWithPrefixPrivate(ctx, file, prefix)
	if err != nil {
		return nil, fmt.Errorf("upload order evidence attachment: %w", err)
	}

	storageKey, err := s.storage.ObjectKey(referenceURL)
	if err != nil {
		_ = s.storage.Delete(ctx, referenceURL)
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
	err = s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var registerErr error
		attachment, registerErr = s.RegisterReference(repos, attachmentInput)
		return registerErr
	})
	if err != nil {
		_ = s.storage.Delete(ctx, referenceURL)
		return nil, err
	}
	return attachment, nil
}

func (s *OrderEvidenceAttachmentService) Open(
	ctx context.Context,
	orderID uint,
	itemID uint,
	attachmentID uint,
) (*orderevidence.OrderEvidenceAttachment, *storage.StoredObject, error) {
	if s == nil || s.evidenceRepo == nil || s.storage == nil {
		return nil, nil, ErrOrderEvidenceAttachmentUnavailable
	}
	attachment, err := s.evidenceRepo.FindAttachmentByIDForOrder(orderID, itemID, attachmentID)
	if repository.IsRecordNotFound(err) {
		return nil, nil, ErrOrderEvidenceAttachmentItemNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	opener, ok := s.storage.(storage.ObjectOpener)
	if !ok {
		return nil, nil, ErrOrderEvidenceAttachmentUnavailable
	}
	object, err := opener.Open(ctx, attachment.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("open order evidence attachment: %w", err)
	}
	return attachment, object, nil
}

func (s *OrderEvidenceAttachmentService) Delete(
	ctx context.Context,
	orderID uint,
	itemID uint,
	attachmentID uint,
) error {
	_ = ctx
	if s == nil || s.txManager == nil || s.evidenceRepo == nil {
		return ErrOrderEvidenceAttachmentUnavailable
	}
	if orderID == 0 || itemID == 0 || attachmentID == 0 {
		return errors.New("order id, evidence item id, and attachment id are required")
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.OrderEvidence == nil {
			return ErrOrderEvidenceAttachmentUnavailable
		}
		item, pkg, err := repos.OrderEvidence.FindItemAndPackageByIDForOrder(itemID, orderID)
		if repository.IsRecordNotFound(err) {
			return ErrOrderEvidenceAttachmentItemNotFound
		}
		if err != nil {
			return err
		}
		if err := pkg.EnsureMutable(); err != nil {
			return err
		}
		attachment, err := repos.OrderEvidence.FindAttachmentByIDForOrder(orderID, item.ID, attachmentID)
		if repository.IsRecordNotFound(err) {
			return ErrOrderEvidenceAttachmentItemNotFound
		}
		if err != nil {
			return err
		}
		if attachment.EvidenceItemID != item.ID {
			return ErrOrderEvidenceAttachmentItemNotFound
		}
		// The storage object remains immutable and is intentionally retained.
		return repos.OrderEvidence.DeleteAttachment(attachment.ID)
	})
}

func (s *OrderEvidenceAttachmentService) RegisterReference(
	repos repository.TxRepositories,
	input OrderEvidenceAttachmentReferenceInput,
) (*orderevidence.OrderEvidenceAttachment, error) {
	if s == nil || repos.OrderEvidence == nil {
		return nil, ErrOrderEvidenceStoreUnavailable
	}
	item, pkg, err := repos.OrderEvidence.FindItemAndPackageByID(input.EvidenceItemID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidenceAttachmentItemNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := pkg.EnsureMutable(); err != nil {
		return nil, err
	}

	normalizedKey, err := orderevidence.NormalizeEvidenceStorageKey(input.StorageKey)
	if err != nil {
		return nil, err
	}
	expectedPrefix := fmt.Sprintf("order-evidence/%d/", item.OrderID)
	if !strings.HasPrefix(normalizedKey, expectedPrefix) {
		return nil, ErrOrderEvidenceAttachmentOwnership
	}

	attachment := &orderevidence.OrderEvidenceAttachment{
		EvidenceItemID:   input.EvidenceItemID,
		StorageKey:       normalizedKey,
		OriginalFilename: input.OriginalFilename,
		MimeType:         input.MimeType,
		SizeBytes:        input.SizeBytes,
		SHA256:           input.SHA256,
		UploadedBy:       input.UploadedBy,
	}
	if err := attachment.Validate(); err != nil {
		return nil, err
	}
	if err := repos.OrderEvidence.CreateAttachment(attachment); err != nil {
		return nil, fmt.Errorf("save order evidence attachment reference: %w", err)
	}
	return attachment, nil
}

func hashMultipartFile(file *multipart.FileHeader) (string, int64, error) {
	if file == nil {
		return "", 0, errors.New("file is required")
	}
	src, err := file.Open()
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = src.Close() }()

	hasher := sha256.New()
	sizeBytes, err := io.Copy(hasher, src)
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), sizeBytes, nil
}
