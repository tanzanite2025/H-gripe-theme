package localbackup

import (
	"net/http"

	"gorm.io/gorm"

	"commerce-platform/internal/api/publicuploadresponse"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

const stoppedProductSiteIdentityImageVerification = "site_identity_owner_authorized_image_bytes_bound_read_only_sql_and_resealed_before_delivery"

func (manager *Manager) bindStoppedProductPublicImageSiteIdentityOwnersAndRestrictedLogoFileServer(database *gorm.DB, shared storage.StorageService, access *service.PublicUploadAccessService) (http.Handler, error) {
	configuration := manager.Configuration.StoppedProductLocalMediaVerification
	logos, err := storage.NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(storage.Config{
		Type: storage.StorageTypeLocal, LocalPath: manager.Configuration.StorageRoots["site_logo_uploads"], PrivateLocalPath: manager.Configuration.StorageRoots["uploads_private"],
		BaseURL: configuration.LogoBaseURL, PrivateBaseURL: configuration.LogoPrivateBaseURL,
	})
	if err != nil {
		return nil, err
	}
	// These are the real business authorization services: current owners and
	// published historical consumers are resolved against the same read-only SQL
	// and owning storage identities. No shared-library collision grants access.
	// A shared media origin is not the storefront origin. Until the latter is
	// independently registered, only root-relative and the owner's own storage
	// references may authorize historical objects; never invent a site alias.
	access.ConfigureSiteLogoService(service.NewSiteLogoService(repository.NewSiteLogoRepository(database), logos, ""))
	access.ConfigureSiteFaviconService(service.NewSiteFaviconService(repository.NewSiteFaviconRepository(database), shared, ""))
	return publicuploadresponse.NewRestrictedLocalUploadFileServer(manager.Configuration.StorageRoots["site_logo_uploads"]), nil
}

func stoppedProductPublicSiteIdentityImageOwnerForObjectKey(key string) string {
	if service.IsSiteLogoStorageKey(key) {
		return "site_logo"
	}
	if service.IsSiteFaviconStorageKey(key) {
		return "site_favicon"
	}
	return ""
}
