package repository

import (
	"fmt"

	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
)

// Every writer takes these locks before any settings row, including first-write
// reservations. Dedicated services already own the same transaction lock.
func lockSiteIdentitySettingNamespacesBeforeSortedSettingRows(tx *gorm.DB, identities []settingKeyAndLocaleIdentity) error {
	keys := make(map[string]bool)
	for _, identity := range identities {
		keys[identity.Key] = true
	}
	for _, entry := range []struct{ key, namespace string }{
		{setting.SiteFaviconSettingKey, dedicatedSiteFaviconObjectNamespace},
		{setting.SiteLogoSettingKey, dedicatedSiteLogoObjectNamespace},
	} {
		if keys[entry.key] {
			if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(tx, entry.namespace); err != nil {
				return err
			}
		}
	}
	return nil
}

func dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(key string) string {
	return storage.DedicatedPublicObjectNamespaceForStorageKey(key)
}

func validateChangedSiteIdentitySettingUsesItsOwningLifecycle(tx *gorm.DB, value setting.Setting) error {
	if value.Key == setting.SiteFaviconSettingKey {
		if value.Value == "" {
			if !tx.Migrator().HasTable("site_favicon_assets") {
				return nil
			}
			current, err := NewSiteFaviconRepository(tx).Current()
			if err != nil {
				return err
			}
			if current == nil {
				return nil
			}
		}
		return fmt.Errorf("%w: favicon settings mirror the dedicated current owner", ErrSettingMediaReferenceInvalid)
	}
	if value.Key != setting.SiteLogoSettingKey || !tx.Migrator().HasTable("site_logo_assets") {
		return nil
	}
	current, err := NewSiteLogoRepository(tx).Current()
	if err != nil {
		return err
	}
	if current != nil {
		return fmt.Errorf("%w: the current dedicated logo must be changed through its own endpoint; reload stale settings", ErrSettingMediaReferenceInvalid)
	}
	return nil
}
