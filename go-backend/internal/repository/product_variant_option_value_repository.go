package repository

import (
	"commerce-platform/internal/domain/product"
	"fmt"

	"gorm.io/gorm"
)

func replaceProductVariantOptionValues(tx *gorm.DB, productID uint, values []product.ProductVariantOptionValue) error {
	var existing []product.ProductVariantOptionValue
	if err := tx.Where("product_id = ?", productID).Find(&existing).Error; err != nil {
		return err
	}

	existingByID := make(map[uint]product.ProductVariantOptionValue, len(existing))
	for _, value := range existing {
		existingByID[value.ID] = value
	}

	keepIDs := make([]uint, 0, len(values))
	for i := range values {
		values[i].ProductID = productID
		if values[i].ID != 0 {
			if _, ok := existingByID[values[i].ID]; !ok {
				return fmt.Errorf("%w: variant option value %d does not belong to product %d", ErrProductVariantOptionValueReferenceInvalid, values[i].ID, productID)
			}
			if err := tx.Save(&values[i]).Error; err != nil {
				return err
			}
			if err := saveCustomOptionPolicy(tx, &values[i]); err != nil {
				return err
			}
			keepIDs = append(keepIDs, values[i].ID)
			continue
		}

		if err := tx.Create(&values[i]).Error; err != nil {
			return err
		}
		if err := saveCustomOptionPolicy(tx, &values[i]); err != nil {
			return err
		}
		keepIDs = append(keepIDs, values[i].ID)
	}

	deleteQuery := tx.Where("product_id = ?", productID)
	if len(keepIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keepIDs)
	}
	return deleteQuery.Delete(&product.ProductVariantOptionValue{}).Error
}

func saveCustomOptionPolicy(tx *gorm.DB, value *product.ProductVariantOptionValue) error {
	if value == nil || value.CustomOptionPolicy == nil {
		return nil
	}
	value.CustomOptionPolicy.ProductVariantOptionValueID = value.ID
	var existing product.ProductCustomOptionPolicy
	err := tx.Where("product_variant_option_value_id = ?", value.ID).First(&existing).Error
	if err == nil {
		value.CustomOptionPolicy.ID = existing.ID
		return tx.Save(value.CustomOptionPolicy).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return tx.Create(value.CustomOptionPolicy).Error
}
