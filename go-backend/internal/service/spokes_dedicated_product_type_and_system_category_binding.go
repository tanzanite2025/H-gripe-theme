package service

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

func (s *ProductService) validateDedicatedSpokeCreation(input ProductCreateInput) error {
	kind := dedicatedProductTypeValue(input.ProductType)
	if kind != product.ProductTypeSpoke {
		if input.SpokeSpecification != nil {
			return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
		}
		return s.validateSpokeProductTypeAndSystemCategoryBinding(kind, input.ProductCategoryID)
	}
	if input.SpokeSpecification == nil {
		return fmt.Errorf("%w: spoke specification document is required", ErrProductSpecInvalid)
	}
	input.SpokeSpecification.Normalize()
	if err := product.ValidateSpokesProductSpecification(input.SpokeSpecification); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	if input.ProductSpecificationTemplateID != nil || len(input.SpecValues) != 0 || len(input.VariantOptionValues) != 0 || len(input.OptionValueRelations) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.ParentID != nil {
		return fmt.Errorf("%w: spokes use their product-level specification, not templates, generic options, repair-kit models or manual translation parents", ErrProductSpecInvalid)
	}
	if len(input.Variants) == 0 && strings.TrimSpace(input.Status) != "inactive" {
		return fmt.Errorf("%w: a spoke without sales variants must remain an inactive draft", ErrProductSpecInvalid)
	}
	return s.validateSpokeProductTypeAndSystemCategoryBinding(kind, input.ProductCategoryID)
}

func (s *ProductService) validateDedicatedSpokeUpdate(existing *product.Product, input *ProductUpdateInput) error {
	if existing.ProductType != product.ProductTypeSpoke {
		if input.UpdateProductCategoryID {
			return s.validateSpokeProductTypeAndSystemCategoryBinding(existing.ProductType, input.ProductCategoryID)
		}
		if input.UpdateSpokeSpecification || input.SpokeSpecification != nil {
			return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
		}
		return nil
	}

	effectiveCategoryID := existing.ProductCategoryID
	if input.UpdateProductCategoryID {
		effectiveCategoryID = input.ProductCategoryID
	}
	if err := s.validateSpokeProductTypeAndSystemCategoryBinding(existing.ProductType, effectiveCategoryID); err != nil {
		return err
	}
	if (input.UpdateProductSpecificationTemplateID && input.ProductSpecificationTemplateID != nil) || len(input.SpecValues) != 0 || len(input.VariantOptionValues) != 0 || len(input.OptionValueRelations) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.UpdateParentID {
		return fmt.Errorf("%w: spoke products cannot use templates, generic options, repair-kit models or manual translation parents", ErrProductSpecInvalid)
	}
	if input.SpokeSpecification != nil && !input.UpdateSpokeSpecification {
		return fmt.Errorf("%w: spoke specification update marker is required", ErrProductSpecInvalid)
	}
	if input.UpdateSpokeSpecification {
		input.SpokeSpecification.Normalize()
		if err := product.ValidateSpokesProductSpecification(input.SpokeSpecification); err != nil {
			return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
		}
		if existing.ParentID != nil && !reflect.DeepEqual(input.SpokeSpecification, existing.SpokeSpecification) {
			return fmt.Errorf("%w: edit shared spoke facts on the source-language product", ErrProductSpecInvalid)
		}
	}
	effectiveSpecification := existing.SpokeSpecification
	if input.UpdateSpokeSpecification {
		effectiveSpecification = input.SpokeSpecification
	}
	if err := product.ValidateSpokesProductSpecification(effectiveSpecification); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	if input.Status != nil && strings.TrimSpace(*input.Status) == "active" {
		variants := existing.Variants
		if input.UpdateVariants {
			variants = make([]product.ProductVariant, 0, len(input.Variants))
			for _, variant := range input.Variants {
				if variant.IsActive == nil || *variant.IsActive {
					variants = append(variants, product.ProductVariant{IsActive: true})
				}
			}
		}
		if len(variants) == 0 {
			return fmt.Errorf("%w: add a sales SKU before activating a spoke product", ErrProductVariantInvalid)
		}
	}
	return nil
}

func (s *ProductService) validateSpokeProductTypeAndSystemCategoryBinding(productType string, categoryID *uint) error {
	if categoryID == nil || *categoryID == 0 {
		if productType == product.ProductTypeSpoke {
			return fmt.Errorf("%w: spoke products require the spokes system category", ErrProductCategoryInvalid)
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
	if strings.EqualFold(strings.TrimSpace(category.Slug), product.SpokesProductCategorySlug) {
		if productType != product.ProductTypeSpoke {
			return fmt.Errorf("%w: spokes is reserved for the dedicated spoke product type", ErrProductCategoryInvalid)
		}
		if !category.IsEnabled {
			return fmt.Errorf("%w: the spokes system category must be enabled", ErrProductCategoryInvalid)
		}
		return nil
	}
	if productType == product.ProductTypeSpoke {
		return fmt.Errorf("%w: spoke products must belong to the spokes system category", ErrProductCategoryInvalid)
	}
	return nil
}
