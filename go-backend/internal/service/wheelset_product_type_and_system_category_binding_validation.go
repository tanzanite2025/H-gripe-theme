package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

func (s *ProductService) validateWheelsetProductTypeAndSystemCategoryBinding(productType string, categoryID *uint) error {
	if productType != product.ProductTypeWheelset {
		return nil
	}
	if categoryID == nil || *categoryID == 0 {
		return fmt.Errorf("%w: wheelset products require the wheelset system category", ErrProductCategoryInvalid)
	}
	if s.productCategoryRepo == nil {
		return fmt.Errorf("%w: category repository is not configured", ErrProductCategoryInvalid)
	}
	category, err := s.productCategoryRepo.FindByID(*categoryID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductCategoryNotFound
	}
	if err != nil {
		return err
	}
	if category.Slug != SystemProductCategoryWheelsetSlug || !category.IsEnabled {
		return fmt.Errorf("%w: wheelset products require the enabled wheelset system category", ErrProductCategoryInvalid)
	}
	return nil
}
