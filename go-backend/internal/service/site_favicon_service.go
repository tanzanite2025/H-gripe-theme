package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	settingdomain "commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const SiteFaviconStoragePrefix = "site-favicon"

var (
	ErrSiteFaviconUnavailable            = errors.New("site favicon service unavailable")
	ErrSiteFaviconUploadFileRequired     = errors.New("site favicon upload file is required")
	ErrSiteFaviconUploadIdentityRequired = errors.New("site favicon upload identity is required")
	ErrSiteFaviconStorageKeyUnavailable  = errors.New("site favicon storage key is unavailable")
)

type SiteFaviconService struct {
	repo     *repository.SiteFaviconRepository
	storage  storage.StorageService
	siteURL  string
	outbox   *repository.OutboxRepository
	settings *SettingService
}

func NewSiteFaviconService(repo *repository.SiteFaviconRepository, storageSvc storage.StorageService, siteURL string) *SiteFaviconService {
	return &SiteFaviconService{repo: repo, storage: storageSvc, siteURL: strings.TrimRight(strings.TrimSpace(siteURL), "/")}
}

func (s *SiteFaviconService) ConfigureObjectCleanupOutbox(repo *repository.OutboxRepository) {
	if s != nil {
		s.outbox = repo
	}
}

func (s *SiteFaviconService) ConfigureSettingService(settings *SettingService) {
	if s != nil {
		s.settings = settings
	}
}

func (s *SiteFaviconService) UploadCurrent(ctx context.Context, file *multipart.FileHeader, uploaderID uint, locales ...string) (*sitefavicondomain.Asset, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return nil, ErrSiteFaviconUnavailable
	}
	if file == nil {
		return nil, ErrSiteFaviconUploadFileRequired
	}
	if uploaderID == 0 {
		return nil, ErrSiteFaviconUploadIdentityRequired
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecSiteFavicon)); err != nil {
		return nil, err
	}
	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return nil, err
	}
	contentSHA256, err := contentSHA256FromMultipartFile(file)
	if err != nil {
		return nil, err
	}
	url, err := s.storage.UploadWithPrefix(ctx, file, SiteFaviconStoragePrefix)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(url) == "" {
		return nil, ErrSiteFaviconStorageKeyUnavailable
	}
	storageKey, err := s.storage.ObjectKey(url)
	if err != nil || !IsSiteFaviconStorageKey(storageKey) {
		_ = s.storage.Delete(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSiteFaviconStorageKeyUnavailable, err)
		}
		return nil, ErrSiteFaviconStorageKeyUnavailable
	}
	asset := &sitefavicondomain.Asset{
		Filename: path.Base(url), OriginalFilename: file.Filename, URL: url, StorageKey: storageKey,
		MimeType: strings.TrimSpace(file.Header.Get("Content-Type")), Size: file.Size, ContentSHA256: contentSHA256,
		Width: width, Height: height, UploaderID: uploaderID,
	}
	if asset.MimeType == "" {
		asset.MimeType = "application/octet-stream"
	}
	locale := "en"
	if len(locales) > 0 && strings.TrimSpace(locales[0]) != "" {
		locale = strings.TrimSpace(locales[0])
	}
	err = s.repo.ReplaceCurrent(asset, func(tx *gorm.DB, previous *sitefavicondomain.Asset) error {
		if previous != nil && previous.StorageKey != "" && previous.StorageKey != asset.StorageKey {
			if s.outbox == nil {
				return ErrObjectStorageCleanupUnavailable
			}
			event, eventErr := newObjectStorageCleanupEvent(objectCleanupResourceSiteFavicon, strconv.FormatUint(uint64(previous.ID), 10), outbox.AggregateTypeSiteFavicon, strconv.FormatUint(uint64(previous.ID), 10), []string{previous.StorageKey})
			if eventErr != nil {
				return eventErr
			}
			if err := s.outbox.WithTx(tx).CreateEvent(event); err != nil {
				return err
			}
		}
		return s.persistSettingTx(tx, s.publicURLForKey(asset.StorageKey), locale)
	})
	if err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}
	asset.URL = s.publicURLForKey(asset.StorageKey)
	return asset, nil
}

