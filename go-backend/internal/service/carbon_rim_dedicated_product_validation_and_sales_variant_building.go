package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"commerce-platform/internal/domain/product"
)

func dedicatedProductTypeValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *ProductService) validateDedicatedCarbonRimCreation(input ProductCreateInput) error {
	kind := dedicatedProductTypeValue(input.ProductType)
	if kind != "" && kind != product.ProductTypeCarbonRim {
		return fmt.Errorf("%w: unsupported product_type", ErrProductSpecInvalid)
	}
	if kind == "" {
		if input.CarbonRimSpecification != nil {
			return fmt.Errorf("%w: carbon rim specification requires carbon_rim product_type", ErrProductSpecInvalid)
		}
		if err := rejectCarbonRimSpecificationsOnGenericVariants(input.Variants); err != nil {
			return err
		}
		return s.validateCarbonRimProductTypeAndSystemCategoryBinding(kind, input.ProductCategoryID)
	}
	if err := product.ValidateCarbonRimSpecification(input.CarbonRimSpecification); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	if err := product.ValidateCarbonRimWheelSizeMaterialAndRidingUses(input.CarbonRimSpecification, true); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	if input.ProductSpecificationTemplateID != nil || len(input.SpecValues) != 0 || len(input.VariantOptionValues) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.ParentID != nil {
		return fmt.Errorf("%w: carbon rims cannot bind a specification template, generic options, repair-kit models or a manual translation parent", ErrProductSpecInvalid)
	}
	return s.validateCarbonRimProductTypeAndSystemCategoryBinding(kind, input.ProductCategoryID)
}

func rejectCarbonRimSpecificationsOnGenericVariants(inputs []ProductVariantInput) error {
	for _, input := range inputs {
		if input.CarbonRimSpecification != nil {
			return fmt.Errorf("%w: carbon rim SKU specification requires carbon_rim product_type", ErrProductVariantInvalid)
		}
	}
	return nil
}

func (s *ProductService) validateDedicatedCarbonRimUpdate(existing *product.Product, input *ProductUpdateInput) error {
	if input.ProductType != nil && *input.ProductType != existing.ProductType {
		return fmt.Errorf("%w: product_type is immutable; existing template products require an explicit migration", ErrProductSpecInvalid)
	}
	effectiveCategoryID := existing.ProductCategoryID
	if input.UpdateProductCategoryID {
		effectiveCategoryID = input.ProductCategoryID
	}
	if err := s.validateCarbonRimProductTypeAndSystemCategoryBinding(existing.ProductType, effectiveCategoryID); err != nil {
		return err
	}
	if existing.ProductType != product.ProductTypeCarbonRim {
		if input.UpdateCarbonRimSpecification || input.CarbonRimSpecification != nil {
			return fmt.Errorf("%w: carbon rim specification requires carbon_rim product_type", ErrProductSpecInvalid)
		}
		return rejectCarbonRimSpecificationsOnGenericVariants(input.Variants)
	}
	if input.ProductSpecificationTemplateID != nil || len(input.SpecValues) != 0 || len(input.VariantOptionValues) != 0 || len(input.OptionValueRelations) != 0 || len(input.SpokeRepairKitModelKeys) != 0 || input.UpdateParentID {
		return fmt.Errorf("%w: carbon rim specifications cannot be replaced by a template, generic options, repair-kit models or another translation parent", ErrProductSpecInvalid)
	}
	if input.UpdateCarbonRimSpecification {
		if err := product.ValidateCarbonRimSpecification(input.CarbonRimSpecification); err != nil {
			return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
		}
		if existing.CarbonRimSpecification != nil && (existing.CarbonRimSpecification.WheelSize != "" || len(existing.CarbonRimSpecification.RidingUses) > 0) {
			if err := product.ValidateCarbonRimWheelSizeMaterialAndRidingUses(input.CarbonRimSpecification, true); err != nil {
				return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
			}
		}
		if existing.ParentID != nil && !reflect.DeepEqual(input.CarbonRimSpecification, existing.CarbonRimSpecification) {
			return fmt.Errorf("%w: edit shared carbon rim engineering facts on the source-language product", ErrProductSpecInvalid)
		}
	}
	if !input.UpdateVariants {
		return nil
	}
	byID := make(map[uint]product.ProductVariant)
	for _, variant := range existing.Variants {
		byID[variant.ID] = variant
	}
	seenIDs := make(map[uint]bool)
	for index := range input.Variants {
		variant := &input.Variants[index]
		if variant.ID == nil || *variant.ID == 0 {
			if existing.ParentID != nil {
				return fmt.Errorf("%w: add carbon rim SKUs on the source-language product", ErrProductVariantInvalid)
			}
			continue
		}
		saved, exists := byID[*variant.ID]
		if !exists || seenIDs[*variant.ID] {
			return fmt.Errorf("%w: duplicated or foreign variant ID", ErrProductVariantInvalid)
		}
		seenIDs[*variant.ID] = true
		if variant.CarbonRimSpecification == nil {
			variant.CarbonRimSpecification = saved.CarbonRimSpecification
		}
		if product.CarbonRimSalesConfigurationLabel(variant.CarbonRimSpecification) != product.CarbonRimSalesConfigurationLabel(saved.CarbonRimSpecification) {
			return fmt.Errorf("%w: existing SKU hole count, drilling and finish are immutable; add a new SKU and disable the old one", ErrProductVariantInvalid)
		}
		if existing.ParentID != nil && !reflect.DeepEqual(variant.CarbonRimSpecification, saved.CarbonRimSpecification) {
			return fmt.Errorf("%w: edit shared SKU engineering facts on the source-language product", ErrProductVariantInvalid)
		}
	}
	for id := range byID {
		if !seenIDs[id] {
			return fmt.Errorf("%w: retain existing carbon rim SKUs; disable unwanted rows instead of deleting them", ErrProductVariantInvalid)
		}
	}
	return nil
}

