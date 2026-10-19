package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/safehtml"
)

func (r *MediaRepository) postReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable("posts") {
		return references, nil
	}
	for afterID := uint(0); ; {
		var rows []struct {
			ID                          uint
			Title, Content, FeaturedImg string
			DeletedAt                   *time.Time
		}
		if err := r.db.Table("posts").Select("id, title, content, featured_img, deleted_at").Where("id > ? AND (COALESCE(content, '') <> '' OR COALESCE(featured_img, '') <> '')", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			var fields []string
			if r.matchesExactRegisteredOrTrustedMediaReference(query, row.FeaturedImg) {
				fields = append(fields, "featured_img")
			}
			content, err := safehtml.Sanitize(row.Content)
			if err != nil {
				return nil, err
			}
			document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(content)
			if err != nil {
				return nil, err
			}
			for _, attribute := range document.References {
				if r.matchesExactRegisteredOrTrustedMediaReference(query, *attribute.URL) {
					fields = append(fields, "content")
					break
				}
			}
			if len(fields) == 0 {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryContent, "post", row.ID, 0, namedMediaReferenceLabel("文章", row.ID, row.Title), strings.Join(fields, ", "))
			reference.UsageState = "active"
			if row.DeletedAt != nil {
				reference.UsageState = "retained_deleted_parent"
				reference.Label += "（已删除文章保留）"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
