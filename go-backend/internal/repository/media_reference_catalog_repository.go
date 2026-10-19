package repository

import (
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
)

func (r *MediaRepository) productMediaReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("product_media") {
		return []media.AssetReference{}, nil
	}

	type row struct {
		ID               uint
		ProductID        uint
		MediaAssetID     *uint
		URL              string
		ThumbnailURL     string
		PosterURL        string
		ImageVariants    string
		DeletedAt        *time.Time
		ExistingParentID *uint
		ParentDeletedAt  *time.Time
	}
	references := make([]media.AssetReference, 0)
	parentStateAvailable := r.hasTable("products")
	variantSnapshotsAvailable := r.hasProductMediaImageVariantsColumn()
	for afterID := uint(0); ; {
		var rows []row
		scan := r.db.Table("product_media").Select("product_media.id, product_media.product_id, product_media.media_asset_id, product_media.url, product_media.thumbnail_url, product_media.poster_url, product_media.deleted_at")
		if parentStateAvailable {
			scan = scan.Select("product_media.id, product_media.product_id, product_media.media_asset_id, product_media.url, product_media.thumbnail_url, product_media.poster_url, product_media.deleted_at, products.id AS existing_parent_id, products.deleted_at AS parent_deleted_at").
				Joins("LEFT JOIN products ON products.id = product_media.product_id")
		}
		if !query.IncludeRetainedDeletedPlacements {
			scan = scan.Where("product_media.deleted_at IS NULL")
		}
		if variantSnapshotsAvailable {
			scan = scan.Select(append(scan.Statement.Selects, "product_media.image_variants"))
		}
		err := scan.Where("product_media.id > ?", afterID).Order("product_media.id ASC").Limit(256).Find(&rows).Error
		if err != nil {
			return nil, err
		}

		for _, item := range rows {
			fields := make([]string, 0, 4)
			if query.AssetID != 0 && item.MediaAssetID != nil && *item.MediaAssetID == query.AssetID {
				fields = append(fields, "media_asset_id")
			}
			if r.matchesExactRegisteredOrTrustedMediaReference(query, item.URL) {
				fields = append(fields, "url")
			}
			if r.matchesExactRegisteredOrTrustedMediaReference(query, item.ThumbnailURL) {
				fields = append(fields, "thumbnail_url")
			}
			if r.matchesExactRegisteredOrTrustedMediaReference(query, item.PosterURL) {
				fields = append(fields, "poster_url")
			}
			if variantSnapshotsAvailable {
				matches, err := r.decodedLegacyMediaJSONContainsExactReference(query, item.ImageVariants, "product_media", "image_variants", item.ID)
				if err != nil {
					return nil, err
				}
				if matches {
					fields = append(fields, "image_variants")
				}
			}
			if len(fields) == 0 {
				continue
			}
			reference := newMediaReference(
				media.ReferenceCategoryCatalog,
				"product_media",
				item.ID,
				item.ProductID,
				fmt.Sprintf("商品 #%d 的媒体 #%d", item.ProductID, item.ID),
				strings.Join(fields, ", "),
			)
			if parentStateAvailable {
				reference.UsageState = "active"
				if item.ExistingParentID == nil {
					reference.UsageState = "missing_parent"
					reference.Label += "（商品记录不存在）"
				} else if item.ParentDeletedAt != nil {
					reference.UsageState = "retained_deleted_parent"
					reference.Label += "（已删除商品保留）"
				}
			}
			if item.DeletedAt != nil {
				reference.UsageState = "retained_deleted_placement"
				reference.Label += "（已删除媒体记录保留）"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
