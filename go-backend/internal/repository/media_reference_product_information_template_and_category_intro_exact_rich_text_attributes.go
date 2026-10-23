package repository

import (
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/safehtml"
)

// Keep dormant, disabled and deleted content in retention. These new reference
// types do not grant anonymous access through publication heuristics.
func (r *MediaRepository) productInformationTemplateAndCategoryIntroRichTextMediaAttributeReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	var references []media.AssetReference
	for _, source := range []struct{ table, field, resourceType, label string }{
		{"product_information_templates", "content", "product_information_template_rich_text", "商品信息模板"},
		{"product_categories", "seo_intro", "product_category_rich_text", "商品分类导语"},
		{"product_category_translations", "seo_intro", "product_category_translation_rich_text", "商品分类翻译导语"},
	} {
		if !r.hasTable(source.table) || !r.db.Migrator().HasColumn(source.table, source.field) {
			continue
		}
		deletedAtSelection := "NULL AS deleted_at"
		if r.db.Migrator().HasColumn(source.table, "deleted_at") {
			deletedAtSelection = "deleted_at"
		}
		for afterID := uint(0); ; {
			var rows []struct {
				ID             uint
				Name, RichText string
				DeletedAt      *time.Time
			}
			if err := r.db.Table(source.table).Select("id, name, "+source.field+" AS rich_text, "+deletedAtSelection).
				Where("id > ? AND COALESCE("+source.field+", '') <> ''", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
				return nil, err
			}
			for _, row := range rows {
				document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(row.RichText)
				if err != nil {
					return nil, err
				}
				for _, attribute := range document.References {
					if !r.matchesExactRegisteredOrTrustedMediaReference(query, *attribute.URL) {
						continue
					}
					reference := newMediaReference(media.ReferenceCategoryCatalog, source.resourceType, row.ID, 0, namedMediaReferenceLabel(source.label, row.ID, row.Name), source.field)
					reference.UsageState = "active"
					if row.DeletedAt != nil {
						reference.UsageState = "retained_deleted_parent"
					}
					references = append(references, reference)
					break
				}
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	return references, nil
}
