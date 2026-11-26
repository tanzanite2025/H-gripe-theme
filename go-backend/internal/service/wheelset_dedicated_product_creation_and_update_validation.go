package service

import (
	"errors"
	"fmt"
	"reflect"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

func (s *ProductService) validateDedicatedWheelsetCreation(input ProductCreateInput) error {
	if dedicatedProductTypeValue(input.ProductType) != product.ProductTypeWheelset {
		return s.rejectNewProductBindingToWheelsetCategory(input.ProductCategoryID)
	}
	input.WheelsetSpecification.Normalize()
	if err := product.ValidateWheelsetSpecification(input.WheelsetSpecification); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	if input.ProductSpecificationTemplateID != nil || len(input.SpecValues) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.ParentID != nil {
		return fmt.Errorf("%w: wheelset products use their dedicated product specification, not a generic template, repair-kit models or a manual translation parent", ErrProductSpecInvalid)
	}
	return s.validateWheelsetProductTypeAndSystemCategoryBinding(product.ProductTypeWheelset, input.ProductCategoryID)
}

func (s *ProductService) validateDedicatedWheelsetUpdate(existing *product.Product, input *ProductUpdateInput) error {
	if existing.ProductType != product.ProductTypeWheelset {
		if input.UpdateProductCategoryID && !uintPointerEqual(existing.ProductCategoryID, input.ProductCategoryID) {
			return s.rejectNewProductBindingToWheelsetCategory(input.ProductCategoryID)
		}
		return nil
	}

	effectiveCategoryID := existing.ProductCategoryID
	if input.UpdateProductCategoryID {
		effectiveCategoryID = input.ProductCategoryID
	}
	if err := s.validateWheelsetProductTypeAndSystemCategoryBinding(existing.ProductType, effectiveCategoryID); err != nil {
		return err
	}
	if input.ProductSpecificationTemplateID != nil || len(input.SpecValues) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.UpdateParentID {
		return fmt.Errorf("%w: wheelset products use their dedicated product specification, not a generic template, repair-kit models or a manual translation parent", ErrProductSpecInvalid)
	}
	if input.WheelsetSpecification != nil && !input.UpdateWheelsetSpecification {
		return fmt.Errorf("%w: wheelset specification update marker is required", ErrProductSpecInvalid)
	}
	if input.UpdateWheelsetSpecification {
		input.WheelsetSpecification.Normalize()
		if err := product.ValidateWheelsetSpecification(input.WheelsetSpecification); err != nil {
			return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
		}
		if existing.ParentID != nil && !reflect.DeepEqual(input.WheelsetSpecification, existing.WheelsetSpecification) {
			return fmt.Errorf("%w: edit shared wheelset engineering facts on the source-language product", ErrProductSpecInvalid)
		}
	}
	effectiveSpecification := existing.WheelsetSpecification
	if input.UpdateWheelsetSpecification {
		effectiveSpecification = input.WheelsetSpecification
	}
	if err := product.ValidateWheelsetSpecification(effectiveSpecification); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	return nil
}

func (s *ProductService) rejectNewProductBindingToWheelsetCategory(categoryID *uint) error {
	if categoryID == nil || *categoryID == 0 {
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
	if category.Slug == SystemProductCategoryWheelsetSlug {
		return fmt.Errorf("%w: the wheelset system category is reserved for wheelset products", ErrProductCategoryInvalid)
	}
	return nil
}