func (s *SiteFaviconService) DeleteCurrent(ctx context.Context, locales ...string) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrSiteFaviconUnavailable
	}
	locale := "en"
	if len(locales) > 0 && strings.TrimSpace(locales[0]) != "" {
		locale = strings.TrimSpace(locales[0])
	}
	if err := s.repo.DeleteCurrent(func(tx *gorm.DB, previous *sitefavicondomain.Asset) error {
		if previous != nil && previous.StorageKey != "" {
			if s.outbox == nil {
				return ErrObjectStorageCleanupUnavailable
			}
			event, eventErr := newObjectStorageCleanupEvent(objectCleanupResourceSiteFavicon, strconv.FormatUint(uint64(previous.ID), 10), outbox.AggregateTypeSiteFavicon, strconv.FormatUint(uint64(previous.ID), 10), []string{previous.StorageKey})
			if eventErr != nil {
				return eventErr
			}
			if err := s.outbox.WithTx(tx).CreateEvent(event); err != nil {
				return err
			}
		}
		return s.persistSettingTx(tx, "", locale)
	}); err != nil {
		return err
	}
	return nil
}

func (s *SiteFaviconService) persistSettingTx(tx *gorm.DB, value, locale string) error {
	if s.settings == nil {
		return nil
	}
	return s.settings.BatchSetTx(tx, []settingdomain.Setting{{
		Key:         "site_favicon",
		Value:       value,
		Type:        "string",
		Group:       "site",
		Locale:      locale,
		IsPublic:    true,
		Description: "Browser favicon URL",
	}})
}

func (s *SiteFaviconService) CurrentPublicURL() string {
	current, ok := s.current()
	if !ok {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if exists, checked := s.objectExists(ctx, current.StorageKey); checked && !exists {
		return ""
	}
	return s.publicURLForKey(current.StorageKey)
}

func (s *SiteFaviconService) CanonicalPublicURL(reference string) string {
	key, ok := s.objectKey(reference)
	if !ok || !IsSiteFaviconStorageKey(key) {
		return strings.TrimSpace(reference)
	}
	return s.publicURLForKey(key)
}

func (s *SiteFaviconService) CanServePublicFavicon(ctx context.Context, key string) (bool, error) {
	normalized, ok := storage.NormalizeObjectKey(key)
	if !ok || !IsSiteFaviconStorageKey(normalized) {
		return false, nil
	}
	current, ok := s.current()
	if !ok || current.StorageKey != normalized {
		return false, nil
	}
	if exists, checked := s.objectExists(ctx, normalized); checked && !exists {
		return false, nil
	}
	return true, nil
}

// objectExists verifies the physical object when the configured adapter
// exposes the read capability. Test doubles without that capability remain
// permissive so URL-focused tests do not need to emulate a storage backend.
func (s *SiteFaviconService) objectExists(ctx context.Context, key string) (exists, checked bool) {
	if s == nil || s.storage == nil {
		return false, false
	}
	opener, ok := s.storage.(storage.ObjectOpener)
	if !ok {
		return true, false
	}
	object, err := opener.Open(ctx, key)
	if err != nil || object == nil {
		return false, true
	}
	if object.ReadCloser != nil {
		_ = object.ReadCloser.Close()
	}
	return true, true
}

func (s *SiteFaviconService) current() (*sitefavicondomain.Asset, bool) {
	if s == nil || s.repo == nil {
		return nil, false
	}
	asset, err := s.repo.Current()
	return asset, err == nil && asset != nil && strings.TrimSpace(asset.StorageKey) != ""
}

func (s *SiteFaviconService) publicURLForKey(key string) string {
	normalized, ok := storage.NormalizeObjectKey(key)
	if !ok || !IsSiteFaviconStorageKey(normalized) {
		return ""
	}
	if s.siteURL == "" {
		return "/uploads/" + normalized
	}
	return s.siteURL + "/uploads/" + normalized
}

func (s *SiteFaviconService) objectKey(reference string) (string, bool) {
	if s != nil && s.storage != nil {
		if key, err := s.storage.ObjectKey(reference); err == nil {
			return key, true
		}
	}
	return storage.ObjectKeyFromReference(reference, s.siteURL)
}

func IsSiteFaviconStorageKey(key string) bool {
	normalized, ok := storage.NormalizeObjectKey(key)
	return ok && (normalized == SiteFaviconStoragePrefix || strings.HasPrefix(normalized, SiteFaviconStoragePrefix+"/"))
}
