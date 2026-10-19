package repository

import (
	"fmt"
	"sort"
	"strings"

	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
)

// The dedicated owner, not a supplied URL or a generic import, determines every
// mirror value. The caller passes its owner transaction so metadata, mirrors and
// cleanup events commit together. Existing deleted settings are never revived.
func (r *SettingRepository) SynchronizeDedicatedSiteIdentitySettingMirrorsFromCurrentOwner(key, requestedLocale, publicSiteBaseURL string) ([]setting.Setting, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	namespace, description := "", ""
	switch key {
	case setting.SiteLogoSettingKey:
		namespace, description = dedicatedSiteLogoObjectNamespace, "Site logo URL"
	case setting.SiteFaviconSettingKey:
		namespace, description = dedicatedSiteFaviconObjectNamespace, "Browser favicon URL"
	default:
		return nil, fmt.Errorf("%w: no dedicated site identity owns this key", ErrSettingMediaReferenceInvalid)
	}
	locale := strings.TrimSpace(requestedLocale)
	if locale == "" {
		locale = "en"
	}
	var mirrors []setting.Setting
	err := runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, namespace, func(tx *gorm.DB) error {
		storageKey := ""
		hasOwner := false
		if key == setting.SiteLogoSettingKey {
			owner, err := NewSiteLogoRepository(tx).Current()
			if err != nil {
				return err
			}
			if owner != nil {
				storageKey, hasOwner = owner.StorageKey, true
			}
		} else {
			owner, err := NewSiteFaviconRepository(tx).Current()
			if err != nil {
				return err
			}
			if owner != nil {
				storageKey, hasOwner = owner.StorageKey, true
			}
		}
		value := ""
		if hasOwner {
			normalized, ok := storage.NormalizeObjectKey(storageKey)
			if !ok || dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(normalized) != namespace {
				return fmt.Errorf("%w: current dedicated site identity has an invalid object key", ErrSettingMediaReferenceInvalid)
			}
			value = strings.TrimRight(strings.TrimSpace(publicSiteBaseURL), "/") + "/uploads/" + normalized
		}
		var existing []setting.Setting
		if err := tx.Where("key = ?", key).Order("locale ASC").Find(&existing).Error; err != nil {
			return err
		}
		locales := map[string]struct{}{locale: {}}
		for _, item := range existing {
			locales[item.Locale] = struct{}{}
		}
		orderedLocales := make([]string, 0, len(locales))
		for locale := range locales {
			orderedLocales = append(orderedLocales, locale)
		}
		sort.Strings(orderedLocales)
		for _, locale := range orderedLocales {
			item, created, err := lockOrReserveSettingKeyAndLocaleForTransactionalMediaWrite(tx, settingKeyAndLocaleIdentity{Key: key, Locale: locale})
			if err != nil {
				return err
			}
			if created {
				item.Type, item.Group, item.IsPublic, item.Description = "string", "site", true, description
			}
			item.Value = value
			result := tx.Model(&setting.Setting{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"value": item.Value, "type": item.Type, "group": item.Group, "is_public": item.IsPublic, "description": item.Description,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			mirrors = append(mirrors, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mirrors, nil
}
