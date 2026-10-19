package repository

import (
	"commerce-platform/internal/domain/media"
	"fmt"
	"strings"
	"time"
)

func (r *MediaRepository) galleryReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	hasParents := r.hasTable("galleries")
	hasParentPublicationStatus := hasParents && r.db.Migrator().HasColumn("galleries", "status")
	if hasParents {
		for afterID := uint(0); ; {
			var rows []struct {
				ID               uint
				Name, CoverImage string
				Status           string
				DeletedAt        *time.Time
			}
			coverColumns := "id, name, cover_image, deleted_at"
			if hasParentPublicationStatus {
				coverColumns += ", status"
			}
			if err := r.db.Table("galleries").Select(coverColumns).Where("id > ? AND COALESCE(cover_image,'') <> ''", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
				return nil, err
			}
			for _, row := range rows {
				if !r.matchesExactRegisteredOrTrustedMediaReference(query, row.CoverImage) {
					continue
				}
				reference := newMediaReference(media.ReferenceCategoryCatalog, "gallery_cover", row.ID, 0, namedMediaReferenceLabel("图库", row.ID, row.Name), "cover_image")
				reference.UsageState = "active"
				if row.DeletedAt != nil {
					reference.UsageState = "retained_deleted_parent"
					reference.Label += "（已删除图库保留）"
				} else if hasParentPublicationStatus && row.Status != "published" {
					reference.UsageState = "retained_unpublished_parent"
					reference.Label += "（未发布图库保留）"
				}
				references = append(references, reference)
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	if !r.hasTable("gallery_images") {
		return references, nil
	}
	for afterID := uint(0); ; {
		var rows []struct {
			ID, GalleryID    uint
			MediaAssetID     *uint
			URL, Thumbnail   string
			DeletedAt        *time.Time
			ExistingParentID *uint
			ParentDeletedAt  *time.Time
			ParentStatus     string
		}
		selectColumns := "gallery_images.id, gallery_images.gallery_id, gallery_images.media_asset_id, gallery_images.url, gallery_images.thumbnail, gallery_images.deleted_at"
		scan := r.db.Table("gallery_images")
		if hasParents {
			selectColumns += ", galleries.id AS existing_parent_id, galleries.deleted_at AS parent_deleted_at"
			scan = scan.Joins("LEFT JOIN galleries ON galleries.id = gallery_images.gallery_id")
			if hasParentPublicationStatus {
				selectColumns += ", galleries.status AS parent_status"
			}
		}
		if err := scan.Select(selectColumns).Where("gallery_images.id > ?", afterID).Order("gallery_images.id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			var fields []string
			if row.MediaAssetID != nil && *row.MediaAssetID == query.AssetID {
				fields = append(fields, "media_asset_id")
			}
			if r.matchesExactRegisteredOrTrustedMediaReference(query, row.URL) {
				fields = append(fields, "url")
			}
			if r.matchesExactRegisteredOrTrustedMediaReference(query, row.Thumbnail) {
				fields = append(fields, "thumbnail")
			}
			if len(fields) == 0 {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryCatalog, "gallery_image", row.ID, row.GalleryID, fmt.Sprintf("图库 #%d 的图片 #%d", row.GalleryID, row.ID), strings.Join(fields, ", "))
			reference.UsageState = "active"
			if hasParents && row.ExistingParentID == nil {
				reference.UsageState = "missing_parent"
				reference.Label += "（图库记录不存在）"
			} else if row.ParentDeletedAt != nil {
				reference.UsageState = "retained_deleted_parent"
				reference.Label += "（已删除图库保留）"
			} else if row.DeletedAt != nil {
				reference.UsageState = "retained_deleted_placement"
				reference.Label += "（已删除图片记录保留）"
			} else if hasParentPublicationStatus && row.ParentStatus != "published" {
				reference.UsageState = "retained_unpublished_parent"
				reference.Label += "（未发布图库保留）"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
