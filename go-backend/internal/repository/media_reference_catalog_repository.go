package repository

import (
	"fmt"
	"strings"
	"tanzanite/internal/domain/media"
)

func (r *MediaRepository) productMediaReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("product_media") {
		return []media.AssetReference{}, nil
	}

	type row struct {
		ID           uint
		ProductID    uint
		MediaAssetID *uint
		URL          string
		ThumbnailURL string
		PosterURL    string
	}
	var rows []row
	err := r.db.Table("product_media").
		Select("id, product_id, media_asset_id, url, thumbnail_url, poster_url").
		Where("deleted_at IS NULL").
		Where("media_asset_id = ? OR url IN ? OR thumbnail_url IN ? OR poster_url IN ?", query.AssetID, query.URLs, query.URLs, query.URLs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		fields := make([]string, 0, 4)
		if item.MediaAssetID != nil && *item.MediaAssetID == query.AssetID {
			fields = append(fields, "media_asset_id")
		}
		if containsMediaReferenceURL(query.URLs, item.URL) {
			fields = append(fields, "url")
		}
		if containsMediaReferenceURL(query.URLs, item.ThumbnailURL) {
			fields = append(fields, "thumbnail_url")
		}
		if containsMediaReferenceURL(query.URLs, item.PosterURL) {
			fields = append(fields, "poster_url")
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryCatalog,
			"product_media",
			item.ID,
			item.ProductID,
			fmt.Sprintf("商品 #%d 的媒体 #%d", item.ProductID, item.ID),
			strings.Join(fields, ", "),
		))
	}
	return references, nil
}

func (r *MediaRepository) galleryReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if r.hasTable("galleries") {
		type galleryRow struct {
			ID   uint
			Name string
		}
		var rows []galleryRow
		if err := r.db.Table("galleries").
			Select("id, name").
			Where("deleted_at IS NULL AND cover_image IN ?", query.URLs).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, item := range rows {
			references = append(references, newMediaReference(
				media.ReferenceCategoryCatalog,
				"gallery_cover",
				item.ID,
				0,
				namedMediaReferenceLabel("图库", item.ID, item.Name),
				"cover_image",
			))
		}
	}

	if !r.hasTable("gallery_images") {
		return references, nil
	}

	type imageRow struct {
		ID        uint
		GalleryID uint
		URL       string
		Thumbnail string
	}
	var imageRows []imageRow
	if err := r.db.Table("gallery_images").
		Select("id, gallery_id, url, thumbnail").
		Where("deleted_at IS NULL AND (url IN ? OR thumbnail IN ?)", query.URLs, query.URLs).
		Find(&imageRows).Error; err != nil {
		return nil, err
	}
	for _, item := range imageRows {
		fields := make([]string, 0, 2)
		if containsMediaReferenceURL(query.URLs, item.URL) {
			fields = append(fields, "url")
		}
		if containsMediaReferenceURL(query.URLs, item.Thumbnail) {
			fields = append(fields, "thumbnail")
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryCatalog,
			"gallery_image",
			item.ID,
			item.GalleryID,
			fmt.Sprintf("图库 #%d 的图片 #%d", item.GalleryID, item.ID),
			strings.Join(fields, ", "),
		))
	}
	return references, nil
}

func (r *MediaRepository) giftCardReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("gift_cards") {
		return []media.AssetReference{}, nil
	}

	type row struct {
		ID   uint
		Code string
	}
	var rows []row
	if err := r.db.Table("gift_cards").
		Select("id, code").
		Where("deleted_at IS NULL AND cover_image IN ?", query.URLs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		label := fmt.Sprintf("礼品卡 #%d", item.ID)
		if strings.TrimSpace(item.Code) != "" {
			label = fmt.Sprintf("礼品卡 #%d：%s", item.ID, item.Code)
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryCatalog,
			"gift_card",
			item.ID,
			0,
			label,
			"cover_image",
		))
	}
	return references, nil
}
