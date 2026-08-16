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
	var rows []row
	if err := r.db.Table("product_categories").
		Select("id, name, image_media_asset_id, image_url").
		Where("image_media_asset_id = ? OR image_url IN ?", query.AssetID, query.URLs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		fields := make([]string, 0, 2)
		if item.ImageMediaAssetID != nil && *item.ImageMediaAssetID == query.AssetID {
			fields = append(fields, "image_media_asset_id")
		}
		if containsMediaReferenceURL(query.URLs, item.ImageURL) {
			fields = append(fields, "image_url")
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
	return references, nil
}
