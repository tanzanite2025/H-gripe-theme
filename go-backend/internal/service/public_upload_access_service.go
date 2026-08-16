package service

import (
	"context"

	"commerce-platform/internal/pkg/storage"
)

type PublicUploadAccessService struct {
	media                 *MediaService
	showcase              *ShowcaseService
	customerServiceAvatar *CustomerServiceAvatarService
}

func NewPublicUploadAccessService(
	mediaService *MediaService,
	showcaseService *ShowcaseService,
	customerServiceAvatarService ...*CustomerServiceAvatarService,
) *PublicUploadAccessService {
	var avatarService *CustomerServiceAvatarService
	if len(customerServiceAvatarService) > 0 {
		avatarService = customerServiceAvatarService[0]
	}
	return &PublicUploadAccessService{
		media:                 mediaService,
		showcase:              showcaseService,
		customerServiceAvatar: avatarService,
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

	if IsCustomerServiceAvatarStorageKey(normalizedKey) {
		if s == nil || s.customerServiceAvatar == nil {
			return false, nil
		}
		return s.customerServiceAvatar.CanServePublicAvatar(ctx, normalizedKey)
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
