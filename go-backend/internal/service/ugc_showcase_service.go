package service

import (
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UGCShowcaseService struct {
	repo              *repository.UGCShowcaseRepository
	storage           storage.StorageService
	uploadEligibility *UGCShowcaseUploadEligibilityService
	pendingLimit      int
}

func NewUGCShowcaseService(repo *repository.UGCShowcaseRepository, st storage.StorageService) *UGCShowcaseService {
	return &UGCShowcaseService{
		repo:    repo,
		storage: st,
	}
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

	if s.uploadEligibility == nil {
		return nil, ErrShowcaseUploadEligibilityUnavailable
	}
	if _, err := s.uploadEligibility.RequireEligibleOrder(ctx, userID, orderID); err != nil {
		return nil, err
	}

	if s.storage == nil {
		return nil, ErrShowcaseStorageUnavailable
	}

	// 1. 上传图片。待审核图片只保存对象 key，避免 pending 文件的公开 URL 泄露到数据库或后台列表。
	var pendingImageKeys []string
	for _, file := range files {
		key, err := s.uploadPendingShowcaseImage(ctx, file)
		if err != nil {
			deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, pendingImageKeys)
			return nil, fmt.Errorf("failed to upload file %s: %w", file.Filename, err)
		}
		pendingImageKeys = append(pendingImageKeys, key)
	}

	imagesJSON, err := json.Marshal(pendingImageKeys)
	if err != nil {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, pendingImageKeys)
		return nil, fmt.Errorf("failed to encode uploaded images: %w", err)
	}

	item := &ugcshowcase.UGCShowcase{
		UserID:    userID,
		OrderID:   showcaseUploadOrderID(orderID),
		Kind:      ugcshowcase.KindUser,
		Region:    params["region"],
		Location:  params["location"],
		Nickname:  params["nickname"],
		Notes:     params["notes"],
		Images:    imagesJSON,
		Status:    ugcshowcase.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.createPendingSubmission(item); err != nil {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, pendingImageKeys)
		return nil, err
	}

	return item, nil
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

func (s *UGCShowcaseService) createPendingSubmission(item *ugcshowcase.UGCShowcase) error {
	if s == nil || s.repo == nil {
		return ErrShowcaseStorageUnavailable
	}
	if s.pendingLimit <= 0 {
		return s.repo.Create(item)
	}
	return s.repo.WithTransaction(func(repo *repository.UGCShowcaseRepository) error {
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

func deleteUploadedShowcaseImagesBestEffort(ctx context.Context, storageService storage.StorageService, imageReferences []string) {
	if storageService == nil {
		return
	}
	for _, imageReference := range imageReferences {
		if imageReference == "" {
			continue
		}
		_ = storageService.Delete(ctx, imageReference)
	}
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
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if item.Status != ugcshowcase.StatusPending {
		return ErrShowcaseInvalidTransition
	}
	if s.storage == nil {
		return ErrShowcaseStorageUnavailable
	}
	imageReferences, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return err
	}
	publication, err := s.publishShowcaseImages(ctx, imageReferences)
	if err != nil {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, publication.CopiedObjectKeys)
		return err
	}
	imagesJSON, err := json.Marshal(publication.PublishedObjectKeys)
	if err != nil {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, publication.CopiedObjectKeys)
		return fmt.Errorf("failed to encode published images: %w", err)
	}
	updated, err := s.repo.UpdatePendingImagesAndStatus(id, datatypes.JSON(imagesJSON), ugcshowcase.StatusApproved, "")
	if err != nil {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, publication.CopiedObjectKeys)
		return err
	}
	if !updated {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, publication.CopiedObjectKeys)
		return ErrShowcaseInvalidTransition
	}
	deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, publication.PendingSourceReferences)
	return nil
}

func (s *UGCShowcaseService) Reject(ctx context.Context, id uint, reason string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if item.Status != ugcshowcase.StatusPending {
		return ErrShowcaseInvalidTransition
	}
	imageReferences, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return err
	}

	pendingImageReferences := make([]string, 0, len(imageReferences))
	if s.storage != nil {
		for _, imageReference := range imageReferences {
			key, keyErr := s.storage.ObjectKey(imageReference)
			if keyErr == nil && showcaseStorageKeyIsPending(key) {
				pendingImageReferences = append(pendingImageReferences, imageReference)
			}
		}
	}

	updated, err := s.repo.UpdatePendingStatus(id, ugcshowcase.StatusRejected, reason)
	if err != nil {
		return err
	}
	if !updated {
		return ErrShowcaseInvalidTransition
	}

	if len(pendingImageReferences) > 0 {
		deleteUploadedShowcaseImagesBestEffort(ctx, s.storage, pendingImageReferences)
	}
	return nil
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
