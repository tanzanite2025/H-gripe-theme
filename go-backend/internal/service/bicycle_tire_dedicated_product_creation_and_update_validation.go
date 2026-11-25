package service

import (
	"fmt"
	"reflect"
	"strings"

	"commerce-platform/internal/domain/product"
)

func (s *ProductService) validateDedicatedBicycleTireCreation(input ProductCreateInput) error {
	if dedicatedProductTypeValue(input.ProductType) != product.ProductTypeBicycleTire {
		return nil
	}
	if input.BicycleTireSpecification == nil {
		return fmt.Errorf("%w: bicycle tire specification document is required", ErrProductSpecInvalid)
	}
	input.BicycleTireSpecification.Normalize()
	if err := product.ValidateBicycleTireSpecification(input.BicycleTireSpecification); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	if input.ProductSpecificationTemplateID != nil || len(input.SpecValues) != 0 || len(input.VariantOptionValues) != 0 || len(input.OptionValueRelations) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.ParentID != nil {
		return fmt.Errorf("%w: bicycle tires use their dedicated product specification, not a template, generic options, repair-kit models or a manual translation parent", ErrProductSpecInvalid)
	}
	if err := s.validateBicycleTireBrandAndArticleNumber(input.BrandID, input.BicycleTireSpecification, 0); err != nil {
		return err
	}
	return s.validateBicycleTireProductTypeAndSystemCategoryBinding(product.ProductTypeBicycleTire, input.ProductCategoryID)
}

func (s *ProductService) validateDedicatedBicycleTireUpdate(existing *product.Product, input *ProductUpdateInput) error {
	if existing.ProductType != product.ProductTypeBicycleTire {
		return nil
	}

	effectiveCategoryID := existing.ProductCategoryID
	if input.UpdateProductCategoryID {
		effectiveCategoryID = input.ProductCategoryID
	}
	if err := s.validateBicycleTireProductTypeAndSystemCategoryBinding(existing.ProductType, effectiveCategoryID); err != nil {
		return err
	}
	if (input.UpdateProductSpecificationTemplateID && input.ProductSpecificationTemplateID != nil) || len(input.SpecValues) != 0 || len(input.VariantOptionValues) != 0 || len(input.OptionValueRelations) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.UpdateParentID {
		return fmt.Errorf("%w: bicycle tires use their dedicated product specification, not a template, generic options, repair-kit models or a manual translation parent", ErrProductSpecInvalid)
	}
	if input.BicycleTireSpecification != nil && !input.UpdateBicycleTireSpecification {
		return fmt.Errorf("%w: bicycle tire specification update marker is required", ErrProductSpecInvalid)
	}
	if input.UpdateBicycleTireSpecification {
		input.BicycleTireSpecification.Normalize()
		if err := product.ValidateBicycleTireSpecification(input.BicycleTireSpecification); err != nil {
			return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
		}
		if existing.ParentID != nil && !sameBicycleTireSpecification(input.BicycleTireSpecification, existing.BicycleTireSpecification) {
			return fmt.Errorf("%w: edit shared tire facts on the source-language product", ErrProductSpecInvalid)
		}
	}

	effectiveSpecification := existing.BicycleTireSpecification
	if input.UpdateBicycleTireSpecification {
		effectiveSpecification = input.BicycleTireSpecification
	}
	effectiveBrandID := existing.BrandID
	if input.UpdateBrandID {
		effectiveBrandID = input.BrandID
	}
	return s.validateBicycleTireBrandAndArticleNumber(effectiveBrandID, effectiveSpecification, existing.ID)
}

func sameBicycleTireSpecification(left, right *product.BicycleTireSpecification) bool {
	if left == nil || right == nil {
		return left == right
	}
	leftCopy, rightCopy := *left, *right
	leftCopy.Normalize()
	rightCopy.Normalize()
	return reflect.DeepEqual(leftCopy, rightCopy)
}

func (s *ProductService) validateBicycleTireBrandAndArticleNumber(brandID *uint, specification *product.BicycleTireSpecification, excludeProductID uint) error {
	if specification == nil || strings.TrimSpace(specification.ArticleNo) == "" {
		return nil
	}
	if brandID == nil || *brandID == 0 {
		return fmt.Errorf("%w: brand is required when Article No. is provided", ErrProductBrandInvalid)
	}
	if err := s.validateProductBrand(brandID, false); err != nil {
		return err
	}
	exists, err := s.productRepo.HasBicycleTireArticleNumber(*brandID, specification.ArticleNo, excludeProductID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: Article No. %q already belongs to another tire product under this brand", ErrProductSpecInvalid, specification.ArticleNo)
	}
	return nil
}
