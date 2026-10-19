package repository

import (
	"commerce-platform/internal/domain/media"
	"strings"
)

func (r *MediaRepository) productCategoryImageReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("product_categories") {
		return []media.AssetReference{}, nil
	}

	type row struct {
		ID                uint
		Name              string
		ImageMediaAssetID *uint
		ImageURL          string
	}
	references := make([]media.AssetReference, 0)
	for afterID := uint(0); ; {
		var rows []row
		if err := r.db.Table("product_categories").
			Select("id, name, image_media_asset_id, image_url").
			Where("id > ?", afterID).Order("id ASC").Limit(256).
			Find(&rows).Error; err != nil {
			return nil, err
		}

		for _, item := range rows {
			fields := make([]string, 0, 2)
			if query.AssetID != 0 && item.ImageMediaAssetID != nil && *item.ImageMediaAssetID == query.AssetID {
				fields = append(fields, "image_media_asset_id")
			}
			if r.matchesExactRegisteredOrTrustedMediaReference(query, item.ImageURL) {
				fields = append(fields, "image_url")
			}
			if len(fields) == 0 {
				continue
			}
			references = append(references, newMediaReference(
				media.ReferenceCategoryCatalog,
				"product_category",
				item.ID,
				0,
				namedMediaReferenceLabel("商品分类", item.ID, item.Name),
				strings.Join(fields, ", "),
			))
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
