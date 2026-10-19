package repository

import (
	"encoding/json"
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/pkg/websitecontent"
)

func (r *MediaRepository) websiteProfileSettingMediaReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable("settings") {
		return references, nil
	}
	keys := []string{setting.WebsiteProfileKeyAvatarURL, setting.WebsiteProfileKeyFactoryImageURL, setting.WebsiteProfileKeyStatementBody, setting.WebsiteProfileKeyStatementParagraph1, setting.WebsiteProfileKeyStatementParagraph2, refundcancellation.Key}
	for afterID := uint(0); ; {
		var rows []setting.Setting
		if err := r.db.Unscoped().Select("id", "key", "locale", "value", "is_public", "deleted_at").Where("id > ? AND key IN ?", afterID, keys).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			matches := false
			if row.Key == refundcancellation.Key {
				var policy refundcancellation.Policy
				if err := json.Unmarshal([]byte(row.Value), &policy); err != nil {
					return nil, fmt.Errorf("decode retained refund policy media references: %w", err)
				}
				matches = r.matchesExactRegisteredOrTrustedMediaReference(query, policy.ContactURL)
				for _, section := range policy.Sections {
					if section.Image != nil && r.matchesExactRegisteredOrTrustedMediaReference(query, section.Image.URL) {
						matches = true
						break
					}
				}
			} else if setting.IsWebsiteProfileSharedMediaImageSettingKey(row.Key) {
				matches = r.matchesExactRegisteredOrTrustedMediaReference(query, row.Value)
			} else {
				content, err := websitecontent.NormalizeWebsiteBody(row.Value)
				if err != nil {
					return nil, err
				}
				document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(content)
				if err != nil {
					return nil, err
				}
				for _, attribute := range document.References {
					if r.matchesExactRegisteredOrTrustedMediaReference(query, *attribute.URL) {
						matches = true
						break
					}
				}
			}
			if !matches {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryContent, "website_profile_setting", row.ID, 0, fmt.Sprintf("站点资料 · %s · %s", row.Locale, row.Key), row.Key)
			if row.Key == refundcancellation.Key {
				reference.ResourceType = "refund_cancellation_policy_setting"
				reference.Label = fmt.Sprintf("退款取消政策 · %s", row.Locale)
			}
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
