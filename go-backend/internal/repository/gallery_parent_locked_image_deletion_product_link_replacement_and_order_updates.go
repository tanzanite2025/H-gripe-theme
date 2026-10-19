package repository

import (
	"gorm.io/gorm"

	"commerce-platform/internal/domain/gallery"
)

func (r *GalleryRepository) deleteGalleryAndRetainItsSoftDeletedImageRecords(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := lockActiveGalleryParentWithinMutation(tx, id, &gallery.Gallery{}); err != nil {
			return err
		}
		if err := tx.Where("gallery_id = ?", id).Delete(&gallery.GalleryImage{}).Error; err != nil {
			return err
		}
		if tx.Migrator().HasTable(&gallery.GalleryProductLink{}) {
			if err := tx.Where("gallery_id = ?", id).Delete(&gallery.GalleryProductLink{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&gallery.Gallery{}, id).Error
	})
}

func (r *GalleryRepository) deleteGalleryImagesWithLockedParentsAndOwnershipCheck(galleryID *uint, ids []uint) (int64, error) {
	unique := make(map[uint]bool)
	for _, id := range ids {
		if id != 0 {
			unique[id] = true
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if len(unique) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	selected := make([]uint, 0, len(unique))
	for id := range unique {
		selected = append(selected, id)
	}
	var deleted int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var images []gallery.GalleryImage
		query := tx.Where("id IN ?", selected)
		if galleryID != nil {
			query = query.Where("gallery_id = ?", *galleryID)
		}
		if err := query.Find(&images).Error; err != nil {
			return err
		}
		if len(images) != len(unique) {
			return gorm.ErrRecordNotFound
		}
		parents := make(map[uint]bool)
		for _, image := range images {
			parents[image.GalleryID] = true
		}
		if err := lockActiveGalleryParentsInAscendingOrderWithinMutation(tx, parents); err != nil {
			return err
		}
		// Recheck after locking: another deletion may have committed while waiting.
		var currentCount int64
		if err := tx.Model(&gallery.GalleryImage{}).Where("id IN ?", selected).Count(&currentCount).Error; err != nil {
			return err
		}
		if currentCount != int64(len(unique)) {
			return gorm.ErrRecordNotFound
		}
		result := tx.Where("id IN ?", selected).Delete(&gallery.GalleryImage{})
		deleted = result.RowsAffected
		return result.Error
	})
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func (r *GalleryRepository) updateGalleryImageOrderWithLockedParent(id uint, order int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var image gallery.GalleryImage
		if err := tx.First(&image, id).Error; err != nil {
			return err
		}
		if err := lockActiveGalleryParentWithinMutation(tx, image.GalleryID, &gallery.Gallery{}); err != nil {
			return err
		}
		result := tx.Model(&gallery.GalleryImage{}).Where("id = ? AND gallery_id = ?", id, image.GalleryID).Update("order", order)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
