package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

func (s *ProductService) validateCarbonRimProductTypeAndSystemCategoryBinding(productType string, categoryID *uint) error {
	isCarbonRim := productType == product.ProductTypeCarbonRim
	if categoryID == nil || *categoryID == 0 {
		if isCarbonRim {
			return fmt.Errorf("%w: carbon rims require the carbon-rim system category", ErrProductCategoryInvalid)
		}
		return nil
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
	if !isCarbonRim {
		if category.Slug == product.CarbonRimProductCategorySlug {
			return fmt.Errorf("%w: carbon-rim is reserved for dedicated carbon_rim products", ErrProductCategoryInvalid)
		}
		return nil
	}
	if category.Slug != product.CarbonRimProductCategorySlug || !category.IsEnabled || category.Depth != 2 || category.ParentID == nil {
		return fmt.Errorf("%w: carbon rims require the enabled carbon-rim system category under wheel-components", ErrProductCategoryInvalid)
	}
	parent, err := s.productCategoryRepo.FindByID(*category.ParentID)
	if err != nil {
		return fmt.Errorf("%w: carbon-rim parent category is unavailable", ErrProductCategoryInvalid)
	}
	if parent.Slug != SystemProductCategoryWheelComponentsSlug || !parent.IsEnabled || parent.Depth != 1 || parent.ParentID != nil {
		return fmt.Errorf("%w: carbon-rim must remain under the enabled wheel-components root", ErrProductCategoryInvalid)
	}
	return nil
}
