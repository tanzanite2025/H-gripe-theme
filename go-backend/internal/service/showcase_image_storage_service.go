package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	showcasedomain "commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/pkg/storage"

	"github.com/google/uuid"
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

type ShowcasePublicImageAccess struct {
	Found   bool
	Allowed bool
}

type ShowcaseImageFile struct {
	ReadCloser  io.ReadCloser
	RedirectURL string
	Filename    string
	MimeType    string
	Size        int64
	ModTime     time.Time
}

type showcaseImagePublication struct {
	PublishedObjectKeys     []string
	CopiedObjectKeys        []string
	PendingSourceReferences []string
}

func (s *ShowcaseService) PublicImageAccess(ctx context.Context, key string) (ShowcasePublicImageAccess, error) {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok {
		return ShowcasePublicImageAccess{Found: true, Allowed: false}, nil
	}
	if showcaseStorageKeyIsPending(normalizedKey) {
		return ShowcasePublicImageAccess{Found: true, Allowed: false}, nil
	}
	if s == nil || s.storage == nil {
		if showcaseStorageKeyIsInShowcaseNamespace(normalizedKey) {
			return ShowcasePublicImageAccess{Found: true, Allowed: false}, nil
		}
		return ShowcasePublicImageAccess{}, nil
	}

	items, err := s.repo.ListByImageStorageKeyCandidate(normalizedKey)
	if err != nil {
		return ShowcasePublicImageAccess{}, err
	}
	for _, item := range items {
		matches, err := s.showcaseImageMatchesStorageKey(item, normalizedKey)
		if err != nil {
			return ShowcasePublicImageAccess{}, err
		}
		if matches {
			return ShowcasePublicImageAccess{
				Found:   true,
				Allowed: item.Status == showcasedomain.StatusApproved,
			}, nil
		}
	}

	if showcaseStorageKeyIsInShowcaseNamespace(normalizedKey) {
		return ShowcasePublicImageAccess{Found: true, Allowed: false}, nil
	}
	return ShowcasePublicImageAccess{}, nil
}

func (s *ShowcaseService) OpenImageFile(ctx context.Context, id uint, imageIndex int) (*ShowcaseImageFile, error) {
	return s.openImageFile(ctx, id, imageIndex, false)
}

func (s *ShowcaseService) OpenPublicImageFile(ctx context.Context, id uint, imageIndex int) (*ShowcaseImageFile, error) {
	return s.openImageFile(ctx, id, imageIndex, true)
}

func (s *ShowcaseService) openImageFile(ctx context.Context, id uint, imageIndex int, approvedOnly bool) (*ShowcaseImageFile, error) {
	if imageIndex < 0 {
		return nil, ErrShowcaseImageNotFound
	}
	if s == nil || s.storage == nil {
		return nil, ErrShowcaseStorageUnavailable
	}
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShowcaseNotFound
		}
		return nil, err
	}
	if approvedOnly && item.Status != showcasedomain.StatusApproved {
		return nil, ErrShowcaseImageNotFound
	}
	imageURLs, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return nil, err
	}
	if imageIndex >= len(imageURLs) {
		return nil, ErrShowcaseImageNotFound
	}
	key, err := s.storage.ObjectKey(imageURLs[imageIndex])
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
		return &ShowcaseImageFile{
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
		return &ShowcaseImageFile{RedirectURL: url, Filename: key}, nil
	}

	return nil, ErrShowcaseStorageUnavailable
}

func (s *ShowcaseService) uploadPendingShowcaseImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	privateUploader, ok := s.storage.(storage.PrivateObjectUploader)
	if !ok {
		return "", ErrShowcaseStorageUnavailable
	}
	url, err := privateUploader.UploadWithPrefixPrivate(ctx, file, showcasePendingUploadPrefix)
	if err != nil {
		return "", err
	}
	key, err := s.storage.ObjectKey(url)
	if err != nil {
		_ = s.storage.Delete(ctx, url)
		return "", fmt.Errorf("resolve pending showcase upload key: %w", err)
	}
	if !showcaseStorageKeyIsPending(key) {
		_ = s.storage.Delete(ctx, url)
		return "", fmt.Errorf("pending showcase upload escaped pending namespace")
	}
	return key, nil
}

func (s *ShowcaseService) publishShowcaseImages(ctx context.Context, imageReferences []string) (showcaseImagePublication, error) {
	publication := showcaseImagePublication{
		PublishedObjectKeys:     make([]string, 0, len(imageReferences)),
		CopiedObjectKeys:        make([]string, 0, len(imageReferences)),
		PendingSourceReferences: make([]string, 0, len(imageReferences)),
	}
	publicationID := uuid.NewString()

	for _, imageReference := range imageReferences {
		key, err := s.storage.ObjectKey(imageReference)
		if err != nil {
			return publication, fmt.Errorf("%w: %v", ErrShowcaseImagesInvalid, err)
		}
		if showcaseStorageKeyIsApproved(key) {
			publication.PublishedObjectKeys = append(publication.PublishedObjectKeys, key)
			continue
		}

		destinationKey, err := approvedShowcaseStorageKeyForSource(key, publicationID)
		if err != nil {
			return publication, err
		}
		if err := s.storage.CopyObject(ctx, key, destinationKey); err != nil {
			return publication, fmt.Errorf("publish showcase image: %w", err)
		}

		publication.PublishedObjectKeys = append(publication.PublishedObjectKeys, destinationKey)
		publication.CopiedObjectKeys = append(publication.CopiedObjectKeys, destinationKey)
		if showcaseStorageKeyIsPending(key) {
			publication.PendingSourceReferences = append(publication.PendingSourceReferences, imageReference)
		}
	}

	return publication, nil
}

func (s *ShowcaseService) showcaseImageMatchesStorageKey(item showcasedomain.Showcase, key string) (bool, error) {
	imageURLs, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return false, err
	}
	for _, imageURL := range imageURLs {
		imageKey, err := s.storage.ObjectKey(imageURL)
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
