package service

import (
	"errors"
	"strings"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

var (
	ErrMediaAssetNotFound               = errors.New("media asset not found")
	ErrUnsupportedMediaType             = errors.New("unsupported media type")
	ErrUnsupportedMediaStatus           = errors.New("unsupported media status")
	ErrUnsupportedVisibility            = errors.New("unsupported media visibility")
	ErrMediaUploadFileRequired          = errors.New("media upload file is required")
	ErrMediaAssetURLUnavailable         = errors.New("media asset url is unavailable")
	ErrMediaStorageUnavailable          = errors.New("media storage unavailable")
	ErrMediaAssetForbidden              = errors.New("media asset is not public")
	ErrMediaUploadIdentityRequired      = errors.New("media upload identity is required")
	ErrMediaAccountStorageQuotaExceeded = errors.New("media account storage quota exceeded")
)

// MediaService coordinates media-asset operations. Individual workflows live
// in dedicated files for upload, catalog, access, evidence, references, and
// deletion so each can grow independently.
type MediaService struct {
	repo                     *repository.MediaRepository
	storage                  storage.StorageService
	settings                 *SettingService
	siteURL                  string
	accountStorageQuotaBytes int64
}

func NewMediaService(
	repo *repository.MediaRepository,
	storageSvc storage.StorageService,
	settingSvc *SettingService,
	siteURL string,
	accountStorageQuotaBytes int64,
) *MediaService {
	return &MediaService{
		repo:                     repo,
		storage:                  storageSvc,
		settings:                 settingSvc,
		siteURL:                  strings.TrimRight(strings.TrimSpace(siteURL), "/"),
		accountStorageQuotaBytes: accountStorageQuotaBytes,
	}
}
