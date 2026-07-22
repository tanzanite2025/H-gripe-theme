package repository

import (
	"fmt"
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

func replaceProductMedia(tx *gorm.DB, productID uint, mediaItems []product.ProductMedia) error {
	var existingItems []product.ProductMedia
	if err := tx.Where("product_id = ?", productID).Find(&existingItems).Error; err != nil {
		return err
	}

	existingByID := make(map[uint]product.ProductMedia, len(existingItems))
	for _, item := range existingItems {
		existingByID[item.ID] = item
	}

	keepIDs := make([]uint, 0, len(mediaItems))
	for i := range mediaItems {
		mediaItems[i].ProductID = productID
		if mediaItems[i].ID != 0 {
			if _, ok := existingByID[mediaItems[i].ID]; !ok {
				return fmt.Errorf("media %d does not belong to product %d", mediaItems[i].ID, productID)
			}
			if err := tx.Save(&mediaItems[i]).Error; err != nil {
				return err
			}
			keepIDs = append(keepIDs, mediaItems[i].ID)
			continue
		}

		if err := tx.Create(&mediaItems[i]).Error; err != nil {
			return err
		}
		keepIDs = append(keepIDs, mediaItems[i].ID)
	}

	deleteQuery := tx.Where("product_id = ?", productID)
	if len(keepIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keepIDs)
	}
	return deleteQuery.Delete(&product.ProductMedia{}).Error
}
