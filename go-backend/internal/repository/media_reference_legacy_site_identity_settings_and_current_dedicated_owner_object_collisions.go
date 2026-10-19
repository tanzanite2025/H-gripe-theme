package repository

import (
	"fmt"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
)

// Settings remain real consumers even when hidden or retained after deletion.
// Mirrors are read directly; no reference ledger or extra asset is created.
func (r *MediaRepository) siteIdentitySettingMediaReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable("settings") {
		return references, nil
	}
	for afterID := uint(0); ; {
		var rows []setting.Setting
		if err := r.db.Unscoped().Select("id", "key", "locale", "value", "is_public", "deleted_at").Where("id > ? AND key IN ?", afterID, []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey}).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if !r.matchesExactRegisteredOrTrustedMediaReference(query, row.Value) {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryContent, "site_identity_setting", row.ID, 0, fmt.Sprintf("站点标识设置 · %s · %s", row.Locale, row.Key), row.Key)
			reference.UsageState = "active"
			if !row.IsPublic {
				reference.Label += "（未公开设置保留）"
			}
			if row.DeletedAt.Valid {
				reference.UsageState = "retained_deleted_parent"
				reference.Label += "（已删除设置保留）"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

// A shared registration with a dedicated owner's full key may be historical
// duplicate metadata, or a different physical store. Protect it for review;
// neither merge identities nor infer that basename matches identify one object.
func (r *MediaRepository) currentDedicatedSiteIdentityAndVisualShowcaseMediaReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	for _, owner := range []struct{ table, resourceType, label string }{
		{"site_logo_assets", "site_logo", "当前站点 Logo"},
		{"site_favicon_assets", "site_favicon", "当前站点 Favicon"},
	} {
		if !r.hasTable(owner.table) {
			continue
		}
		var rows []struct {
			ID              uint
			URL, StorageKey string
		}
		if err := r.db.Table(owner.table).Select("id", "url", "storage_key").Where("id = ?", 1).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if r.matchesExactRegisteredOrTrustedMediaReference(query, row.URL) || registeredMediaReferenceQueryContainsExactObjectKey(query, row.StorageKey) {
				reference := newMediaReference(media.ReferenceCategoryContent, owner.resourceType, row.ID, 0, owner.label+"（专用所有者，需核对存储身份）", "storage_key")
				reference.UsageState = "active"
				references = append(references, reference)
			}
		}
	}
	if !r.hasTable("visual_showcase_items") {
		return references, nil
	}
	err := NewHomeVisualTileRepository(r.db).VisitActiveVisualShowcaseImageReferences(false, func(tile homevisualtile.Tile) bool {
		for _, field := range []struct {
			name, value string
			objectKey   bool
		}{
			{"storage_key", tile.StorageKey, true}, {"image_url", tile.ImageURL, false}, {"thumbnail_url", tile.ThumbnailURL, false},
		} {
			matches := registeredMediaReferenceQueryContainsExactObjectKey(query, field.value)
			if !field.objectKey {
				matches = r.matchesExactRegisteredOrTrustedMediaReference(query, field.value)
			}
			if !matches {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryContent, "home_visual_tile", tile.ID, 0, fmt.Sprintf("首页展示 · %s · %s（含草稿保留）", tile.TileSetKey, tile.Locale), field.name)
			reference.UsageState = "active"
			references = append(references, reference)
		}
		return true
	})
	return references, err
}

func registeredMediaReferenceQueryContainsExactObjectKey(query mediaAssetReferenceQuery, value string) bool {
	key, valid := storage.NormalizeObjectKey(value)
	if !valid {
		return false
	}
	for _, candidate := range query.ObjectKeys {
		if key == candidate {
			return true
		}
	}
	return false
}
