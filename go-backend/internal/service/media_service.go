package service

import (
	"errors"
	"strings"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

var (
	ErrMediaAssetNotFound                 = errors.New("media asset not found")
	ErrUnsupportedMediaType               = errors.New("unsupported media type")
	ErrUnsupportedMediaStatus             = errors.New("unsupported media status")
	ErrUnsupportedVisibility              = errors.New("unsupported media visibility")
	ErrMediaUploadFileRequired            = errors.New("media upload file is required")
	ErrMediaAssetURLUnavailable           = errors.New("media asset url is unavailable")
	ErrMediaStorageUnavailable            = errors.New("media storage unavailable")
	ErrMediaAssetForbidden                = errors.New("media asset is not public")
	ErrMediaUploadIdentityRequired        = errors.New("media upload identity is required")
	ErrMediaAccountStorageQuotaExceeded   = errors.New("media account storage quota exceeded")
	ErrMediaDerivativeGenerationFailed    = errors.New("media derivative generation failed")
	ErrMediaDerivativePresetUnavailable   = errors.New("media derivative preset configuration is unavailable")
	ErrInvalidMediaDerivativePreset       = errors.New("invalid media derivative preset")
	ErrMediaDerivativePresetNotFound      = errors.New("media derivative preset not found")
	ErrMediaDerivativePresetConflict      = errors.New("media derivative preset already exists")
	ErrMediaDerivativePresetInUse         = errors.New("media derivative preset has generated files")
	ErrMediaDerivativePresetProtected     = errors.New("media derivative preset is protected")
	ErrMediaDerivativePresetCodeImmutable = errors.New("media derivative preset code cannot be changed")
	ErrMediaDerivativePresetLimitReached  = errors.New("active image conversion limit reached")
)

// MediaService coordinates media-asset operations. Individual workflows live
// in dedicated files for upload, catalog, access, evidence, references, and
// deletion so each can grow independently.
type MediaService struct {
	repo                     *repository.MediaRepository
	derivativePresets        *repository.MediaDerivativePresetRepository
	derivativeRebuildJobs    *repository.MediaDerivativeRebuildJobRepository
	storage                  storage.StorageService
	settings                 *SettingService
	siteURL                  string
	accountStorageQuotaBytes int64
}

// PublicMediaURLResolver is the boundary used by public response mappers when
// they need to expose a stored media reference.
type PublicMediaURLResolver interface {
	CanonicalPublicMediaURL(reference string) string
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

func (s *MediaService) ConfigureDerivativePresetRepository(repo *repository.MediaDerivativePresetRepository) {
	if s == nil {
		return
	}
	s.derivativePresets = repo
}

func (s *MediaService) ConfigureDerivativeRebuildJobRepository(repo *repository.MediaDerivativeRebuildJobRepository) {
	if s == nil {
		return
	}
	s.derivativeRebuildJobs = repo
}
