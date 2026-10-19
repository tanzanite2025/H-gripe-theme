package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
)

var ErrMediaAssociationRepairConflict = errors.New("media association changed or is outside the supported repair scope")

// ReadMediaAssociationDiagnosisSnapshot never locks business writers or changes receipts.
func (r *MediaRepository) ReadMediaAssociationDiagnosisSnapshot(ctx context.Context, consume func(*MediaRepository, *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return consume(r.WithTx(tx), tx)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
}

// LockMissingGalleryParentRepairTables also protects the absence of the old
// parent against insert/restore. This rare manual operation deliberately uses
// short, bounded table locks; every normal gallery writer participates through
// PostgreSQL's ordinary table locks, including writers unaware of this feature.
func (r *MediaRepository) LockMissingGalleryParentRepairTables() error {
	if r.db.Name() != "postgres" {
		return ErrMediaAssociationRepairConflict
	}
	if err := r.db.Exec("SET LOCAL lock_timeout = '2s'").Error; err != nil {
		return err
	}
	return r.db.Exec("LOCK TABLE galleries, gallery_images IN SHARE ROW EXCLUSIVE MODE").Error
}

func (r *MediaRepository) LoadMissingGalleryParentRepairRecords(assetID, imageID, destinationID uint, lockRows bool) (*media.MediaAsset, *gallery.GalleryImage, *gallery.Gallery, error) {
	queryMissingGalleryRepairRecord := func() *gorm.DB {
		query := r.db.Session(&gorm.Session{})
		if lockRows {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		return query
	}
	var image gallery.GalleryImage
	if err := queryMissingGalleryRepairRecord().Unscoped().First(&image, imageID).Error; err != nil {
		return nil, nil, nil, err
	}
	var parentCount int64
	if err := r.db.Unscoped().Model(&gallery.Gallery{}).Where("id = ?", image.GalleryID).Count(&parentCount).Error; err != nil {
		return nil, nil, nil, err
	}
	if image.DeletedAt.Valid || parentCount != 0 || image.MediaAssetID == nil || *image.MediaAssetID != assetID {
		return nil, nil, nil, ErrMediaAssociationRepairConflict
	}
	var destination gallery.Gallery
	if err := queryMissingGalleryRepairRecord().First(&destination, destinationID).Error; err != nil {
		return nil, nil, nil, err
	}
	if destination.Status != "draft" || destination.ID == image.GalleryID {
		return nil, nil, nil, ErrMediaAssociationRepairConflict
	}
	var asset media.MediaAsset
	if err := queryMissingGalleryRepairRecord().First(&asset, assetID).Error; err != nil {
		return nil, nil, nil, err
	}
	// First release supports only an exact original binding. Aliases,
	// derivatives and independently registered thumbnails need a separate review.
	if !media.CanMediaAssetBeReusedByPublicBusinessContent(&asset) || asset.MediaType != "image" ||
		asset.StorageKey == "" || image.URL != asset.URL || (image.Thumbnail != "" && image.Thumbnail != asset.URL) {
		return nil, nil, nil, ErrMediaAssociationRepairConflict
	}
	return &asset, &image, &destination, nil
}

func (r *MediaRepository) ReattachMissingGalleryParentWithoutChangingImageMetadata(image *gallery.GalleryImage, destination *gallery.Gallery) error {
	image.UpdatedAt = r.db.NowFunc()
	changed := r.db.Model(&gallery.GalleryImage{}).Where("id = ? AND gallery_id = ?", image.ID, image.GalleryID).
		Updates(map[string]interface{}{"gallery_id": destination.ID, "updated_at": image.UpdatedAt})
	if changed.Error != nil {
		return changed.Error
	}
	if changed.RowsAffected != 1 {
		return ErrMediaAssociationRepairConflict
	}
	if err := r.db.Model(&gallery.Gallery{}).Where("id = ?", destination.ID).Update("updated_at", r.db.NowFunc()).Error; err != nil {
		return err
	}
	// Audit the persisted timestamp precision, rather than the in-memory clock.
	return r.db.First(image, image.ID).Error
}

func (r *MediaRepository) ValidateMediaAssetHasNoUnboundBusinessDraftUploadReceipt(assetID uint) error {
	if !r.hasTable("business_draft_media_upload_receipts") {
		return nil
	}
	var count int64
	if err := r.db.Model(&media.BusinessDraftMediaUploadReceipt{}).Where("asset_id = ? AND status <> ?", assetID, "bound").Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return ErrMediaAssociationRepairConflict
	}
	return nil
}
