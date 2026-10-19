package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/gorm"
)

const (
	showcasePendingUploadPrefix  = "showcase/pending"
	showcaseApprovedUploadPrefix = "showcase/approved"
	showcaseAdminImageFileTTL    = 10 * time.Minute
)

var (
	ErrShowcaseNotFound           = errors.New("showcase not found")
	ErrShowcaseImageNotFound      = errors.New("showcase image not found")
	ErrShowcaseStorageUnavailable = errors.New("showcase storage unavailable")
	ErrShowcaseImagesInvalid      = errors.New("showcase images are invalid")
	ErrShowcaseInvalidTransition  = errors.New("showcase status transition is invalid")
)

type UGCShowcasePublicImageAccess struct {
	Found   bool
	Allowed bool
}

type UGCShowcaseImageFile struct {
	ReadCloser  io.ReadCloser
	RedirectURL string
	Filename    string
	MimeType    string
	Size        int64
	ModTime     time.Time
}

func (s *UGCShowcaseService) PublicImageAccess(ctx context.Context, key string) (UGCShowcasePublicImageAccess, error) {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok {
		return UGCShowcasePublicImageAccess{Found: true, Allowed: false}, nil
	}
	if showcaseStorageKeyIsPending(normalizedKey) {
		return UGCShowcasePublicImageAccess{Found: true, Allowed: false}, nil
	}
	if s == nil || s.storage == nil {
		if showcaseStorageKeyIsInShowcaseNamespace(normalizedKey) {
			return UGCShowcasePublicImageAccess{Found: true, Allowed: false}, nil
		}
		return UGCShowcasePublicImageAccess{}, nil
	}

	access := UGCShowcasePublicImageAccess{Found: showcaseStorageKeyIsInShowcaseNamespace(normalizedKey)}
	for afterID := uint(0); ; {
		items, err := s.repo.WithContext(ctx).ListUGCShowcaseImageReferenceScanPageAfterID(afterID)
		if err != nil {
			return UGCShowcasePublicImageAccess{}, err
		}
		for _, item := range items {
			matches, err := s.showcaseImageMatchesStorageKey(item, normalizedKey)
			if err != nil {
				return UGCShowcasePublicImageAccess{}, err
			}
			if matches {
				access.Found = true
				if item.Status == ugcshowcasedomain.StatusApproved {
					access.Allowed = true
					return access, nil
				}
			}
		}
		if len(items) < 256 {
			return access, nil
		}
		afterID = items[len(items)-1].ID
	}
}

func (s *UGCShowcaseService) OpenImageFile(ctx context.Context, id uint, imageIndex int) (*UGCShowcaseImageFile, error) {
	return s.openImageFile(ctx, id, imageIndex, false)
}

func (s *UGCShowcaseService) OpenPublicImageFile(ctx context.Context, id uint, imageIndex int) (*UGCShowcaseImageFile, error) {
	return s.openImageFile(ctx, id, imageIndex, true)
}

func (s *UGCShowcaseService) openImageFile(ctx context.Context, id uint, imageIndex int, approvedOnly bool) (*UGCShowcaseImageFile, error) {
	if imageIndex < 0 {
		return nil, ErrShowcaseImageNotFound
	}
	if s == nil || s.storage == nil {
		return nil, ErrShowcaseStorageUnavailable
	}
	item, err := s.repo.WithContext(ctx).GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShowcaseNotFound
		}
		return nil, err
	}
	if approvedOnly && item.Status != ugcshowcasedomain.StatusApproved {
		return nil, ErrShowcaseImageNotFound
	}
	imageURLs, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return nil, err
	}
	if imageIndex >= len(imageURLs) {
		return nil, ErrShowcaseImageNotFound
	}
	key, err := s.resolveUGCShowcaseStoredImageObjectKey(imageURLs[imageIndex])
	if err != nil {
		return nil, ErrShowcaseImageNotFound
	}
	if approvedOnly && !showcaseStorageKeyIsApproved(key) {
		return nil, ErrShowcaseImageNotFound
	}

	if opener, ok := s.storage.(storage.ObjectOpener); ok {
		object, err := opener.Open(ctx, key)
		if err != nil {
			return nil, err
		}
		return &UGCShowcaseImageFile{
			ReadCloser: object.ReadCloser,
			Filename:   object.Name,
			MimeType:   object.MimeType,
			Size:       object.Size,
			ModTime:    object.ModTime,
		}, nil
	}

	if signer, ok := s.storage.(storage.PresignedURLProvider); ok {
		url, err := signer.GetPresignedURL(ctx, key, showcaseAdminImageFileTTL)
		if err != nil {
			return nil, err
		}
		return &UGCShowcaseImageFile{RedirectURL: url, Filename: key}, nil
	}

	return nil, ErrShowcaseStorageUnavailable
}

func (s *UGCShowcaseService) showcaseImageMatchesStorageKey(item ugcshowcasedomain.UGCShowcase, key string) (bool, error) {
	imageURLs, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return false, err
	}
	for _, imageURL := range imageURLs {
		imageKey, err := s.resolveUGCShowcaseStoredImageObjectKey(imageURL)
		if err != nil {
			continue
		}
		if imageKey == key {
			return true, nil
		}
	}
	return false, nil
}

func decodeShowcaseImageURLs(raw []byte) ([]string, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return []string{}, nil
	}
	var imageURLs []string
	if err := json.Unmarshal(raw, &imageURLs); err != nil {
		return nil, ErrShowcaseImagesInvalid
	}
	normalized := make([]string, 0, len(imageURLs))
	for _, imageURL := range imageURLs {
		if trimmed := strings.TrimSpace(imageURL); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized, nil
}

func showcaseStorageKeyIsInShowcaseNamespace(key string) bool {
	return showcaseStorageKeyIsPending(key) || showcaseStorageKeyIsApproved(key)
}

func showcaseStorageKeyIsPending(key string) bool {
	return storageKeyHasPrefix(key, showcasePendingUploadPrefix)
}

func showcaseStorageKeyIsApproved(key string) bool {
	return storageKeyHasPrefix(key, showcaseApprovedUploadPrefix)
}

func storageKeyHasPrefix(key string, prefix string) bool {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok {
		return false
	}
	normalizedPrefix, ok := storage.NormalizeObjectKey(prefix)
	if !ok {
		return false
	}
	return normalizedKey == normalizedPrefix || strings.HasPrefix(normalizedKey, normalizedPrefix+"/")
}

func approvedShowcaseStorageKeyForSource(sourceKey string, publicationID string) (string, error) {
	normalizedSourceKey, ok := storage.NormalizeObjectKey(sourceKey)
	if !ok {
		return "", ErrShowcaseImagesInvalid
	}
	if showcaseStorageKeyIsApproved(normalizedSourceKey) {
		return normalizedSourceKey, nil
	}
	if showcaseStorageKeyIsPending(normalizedSourceKey) {
		suffix := strings.TrimPrefix(normalizedSourceKey, showcasePendingUploadPrefix)
		suffix = strings.TrimPrefix(suffix, "/")
		publicationPrefix, err := storage.JoinObjectKey(showcaseApprovedUploadPrefix, publicationID)
		if err != nil {
			return "", err
		}
		return storage.JoinObjectKey(publicationPrefix, suffix)
	}
	publicationPrefix, err := storage.JoinObjectKey(showcaseApprovedUploadPrefix, publicationID)
	if err != nil {
		return "", err
	}
	return storage.JoinObjectKey(publicationPrefix, normalizedSourceKey)
}
