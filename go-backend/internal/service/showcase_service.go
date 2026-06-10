package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"tanzanite/internal/domain/showcase"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"time"
)

type ShowcaseService struct {
	repo    *repository.ShowcaseRepository
	storage storage.StorageService
}

func NewShowcaseService(repo *repository.ShowcaseRepository, st storage.StorageService) *ShowcaseService {
	return &ShowcaseService{
		repo:    repo,
		storage: st,
	}
}

// UploadPhotos 处理多图上传和买家秀创建
func (s *ShowcaseService) UploadPhotos(ctx context.Context, userID uint, files []*multipart.FileHeader, params map[string]string) (*showcase.Showcase, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// 1. 上传图片
	var imageUrls []string
	for _, file := range files {
		url, err := s.storage.Upload(ctx, file)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %s: %w", file.Filename, err)
		}
		imageUrls = append(imageUrls, url)
	}

	imagesJSON, _ := json.Marshal(imageUrls)

	// 2. 构造对象
	item := &showcase.Showcase{
		UserID:    userID,
		Kind:      showcase.KindUser,
		Region:    params["region"],
		Location:  params["location"],
		Nickname:  params["nickname"],
		BikeModel: params["bike_model"],
		Notes:     params["notes"],
		Images:    imagesJSON,
		Status:    showcase.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 3. 保存到库
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ShowcaseService) List(kind string, status string, page int, perPage int) ([]showcase.Showcase, error) {
	offset := (page - 1) * perPage
	return s.repo.List(kind, status, perPage, offset)
}

func (s *ShowcaseService) Approve(id uint) error {
	return s.repo.UpdateStatus(id, showcase.StatusApproved, "")
}

func (s *ShowcaseService) Reject(id uint, reason string) error {
	return s.repo.UpdateStatus(id, showcase.StatusRejected, reason)
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
