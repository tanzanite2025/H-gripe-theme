package repository

import (
	"encoding/json"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

func synchronizeTranslatedSpokesProductSpecificationsInTransaction(tx *gorm.DB, source *product.Product) error {
	if tx == nil || source == nil || source.ProductType != product.ProductTypeSpoke || !source.SpokeSpecificationDirty || source.ParentID != nil {
		return nil
	}
	encodedSpecification, err := json.Marshal(source.SpokeSpecification)
	if err != nil {
		return err
	}
	return tx.Model(&product.Product{}).
		Where("parent_id = ? AND product_type = ?", source.ID, product.ProductTypeSpoke).
		Update("spoke_specification", string(encodedSpecification)).Error
}
