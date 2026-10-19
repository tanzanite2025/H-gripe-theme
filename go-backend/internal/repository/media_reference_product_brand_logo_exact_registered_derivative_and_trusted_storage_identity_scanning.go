package repository

import "commerce-platform/internal/domain/media"

func (r *MediaRepository) productBrandLogoReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("product_brands") {
		return nil, nil
	}
	type brandLogoReferenceRow struct {
		ID      uint
		Name    string
		LogoURL string
	}
	references := make([]media.AssetReference, 0)
	for afterID := uint(0); ; {
		var rows []brandLogoReferenceRow
		if err := r.db.Table("product_brands").Select("id, name, logo_url").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if !r.matchesExactRegisteredOrTrustedMediaReference(query, row.LogoURL) {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryCatalog, "product_brand_logo", row.ID, 0, namedMediaReferenceLabel("商品品牌 Logo", row.ID, row.Name), "logo_url")
			reference.UsageState = "active"
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
