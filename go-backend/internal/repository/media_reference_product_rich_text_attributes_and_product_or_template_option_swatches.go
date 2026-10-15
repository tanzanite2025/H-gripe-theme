package repository

import (
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/safehtml"
)

func (r *MediaRepository) productRichTextMediaAttributeReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable("products") {
		return references, nil
	}
	// HTML entities and historical aliases make a substring SQL filter unsafe.
	// Bound memory with ID pages and verify parsed attributes exactly.
	for afterID := uint(0); ; {
		var rows []struct {
			ID                           uint
			Name, Description, ShortDesc string
			DeletedAt                    *time.Time
		}
		if err := r.db.Table("products").Select("id, name, description, short_desc, deleted_at").Where("id > ? AND (COALESCE(description, '') <> '' OR COALESCE(short_desc, '') <> '')", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			var fields []string
			for _, field := range []struct{ name, html string }{{"description", row.Description}, {"short_desc", row.ShortDesc}} {
				document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(field.html)
				if err != nil {
					return nil, err
				}
				for _, attribute := range document.References {
					if r.matchesExactRegisteredOrTrustedMediaReference(query, *attribute.URL) {
						fields = append(fields, field.name)
						break
					}
				}
			}
			if len(fields) == 0 {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryCatalog, "product_rich_text", row.ID, 0, namedMediaReferenceLabel("商品详情", row.ID, row.Name), strings.Join(fields, ", "))
			reference.UsageState = "active"
			if row.DeletedAt != nil {
				reference.UsageState = "retained_deleted_parent"
				reference.Label += "（已删除商品保留）"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

func (r *MediaRepository) productAndTemplateOptionSwatchReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	for _, table := range []string{"product_variant_option_values", "product_spec_option_items"} {
		if !r.hasTable(table) {
			continue
		}
		template := table == "product_spec_option_items"
		for afterID := uint(0); ; {
			var rows []struct {
				ID, ParentID       uint
				SwatchMediaAssetID *uint
				SwatchURL, Label   string
				ExistingParentID   *uint
				ParentDeletedAt    *time.Time
			}
			scan := r.db.Table(table+" AS swatches").Where("swatches.id > ? AND (swatches.swatch_media_asset_id IS NOT NULL OR COALESCE(swatches.swatch_url, '') <> '')", afterID).Order("swatches.id ASC").Limit(256)
			parentStateAvailable := false
			if template {
				scan = scan.Select("swatches.id, swatches.spec_definition_id AS parent_id, swatches.default_label AS label, swatches.swatch_media_asset_id, swatches.swatch_url")
				if r.hasTable("product_spec_definitions") && r.hasTable("product_specification_templates") {
					parentStateAvailable = true
					scan = scan.Select("swatches.id, COALESCE(definitions.product_specification_template_id, 0) AS parent_id, swatches.default_label AS label, swatches.swatch_media_asset_id, swatches.swatch_url, templates.id AS existing_parent_id").Joins("LEFT JOIN product_spec_definitions AS definitions ON definitions.id = swatches.spec_definition_id").Joins("LEFT JOIN product_specification_templates AS templates ON templates.id = definitions.product_specification_template_id")
				}
			} else {
				scan = scan.Select("swatches.id, swatches.product_id AS parent_id, swatches.label, swatches.swatch_media_asset_id, swatches.swatch_url")
				if r.hasTable("products") {
					parentStateAvailable = true
					scan = scan.Select("swatches.id, swatches.product_id AS parent_id, swatches.label, swatches.swatch_media_asset_id, swatches.swatch_url, products.id AS existing_parent_id, products.deleted_at AS parent_deleted_at").Joins("LEFT JOIN products ON products.id = swatches.product_id")
				}
			}
			if err := scan.Find(&rows).Error; err != nil {
				return nil, err
			}
			for _, row := range rows {
				var fields []string
				if row.SwatchMediaAssetID != nil && *row.SwatchMediaAssetID == query.AssetID {
					fields = append(fields, "swatch_media_asset_id")
				}
				if r.matchesExactRegisteredOrTrustedMediaReference(query, row.SwatchURL) {
					fields = append(fields, "swatch_url")
				}
				if len(fields) == 0 {
					continue
				}
				resourceType, ownerLabel := "product_option_swatch", "商品"
				if template {
					resourceType, ownerLabel = "product_template_option_swatch", "规格模板"
				}
				reference := newMediaReference(media.ReferenceCategoryCatalog, resourceType, row.ID, row.ParentID, fmt.Sprintf("%s #%d 的色板 #%d：%s", ownerLabel, row.ParentID, row.ID, truncateMediaReferenceLabel(row.Label)), strings.Join(fields, ", "))
				if parentStateAvailable {
					reference.UsageState = "active"
					if row.ExistingParentID == nil {
						reference.UsageState = "missing_parent"
						reference.Label += "（所属记录不存在）"
					} else if row.ParentDeletedAt != nil {
						reference.UsageState = "retained_deleted_parent"
						reference.Label += "（已删除商品保留）"
					}
				}
				references = append(references, reference)
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	return references, nil
}
