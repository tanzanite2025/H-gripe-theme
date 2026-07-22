package service

import (
	"context"
	"mime/multipart"
	"path"
	"strings"
	"tanzanite/internal/domain/media"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
)

type MediaService struct {
	repo    *repository.MediaRepository
	storage storage.StorageService
}

type MediaUploadInput struct {
	File       *multipart.FileHeader
	MediaType  string
	Alt        string
	Caption    string
	UploaderID uint
}

func NewMediaService(repo *repository.MediaRepository, storageSvc storage.StorageService) *MediaService {
	return &MediaService{
		repo:    repo,
		storage: storageSvc,
	}
}

func (s *MediaService) UploadAsset(ctx context.Context, input MediaUploadInput) (*media.MediaAsset, error) {
	url, err := s.storage.Upload(ctx, input.File)
	if err != nil {
		return nil, err
	}

	asset := &media.MediaAsset{
		Filename:         path.Base(url),
		OriginalFilename: input.File.Filename,
		URL:              url,
		StorageKey:       strings.TrimPrefix(path.Base(url), "/"),
		MimeType:         input.File.Header.Get("Content-Type"),
		MediaType:        input.MediaType,
		Size:             input.File.Size,
		Alt:              input.Alt,
		Caption:          input.Caption,
		UploaderID:       input.UploaderID,
		Status:           "active",
		Visibility:       "public",
	}
	if asset.MediaType == "" {
		asset.MediaType = "image"
	}

	if err := s.repo.CreateAsset(asset); err != nil {
		return nil, err
	}

	return asset, nil
}
