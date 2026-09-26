package service

import (
	"context"

	"commerce-platform/internal/pkg/storage"
)

type PublicUploadAccessService struct {
	media                 *MediaService
	showcase              *UGCShowcaseService
	customerServiceAvatar *CustomerServiceAvatarService
	siteLogo              *SiteLogoService
	siteFavicon           *SiteFaviconService
}

func NewPublicUploadAccessService(
	mediaService *MediaService,
	ugcShowcaseService *UGCShowcaseService,
	customerServiceAvatarService ...*CustomerServiceAvatarService,
) *PublicUploadAccessService {
	var avatarService *CustomerServiceAvatarService
	if len(customerServiceAvatarService) > 0 {
		avatarService = customerServiceAvatarService[0]
	}
	return &PublicUploadAccessService{
		media:                 mediaService,
		showcase:              ugcShowcaseService,
		customerServiceAvatar: avatarService,
	}
}

func (s *PublicUploadAccessService) ConfigureSiteLogoService(siteLogoService *SiteLogoService) {
	if s == nil {
		return
	}
	s.siteLogo = siteLogoService
}

func (s *PublicUploadAccessService) ConfigureSiteFaviconService(siteFaviconService *SiteFaviconService) {
	if s == nil {
		return
	}
	s.siteFavicon = siteFaviconService
}

func (s *PublicUploadAccessService) CanServePublicUpload(ctx context.Context, key string) (bool, error) {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	if !ok {
		return false, nil
	}
	if storage.IsPrivateObjectKey(normalizedKey) {
		return false, nil
	}
	if showcaseStorageKeyIsPending(normalizedKey) {
		return false, nil
	}
	if IsAfterSalesEvidenceStorageKey(normalizedKey) {
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

	if IsSiteLogoStorageKey(normalizedKey) {
		if s == nil || s.siteLogo == nil {
			return false, nil
		}
		return s.siteLogo.CanServePublicLogo(ctx, normalizedKey)
	}
	if IsSiteFaviconStorageKey(normalizedKey) {
		if s == nil || s.siteFavicon == nil {
			return false, nil
		}
		return s.siteFavicon.CanServePublicFavicon(ctx, normalizedKey)
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

	// Public upload access is an explicit allow-list. Unknown keys must never
	// be served: callers need a registered media/showcase/avatar/logo record
	// whose visibility has been positively verified above.
	return false, nil
}
