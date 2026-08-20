package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

const SiteLogoStoragePrefix = "site-logo"

var (
	ErrSiteLogoUnavailable            = errors.New("site logo service unavailable")
	ErrSiteLogoUploadFileRequired     = errors.New("site logo upload file is required")
	ErrSiteLogoUploadIdentityRequired = errors.New("site logo upload identity is required")
	ErrSiteLogoStorageKeyUnavailable  = errors.New("site logo storage key is unavailable")
	ErrSiteLogoPreviousDestroyFailed  = errors.New("previous site logo could not be destroyed")
	ErrSiteLogoCurrentDestroyFailed   = errors.New("current site logo could not be destroyed")
)

type SiteLogoService struct {
	repo    *repository.SiteLogoRepository
	storage storage.StorageService
	siteURL string
}

func NewSiteLogoService(repo *repository.SiteLogoRepository, storageSvc storage.StorageService, siteURL string) *SiteLogoService {
	return &SiteLogoService{
		repo:    repo,
		storage: storageSvc,
		siteURL: strings.TrimRight(strings.TrimSpace(siteURL), "/"),
	}
}

func (s *SiteLogoService) UploadCurrent(ctx context.Context, file *multipart.FileHeader, uploaderID uint) (*sitelogodomain.Asset, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return nil, ErrSiteLogoUnavailable
	}
	if file == nil {
		return nil, ErrSiteLogoUploadFileRequired
	}
	if uploaderID == 0 {
		return nil, ErrSiteLogoUploadIdentityRequired
	}
	if err := upload.ValidateSVGFile(file, upload.SiteLogoSVGRule); err != nil {
		return nil, err
	}

	contentSHA256, err := contentSHA256FromMultipartFile(file)
	if err != nil {
		return nil, err
	}

	url, err := s.storage.UploadWithPrefix(ctx, file, SiteLogoStoragePrefix)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(url) == "" {
		return nil, ErrSiteLogoStorageKeyUnavailable
	}

	storageKey, err := s.storage.ObjectKey(url)
	if err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, fmt.Errorf("%w: %v", ErrSiteLogoStorageKeyUnavailable, err)
	}
	if !IsSiteLogoStorageKey(storageKey) {
		_ = s.storage.Delete(ctx, url)
		return nil, ErrSiteLogoStorageKeyUnavailable
	}

	asset := &sitelogodomain.Asset{
		Filename:         path.Base(url),
		OriginalFilename: file.Filename,
		URL:              url,
		StorageKey:       storageKey,
		MimeType:         "image/svg+xml",
		Size:             file.Size,
		ContentSHA256:    contentSHA256,
		Width:            upload.SiteLogoSVGDimension,
		Height:           upload.SiteLogoSVGDimension,
		UploaderID:       uploaderID,
	}

	previous, err := s.repo.ReplaceCurrent(asset)
	if err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}
	if err := s.destroyPrevious(ctx, previous, asset.StorageKey); err != nil {
		return asset, err
	}

	asset.URL = s.CanonicalPublicURL(asset.URL)
	return asset, nil
}

func (s *SiteLogoService) DeleteCurrent(ctx context.Context) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrSiteLogoUnavailable
	}

	previous, err := s.repo.DeleteCurrent()
	if err != nil {
		return err
	}
	if previous == nil || strings.TrimSpace(previous.URL) == "" {
		return nil
	}
	if err := s.storage.Delete(ctx, previous.URL); err != nil {
		return fmt.Errorf("%w: %v", ErrSiteLogoCurrentDestroyFailed, err)
	}
	return nil
}

func (s *SiteLogoService) CurrentPublicURL() string {
	current, ok := s.current()
	if !ok || strings.TrimSpace(current.StorageKey) == "" {
		return ""
	}
	return s.publicURLForKey(current.StorageKey)
}

func (s *SiteLogoService) PublicDimensions(reference string) (int, int, bool) {
	key, ok := s.objectKey(reference)
	if !ok || !IsSiteLogoStorageKey(key) {
		return 0, 0, false
	}
	current, ok := s.current()
	if !ok || current.StorageKey != key || current.Width <= 0 || current.Height <= 0 {
		return 0, 0, false
	}
	return current.Width, current.Height, true
}

func (s *SiteLogoService) CanonicalPublicURL(reference string) string {
	value := strings.TrimSpace(reference)
	if value == "" {
		return ""
	}
	key, ok := s.objectKey(value)
	if !ok || !IsSiteLogoStorageKey(key) {
		return value
	}
	return s.publicURLForKey(key)
}

func (s *SiteLogoService) publicURLForKey(key string) string {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok || !IsSiteLogoStorageKey(normalizedKey) {
		return ""
	}
	if s.siteURL == "" {
		return "/uploads/" + normalizedKey
	}
	return s.siteURL + "/uploads/" + normalizedKey
}

func (s *SiteLogoService) CanServePublicLogo(_ context.Context, key string) (bool, error) {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok || !IsSiteLogoStorageKey(normalizedKey) {
		return false, nil
	}
	current, ok := s.current()
	if !ok {
		return false, nil
	}
	return current.StorageKey == normalizedKey, nil
}

func (s *SiteLogoService) destroyPrevious(ctx context.Context, previous *sitelogodomain.Asset, currentStorageKey string) error {
	if previous == nil || strings.TrimSpace(previous.StorageKey) == "" || previous.StorageKey == currentStorageKey {
		return nil
	}
	if err := s.storage.Delete(ctx, previous.URL); err != nil {
		return fmt.Errorf("%w: %v", ErrSiteLogoPreviousDestroyFailed, err)
	}
	return nil
}

func (s *SiteLogoService) current() (*sitelogodomain.Asset, bool) {
	if s == nil || s.repo == nil {
		return nil, false
	}
	current, err := s.repo.Current()
	if err != nil || current == nil {
		return nil, false
	}
	return current, true
}

func (s *SiteLogoService) objectKey(reference string) (string, bool) {
	value := strings.TrimSpace(reference)
	if value == "" {
		return "", false
	}
	if s != nil && s.storage != nil {
		if key, err := s.storage.ObjectKey(value); err == nil {
			return key, true
		}
	}
	if s != nil && s.siteURL != "" {
		if key, ok := storage.ObjectKeyFromReference(value, s.siteURL); ok {
			return key, true
		}
	}
	return storage.ObjectKeyFromReference(value, "")
}

func IsSiteLogoStorageKey(key string) bool {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	return ok && (normalizedKey == SiteLogoStoragePrefix || strings.HasPrefix(normalizedKey, SiteLogoStoragePrefix+"/"))
}
