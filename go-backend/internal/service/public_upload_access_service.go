package service

import (
	"context"

	"commerce-platform/internal/pkg/storage"
)

type PublicUploadAccessService struct {
	media    *MediaService
	showcase *ShowcaseService
}

func NewPublicUploadAccessService(mediaService *MediaService, showcaseService *ShowcaseService) *PublicUploadAccessService {
	return &PublicUploadAccessService{
		media:    mediaService,
		showcase: showcaseService,
	}
}

func (s *PublicUploadAccessService) CanServePublicUpload(ctx context.Context, key string) (bool, error) {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok {
		return false, nil
	}
	if showcaseStorageKeyIsPending(normalizedKey) {
		return false, nil
	}

	// Showcase owns its namespace. Resolve it before the generic media library
	// so a colliding media record cannot bypass moderation state.
	if s != nil && s.showcase != nil && showcaseStorageKeyIsInShowcaseNamespace(normalizedKey) {
		access, err := s.showcase.PublicImageAccess(ctx, normalizedKey)
		if err != nil {
			return false, err
		}
		if access.Found {
			return access.Allowed, nil
		}
		return false, nil
	}
	if showcaseStorageKeyIsInShowcaseNamespace(normalizedKey) {
		return false, nil
	}

	if s != nil && s.media != nil {
		access, err := s.media.PublicUploadAssetAccess(normalizedKey)
		if err != nil {
			return false, err
		}
		if access.Found {
			return access.Allowed, nil
		}
	}

	return true, nil
}
