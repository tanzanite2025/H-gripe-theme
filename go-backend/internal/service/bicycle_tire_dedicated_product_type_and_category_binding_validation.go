package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

func (s *ProductService) validateBicycleTireProductTypeAndSystemCategoryBinding(productType string, categoryID *uint) error {
	if productType != product.ProductTypeBicycleTire {
		return nil
	}
	if categoryID == nil || *categoryID == 0 {
		return fmt.Errorf("%w: bicycle tires require the tire system category", ErrProductCategoryInvalid)
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
	if category.Slug != product.BicycleTireProductCategorySlug || !category.IsEnabled || category.Depth != 2 || category.ParentID == nil {
		return fmt.Errorf("%w: bicycle tires require the enabled tire category under wheel-components", ErrProductCategoryInvalid)
	}
	parent, err := s.productCategoryRepo.FindByID(*category.ParentID)
	if err != nil {
		return fmt.Errorf("%w: tire category parent is unavailable", ErrProductCategoryInvalid)
	}
	if parent.Slug != SystemProductCategoryWheelComponentsSlug || !parent.IsEnabled || parent.Depth != 1 || parent.ParentID != nil {
		return fmt.Errorf("%w: tire category must remain under the enabled wheel-components root", ErrProductCategoryInvalid)
	}
	return nil
}