func (s *ProductService) buildProductVariantsForDedicatedType(kind string, templateID *uint, inputs []ProductVariantInput, currency string, options []product.ProductVariantOptionValue) ([]product.ProductVariant, error) {
	if kind != product.ProductTypeCarbonRim {
		return s.buildVariants(templateID, inputs, currency, options)
	}
	variants := make([]product.ProductVariant, 0, len(inputs))
	seenConfigurations := map[string]bool{}
	seenSKUs := map[string]bool{}
	defaultCount, activeCount := 0, 0
	for _, input := range inputs {
		if err := product.ValidateCarbonRimVariantSpecification(input.CarbonRimSpecification); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrProductVariantInvalid, err)
		}
		if len(input.OptionValues) != 0 || len(input.OptionGroupRules) != 0 || len(input.OptionValueRules) != 0 {
			return nil, fmt.Errorf("%w: carbon rim choices are derived from their SKU specification", ErrProductVariantInvalid)
		}
		if input.Weight <= 0 {
			return nil, fmt.Errorf("%w: carbon rim SKU weight_grams must be positive", ErrProductVariantInvalid)
		}
		label := product.CarbonRimSalesConfigurationLabel(input.CarbonRimSpecification)
		skuKey := strings.ToLower(strings.TrimSpace(input.SKU))
		if seenConfigurations[label] || seenSKUs[skuKey] {
			return nil, fmt.Errorf("%w: duplicate carbon rim configuration or SKU", ErrProductVariantInvalid)
		}
		seenConfigurations[label], seenSKUs[skuKey] = true, true
		isActive := input.IsActive == nil || *input.IsActive
		if input.IsDefault {
			defaultCount++
			if !isActive {
				return nil, fmt.Errorf("%w: default SKU must be enabled", ErrProductVariantInvalid)
			}
		}
		if isActive {
			activeCount++
		}
		validationInput := input
		validationInput.IsActive = nil
		validationInput.IsDefault = true
		built, err := s.buildVariants(nil, []ProductVariantInput{validationInput}, currency, nil)
		if err != nil {
			return nil, err
		}
		variant := built[0]
		variant.CarbonRimSpecification = input.CarbonRimSpecification
		variant.IsDefault, variant.IsActive = input.IsDefault, isActive
		encodedOptions, _ := json.Marshal(map[string]string{"rim_configuration": label})
		variant.OptionValues = string(encodedOptions)
		if strings.TrimSpace(input.Title) == "" {
			variant.Title = label
		}
		variants = append(variants, variant)
	}
	if defaultCount != 1 || activeCount == 0 {
		return nil, fmt.Errorf("%w: carbon rims require exactly one enabled default SKU", ErrProductVariantInvalid)
	}
	return variants, nil
}
