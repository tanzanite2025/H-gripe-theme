package service

import (
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type UGCShowcaseService struct {
	repo              *repository.UGCShowcaseRepository
	storage           storage.StorageService
	outbox            *repository.OutboxRepository
	uploadEligibility *UGCShowcaseUploadEligibilityService
	pendingLimit      int
	siteURL           string
}

func (s *UGCShowcaseService) ConfigureObjectCleanupOutbox(repo *repository.OutboxRepository) {
	if s == nil {
		return
	}
	s.outbox = repo
}

func NewUGCShowcaseService(repo *repository.UGCShowcaseRepository, st storage.StorageService, siteURLs ...string) *UGCShowcaseService {
	service := &UGCShowcaseService{
		repo:    repo,
		storage: st,
	}
	if len(siteURLs) > 0 {
		service.siteURL = siteURLs[0]
	}
	if repo != nil {
		repo.ConfigureTransactionalUGCShowcaseImageOwnershipValidation(service.validateUGCShowcaseImageBindingsAfterLifecycleLock, service.resolveUGCShowcaseStoredImageObjectKey)
	}
	return service
}

func (s *UGCShowcaseService) ConfigureUploadEligibility(eligibility *UGCShowcaseUploadEligibilityService) {
	if s != nil {
		s.uploadEligibility = eligibility
	}
}

func (s *UGCShowcaseService) ConfigurePendingSubmissionLimit(limit int) {
	if s != nil {
		s.pendingLimit = limit
	}
}

// UploadPhotos 处理多图上传和买家秀创建
func (s *UGCShowcaseService) UploadPhotos(ctx context.Context, userID, orderID uint, files []*multipart.FileHeader, params map[string]string) (*ugcshowcase.UGCShowcase, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}
	if s.repo != nil {
		if err := s.repo.RequireIndependentUGCShowcaseProviderOperationTransaction(); err != nil {
			return nil, err
		}
	}
	if s.outbox != nil {
		if err := s.outbox.RequireIndependentUGCShowcaseObjectWriteReceiptTransaction(); err != nil {
			return nil, err
		}
	}

	if s.uploadEligibility == nil {
		return nil, ErrShowcaseUploadEligibilityUnavailable
	}
	if _, err := s.uploadEligibility.RequireEligibleOrder(ctx, userID, orderID); err != nil {
		return nil, err
	}

	if s.storage == nil || s.outbox == nil || s.repo == nil {
		return nil, ErrShowcaseStorageUnavailable
	}

	return s.uploadPendingShowcaseImagesWithPrecommittedKeysAndAtomicSubmission(ctx, userID, orderID, files, params)
}

func showcaseUploadOrderID(orderID uint) *uint {
	return &orderID
}

func (s *UGCShowcaseService) CountPendingSubmissions(userID uint) (int64, error) {
	return s.repo.CountByUserAndStatus(userID, ugcshowcase.StatusPending)
}

func (s *UGCShowcaseService) ValidateUploadOrder(ctx context.Context, userID, orderID uint) error {
	if s == nil || s.uploadEligibility == nil {
		return ErrShowcaseUploadEligibilityUnavailable
	}
	_, err := s.uploadEligibility.RequireEligibleOrder(ctx, userID, orderID)
	return err
}

func (s *UGCShowcaseService) createPendingSubmission(ctx context.Context, item *ugcshowcase.UGCShowcase) error {
	if s == nil || s.repo == nil {
		return ErrShowcaseStorageUnavailable
	}
	if s.pendingLimit <= 0 {
		return s.repo.WithContext(ctx).Create(item)
	}
	return s.repo.WithContext(ctx).WithTransaction(func(repo *repository.UGCShowcaseRepository, _ *gorm.DB) error {
		if err := repo.LockUserForSubmissionLimit(item.UserID); err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrShowcaseUploadOrderNotEligible
			}
			return err
		}
		pendingCount, err := repo.CountByUserAndStatus(item.UserID, ugcshowcase.StatusPending)
		if err != nil {
			return err
		}
		if pendingCount >= int64(s.pendingLimit) {
			return ErrShowcaseUploadPendingLimitExceeded
		}
		return repo.Create(item)
	})
}

func (s *UGCShowcaseService) List(kind string, status string, page int, perPage int) ([]ugcshowcase.UGCShowcase, error) {
	offset := (page - 1) * perPage
	return s.repo.List(kind, status, perPage, offset)
}

func (s *UGCShowcaseService) Count(kind string, status string) (int64, error) {
	return s.repo.Count(kind, status)
}

func (s *UGCShowcaseService) Get(id uint) (*ugcshowcase.UGCShowcase, error) {
	item, err := s.repo.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrShowcaseNotFound
	}
	return item, err
}

func (s *UGCShowcaseService) ListPublic(kind string, page int, perPage int) ([]ugcshowcase.UGCShowcase, error) {
	return s.List(kind, ugcshowcase.StatusApproved, page, perPage)
}

func (s *UGCShowcaseService) Approve(ctx context.Context, id uint) error {
	if s == nil || s.repo == nil || s.storage == nil || s.outbox == nil {
		return ErrShowcaseStorageUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return s.approveUGCShowcaseWithPrecommittedCopyDestinationsAndAtomicPublication(ctx, id)
}

func (s *UGCShowcaseService) Reject(ctx context.Context, id uint, reason string) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrShowcaseStorageUnavailable
	}
	return s.repo.WithContext(ctx).WithTransaction(func(repo *repository.UGCShowcaseRepository, tx *gorm.DB) error {
		item, err := repo.FindUGCShowcaseByIDForUpdateWithinLifecycleTransaction(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShowcaseNotFound
		}
		if err != nil {
			return err
		}
		if item.Status != ugcshowcase.StatusPending {
			return ErrShowcaseInvalidTransition
		}
		references, err := decodeShowcaseImageURLs(item.Images)
		if err != nil {
			return err
		}
		updated, err := repo.UpdatePendingStatus(id, ugcshowcase.StatusRejected, reason)
		if err != nil {
			return err
		}
		if !updated {
			return ErrShowcaseInvalidTransition
		}
		return s.enqueueShowcaseCleanup(tx, id, s.showcasePendingCleanupKeys(references))
	})
}

func (s *UGCShowcaseService) showcasePendingCleanupKeys(references []string) []string {
	if s == nil || s.storage == nil {
		return nil
	}
	return s.canonicalPendingUGCShowcaseCleanupKeys(references)
}

func (s *UGCShowcaseService) enqueueShowcaseCleanup(tx *gorm.DB, id uint, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	if s == nil || s.outbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	resourceID := strconv.FormatUint(uint64(id), 10)
	event, err := newObjectStorageCleanupEvent(
		objectCleanupResourceUGCShowcase,
		resourceID,
		outbox.AggregateTypeUGCShowcase,
		resourceID,
		keys,
	)
	if err != nil {
		return err
	}
	return s.outbox.WithTx(tx).CreateEvent(event)
}

func (s *UGCShowcaseService) AddComment(showcaseID uint, userID uint, author string, content string, location string) (*ugcshowcase.UGCShowcaseComment, error) {
	comment := &ugcshowcase.UGCShowcaseComment{
		ShowcaseID: showcaseID,
		UserID:     userID,
		Author:     author,
		Content:    content,
		Location:   location,
		Status:     ugcshowcase.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := s.repo.CreateComment(comment)
	return comment, err
}

func (s *UGCShowcaseService) ListComments(showcaseID uint, page int, perPage int) ([]ugcshowcase.UGCShowcaseComment, error) {
	offset := (page - 1) * perPage
	return s.repo.ListComments(showcaseID, perPage, offset)
}
