package service

import (
	"commerce-platform/internal/domain/showcase"
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

type ShowcaseService struct {
	repo              *repository.ShowcaseRepository
	storage           storage.StorageService
	uploadEligibility *ShowcaseUploadEligibilityService
	pendingLimit      int
}

func NewShowcaseService(repo *repository.ShowcaseRepository, st storage.StorageService) *ShowcaseService {
	return &ShowcaseService{
		repo:    repo,
		storage: st,
	}
}

func (s *ShowcaseService) ConfigureUploadEligibility(eligibility *ShowcaseUploadEligibilityService) {
	if s != nil {
		s.uploadEligibility = eligibility
	}
}

func (s *ShowcaseService) ConfigurePendingSubmissionLimit(limit int) {
	if s != nil {
		s.pendingLimit = limit
	}
}

// UploadPhotos 处理多图上传和买家秀创建
func (s *ShowcaseService) UploadPhotos(ctx context.Context, userID, orderID uint, files []*multipart.FileHeader, params map[string]string) (*showcase.Showcase, error) {
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

	item := &showcase.Showcase{
		UserID:    userID,
		OrderID:   showcaseUploadOrderID(orderID),
		Kind:      showcase.KindUser,
		Region:    params["region"],
		Location:  params["location"],
		Nickname:  params["nickname"],
		Notes:     params["notes"],
		Images:    imagesJSON,
		Status:    showcase.StatusPending,
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

func (s *ShowcaseService) CountPendingSubmissions(userID uint) (int64, error) {
	return s.repo.CountByUserAndStatus(userID, showcase.StatusPending)
}

func (s *ShowcaseService) ValidateUploadOrder(ctx context.Context, userID, orderID uint) error {
	if s == nil || s.uploadEligibility == nil {
		return ErrShowcaseUploadEligibilityUnavailable
	}
	_, err := s.uploadEligibility.RequireEligibleOrder(ctx, userID, orderID)
	return err
}

func (s *ShowcaseService) createPendingSubmission(item *showcase.Showcase) error {
	if s == nil || s.repo == nil {
		return ErrShowcaseStorageUnavailable
	}
	if s.pendingLimit <= 0 {
		return s.repo.Create(item)
	}
	return s.repo.WithTransaction(func(repo *repository.ShowcaseRepository) error {
		if err := repo.LockUserForSubmissionLimit(item.UserID); err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrShowcaseUploadOrderNotEligible
			}
			return err
		}
		pendingCount, err := repo.CountByUserAndStatus(item.UserID, showcase.StatusPending)
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

func (s *ShowcaseService) List(kind string, status string, page int, perPage int) ([]showcase.Showcase, error) {
	offset := (page - 1) * perPage
	return s.repo.List(kind, status, perPage, offset)
}

func (s *ShowcaseService) Count(kind string, status string) (int64, error) {
	return s.repo.Count(kind, status)
}

func (s *ShowcaseService) Get(id uint) (*showcase.Showcase, error) {
	item, err := s.repo.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrShowcaseNotFound
	}
	return item, err
}

func (s *ShowcaseService) ListPublic(kind string, page int, perPage int) ([]showcase.Showcase, error) {
	return s.List(kind, showcase.StatusApproved, page, perPage)
}

func (s *ShowcaseService) Approve(ctx context.Context, id uint) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if item.Status != showcase.StatusPending {
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
	updated, err := s.repo.UpdatePendingImagesAndStatus(id, datatypes.JSON(imagesJSON), showcase.StatusApproved, "")
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

func (s *ShowcaseService) Reject(ctx context.Context, id uint, reason string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if item.Status != showcase.StatusPending {
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

	updated, err := s.repo.UpdatePendingStatus(id, showcase.StatusRejected, reason)
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

func (s *ShowcaseService) AddComment(showcaseID uint, userID uint, author string, content string, location string) (*showcase.Comment, error) {
	comment := &showcase.Comment{
		ShowcaseID: showcaseID,
		UserID:     userID,
		Author:     author,
		Content:    content,
		Location:   location,
		Status:     showcase.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := s.repo.CreateComment(comment)
	return comment, err
}

func (s *ShowcaseService) ListComments(showcaseID uint, page int, perPage int) ([]showcase.Comment, error) {
	offset := (page - 1) * perPage
	return s.repo.ListComments(showcaseID, perPage, offset)
}
