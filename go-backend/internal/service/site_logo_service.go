package service

import (
	"context"
	"errors"
	"mime/multipart"
	"path"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

const SiteLogoStoragePrefix = "site-logo"

var (
	ErrSiteLogoUnavailable            = errors.New("site logo service unavailable")
	ErrSiteLogoUploadFileRequired     = errors.New("site logo upload file is required")
	ErrSiteLogoUploadIdentityRequired = errors.New("site logo upload identity is required")
	ErrSiteLogoStorageKeyUnavailable  = errors.New("site logo storage key is unavailable")
)

type SiteLogoService struct {
	repo     *repository.SiteLogoRepository
	storage  storage.StorageService
	siteURL  string
	outbox   *repository.OutboxRepository
	settings *SettingService
}

func (s *SiteLogoService) ConfigureSettingService(settings *SettingService) {
	if s != nil {
		s.settings = settings
	}
}

func (s *SiteLogoService) ConfigureObjectCleanupOutbox(repo *repository.OutboxRepository) {
	if s == nil {
		return
	}
	s.outbox = repo
}

func NewSiteLogoService(repo *repository.SiteLogoRepository, storageSvc storage.StorageService, siteURL string) *SiteLogoService {
	if repo != nil {
		repo.ConfigureTransactionalSiteLogoObjectValidation(func(ctx context.Context, asset *sitelogodomain.Asset) error {
			return validateDedicatedSiteIdentityObjectIdentityAndExistenceBeforeOwnerRegistration(ctx, storageSvc, SiteLogoStoragePrefix, asset.StorageKey, asset.URL)
		})
	}
	return &SiteLogoService{
		repo:    repo,
		storage: storageSvc,
		siteURL: strings.TrimRight(strings.TrimSpace(siteURL), "/"),
	}
}

func (s *SiteLogoService) UploadCurrent(ctx context.Context, file *multipart.FileHeader, uploaderID uint, locales ...string) (*sitelogodomain.Asset, error) {
	if s == nil || s.repo == nil || s.storage == nil || s.settings == nil {
		return nil, ErrSiteLogoUnavailable
	}
	if file == nil {
		return nil, ErrSiteLogoUploadFileRequired
	}
	if uploaderID == 0 {
		return nil, ErrSiteLogoUploadIdentityRequired
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecSiteLogo)); err != nil {
		return nil, err
	}

	contentSHA256, err := contentSHA256FromMultipartFile(file)
	if err != nil {
		return nil, err
	}

	var asset *sitelogodomain.Asset

	locale := "en"
	if len(locales) > 0 && strings.TrimSpace(locales[0]) != "" {
		locale = strings.TrimSpace(locales[0])
	}
	var mirrors []setting.Setting
	_, err = uploadDedicatedImmediateOwnerObjectWithWriteAheadCleanupReceiptAndTransactionalBinding(ctx, s.outbox, s.storage, SiteLogoStoragePrefix, 0, file, func(uploadTx *gorm.DB, url, storageKey string) error {
		asset = &sitelogodomain.Asset{
			Filename:         path.Base(url),
			OriginalFilename: file.Filename,
			URL:              url,
			StorageKey:       storageKey,
			MimeType:         upload.SiteLogoImageContentType,
			Size:             file.Size,
			ContentSHA256:    contentSHA256,
			Width:            upload.SiteLogoImageDimension,
			Height:           upload.SiteLogoImageDimension,
			UploaderID:       uploaderID,
		}

		_, err := s.repo.WithTx(uploadTx).ReplaceCurrent(asset, func(tx *gorm.DB, previous *sitelogodomain.Asset) error {
			if previous != nil && previous.StorageKey != "" && previous.StorageKey != asset.StorageKey {
				if s.outbox == nil {
					return ErrObjectStorageCleanupUnavailable
				}
				event, eventErr := newObjectStorageCleanupEvent(objectCleanupResourceSiteLogo, strconv.FormatUint(uint64(previous.ID), 10), outbox.AggregateTypeSiteLogo, strconv.FormatUint(uint64(previous.ID), 10), []string{previous.StorageKey})
				if eventErr != nil {
					return eventErr
				}
				if err := s.outbox.WithTx(tx).CreateEvent(event); err != nil {
					return err
				}
			}
			var err error
			mirrors, err = s.settings.synchronizeDedicatedSiteIdentityMirrorsFromOwnerTransaction(tx, setting.SiteLogoSettingKey, locale, s.siteURL)
			return err
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	s.settings.invalidateDedicatedSiteIdentityMirrorCachesAfterOwnerCommit(mirrors)
	asset.URL = s.CanonicalPublicURL(asset.URL)
	return asset, nil
}

func (s *SiteLogoService) persistFailedUploadCompensationWithOwnerAndHistoricalReferenceProtection(ctx context.Context, reference string) error {
	return persistFailedDedicatedPublicUploadCompensationBeforeAttemptingOwnerCheckedCleanup(ctx, s.outbox, s.storage, SiteLogoStoragePrefix, objectCleanupResourceSiteLogo, outbox.AggregateTypeSiteLogo, reference, NewObjectStorageCleanupOutboxHandler(nil, s, nil, nil))
}

func (s *SiteLogoService) DeleteCurrent(ctx context.Context, locales ...string) error {
	if s == nil || s.repo == nil || s.storage == nil || s.settings == nil {
		return ErrSiteLogoUnavailable
	}

	locale := "en"
	if len(locales) > 0 && strings.TrimSpace(locales[0]) != "" {
		locale = strings.TrimSpace(locales[0])
	}
	var mirrors []setting.Setting
	_, err := s.repo.WithContext(ctx).DeleteCurrent(func(tx *gorm.DB, previous *sitelogodomain.Asset) error {
		if previous != nil && previous.StorageKey != "" {
			if s.outbox == nil {
				return ErrObjectStorageCleanupUnavailable
			}
			event, eventErr := newObjectStorageCleanupEvent(objectCleanupResourceSiteLogo, strconv.FormatUint(uint64(previous.ID), 10), outbox.AggregateTypeSiteLogo, strconv.FormatUint(uint64(previous.ID), 10), []string{previous.StorageKey})
			if eventErr != nil {
				return eventErr
			}
			if err := s.outbox.WithTx(tx).CreateEvent(event); err != nil {
				return err
			}
		}
		var err error
		mirrors, err = s.settings.synchronizeDedicatedSiteIdentityMirrorsFromOwnerTransaction(tx, setting.SiteLogoSettingKey, locale, s.siteURL)
		return err
	})
	if err != nil {
		return err
	}
	s.settings.invalidateDedicatedSiteIdentityMirrorCachesAfterOwnerCommit(mirrors)
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

func (s *SiteLogoService) CanServePublicLogo(ctx context.Context, key string) (bool, error) {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok || !IsSiteLogoStorageKey(normalizedKey) || s == nil || s.repo == nil {
		return false, nil
	}
	current, err := s.repo.WithContext(ctx).Current()
	if err != nil {
		return false, err
	}
	if current != nil && current.StorageKey == normalizedKey {
		return true, nil
	}
	return s.repo.HasPublishedBusinessConsumerForDedicatedPublicObject(ctx, normalizedKey, &MediaService{storage: s.storage, siteURL: s.siteURL})
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
