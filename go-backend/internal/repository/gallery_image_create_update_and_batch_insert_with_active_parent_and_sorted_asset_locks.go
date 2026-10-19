package repository

import (
	"commerce-platform/internal/domain/gallery"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
)

func (r *GalleryRepository) createGalleryImagesWithParentAndAssetLocks(images []gallery.GalleryImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		parents := make(map[uint]bool)
		for _, image := range images {
			if image.ID != 0 {
				return gorm.ErrInvalidData
			}
			parents[image.GalleryID] = true
		}
		if err := lockActiveGalleryParentsInAscendingOrderWithinMutation(tx, parents); err != nil {
			return err
		}
		var bindings []publicBusinessMediaReferenceBinding
		for index := range images {
			var err error
			bindings, err = appendGalleryImageMediaBindingsForChangedPlacements(bindings, &images[index], &gallery.GalleryImage{})
			if err != nil {
				return err
			}
		}
		if err := r.validateAndCanonicalizeGalleryMediaBindingsWithOneSortedAssetLockPlan(tx, bindings); err != nil {
			return err
		}
		for index := range images {
			if images[index].Thumbnail == "" {
				images[index].Thumbnail = images[index].URL
			}
		}
		return tx.Omit(clause.Associations).Create(&images).Error
	})
}

func lockActiveGalleryParentsInAscendingOrderWithinMutation(tx *gorm.DB, parents map[uint]bool) error {
	ids := make([]uint, 0, len(parents))
	for id := range parents {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if err := lockActiveGalleryParentWithinMutation(tx, id, &gallery.Gallery{}); err != nil {
			return err
		}
	}
	return nil
}

func (r *GalleryRepository) updateGalleryImageWithParentAndAssetLocks(image *gallery.GalleryImage, mutations ...GalleryImageMediaFieldMutation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := lockActiveGalleryParentWithinMutation(tx, image.GalleryID, &gallery.Gallery{}); err != nil {
			return err
		}
		var previous gallery.GalleryImage
		if image.ID == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("gallery_id = ?", image.GalleryID).First(&previous, image.ID).Error; err != nil {
			return err
		}
		if len(mutations) > 0 {
			if !mutations[0].UpdateOriginalImage {
				image.MediaAssetID, image.URL = previous.MediaAssetID, previous.URL
			}
			if !mutations[0].UpdateThumbnail {
				image.Thumbnail = previous.Thumbnail
			}
		}
		bindings, err := appendGalleryImageMediaBindingsForChangedPlacements(nil, image, &previous)
		if err != nil {
			return err
		}
		if err := r.validateAndCanonicalizeGalleryMediaBindingsWithOneSortedAssetLockPlan(tx, bindings); err != nil {
			return err
		}
		if image.Thumbnail == "" && (image.URL != previous.URL || !sameGalleryMediaAssetID(image.MediaAssetID, previous.MediaAssetID)) {
			image.Thumbnail = image.URL
		}
		return tx.Model(&gallery.GalleryImage{}).Where("id = ? AND gallery_id = ?", image.ID, image.GalleryID).Select("*").Omit(clause.Associations, "id", "gallery_id", "created_at", "deleted_at").Updates(image).Error
	})
}
