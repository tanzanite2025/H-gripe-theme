package service

import (
	"fmt"

	"commerce-platform/internal/domain/product"
)

func dedicatedProductTypeValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validateDedicatedProductTypeCreation(input ProductCreateInput) error {
	switch dedicatedProductTypeValue(input.ProductType) {
	case "":
		if input.BicycleTireSpecification != nil {
			return fmt.Errorf("%w: bicycle tire specification requires bicycle_tire product_type", ErrProductSpecInvalid)
		}
		if input.CarbonRimSpecification != nil {
			return fmt.Errorf("%w: carbon rim specification requires carbon_rim product_type", ErrProductSpecInvalid)
		}
		if input.WheelsetSpecification != nil {
			return fmt.Errorf("%w: wheelset specification requires wheelset product_type", ErrProductSpecInvalid)
		}
		if input.SpokeSpecification != nil {
			return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
		}
		return rejectCarbonRimVariantSpecificationsOutsideCarbonRimProduct(input.Variants)
	case product.ProductTypeCarbonRim:
		if input.BicycleTireSpecification != nil {
			return fmt.Errorf("%w: bicycle tire specification requires bicycle_tire product_type", ErrProductSpecInvalid)
		}
		if input.WheelsetSpecification != nil {
			return fmt.Errorf("%w: wheelset specification requires wheelset product_type", ErrProductSpecInvalid)
		}
		if input.SpokeSpecification != nil {
			return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
		}
	case product.ProductTypeBicycleTire:
		if input.CarbonRimSpecification != nil {
			return fmt.Errorf("%w: carbon rim specification requires carbon_rim product_type", ErrProductSpecInvalid)
		}
		if input.WheelsetSpecification != nil {
			return fmt.Errorf("%w: wheelset specification requires wheelset product_type", ErrProductSpecInvalid)
		}
		if input.SpokeSpecification != nil {
			return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
		}
		return rejectCarbonRimVariantSpecificationsOutsideCarbonRimProduct(input.Variants)
	case product.ProductTypeWheelset:
		if input.CarbonRimSpecification != nil || input.BicycleTireSpecification != nil {
			return fmt.Errorf("%w: wheelset products cannot include carbon rim or bicycle tire specifications", ErrProductSpecInvalid)
		}
		if input.SpokeSpecification != nil {
			return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
		}
		return rejectCarbonRimVariantSpecificationsOutsideCarbonRimProduct(input.Variants)
	case product.ProductTypeSpoke:
		if input.CarbonRimSpecification != nil || input.BicycleTireSpecification != nil || input.WheelsetSpecification != nil {
			return fmt.Errorf("%w: spoke products cannot include carbon rim, bicycle tire or wheelset specifications", ErrProductSpecInvalid)
		}
		return rejectCarbonRimVariantSpecificationsOutsideCarbonRimProduct(input.Variants)
	default:
		return fmt.Errorf("%w: unsupported product_type", ErrProductSpecInvalid)
	}
	return nil
}

func validateDedicatedProductTypeUpdate(existing *product.Product, input *ProductUpdateInput) error {
	if input.ProductType != nil && *input.ProductType != existing.ProductType {
		return fmt.Errorf("%w: product_type is immutable; dedicated product type changes require an explicit migration", ErrProductSpecInvalid)
	}
	if existing.ProductType != product.ProductTypeBicycleTire && (input.UpdateBicycleTireSpecification || input.BicycleTireSpecification != nil) {
		return fmt.Errorf("%w: bicycle tire specification requires bicycle_tire product_type", ErrProductSpecInvalid)
	}
	if existing.ProductType != product.ProductTypeCarbonRim && (input.UpdateCarbonRimSpecification || input.CarbonRimSpecification != nil) {
		return fmt.Errorf("%w: carbon rim specification requires carbon_rim product_type", ErrProductSpecInvalid)
	}
	if existing.ProductType != product.ProductTypeWheelset && (input.UpdateWheelsetSpecification || input.WheelsetSpecification != nil) {
		return fmt.Errorf("%w: wheelset specification requires wheelset product_type", ErrProductSpecInvalid)
	}
	if existing.ProductType != product.ProductTypeSpoke && (input.UpdateSpokeSpecification || input.SpokeSpecification != nil) {
		return fmt.Errorf("%w: spoke specification requires spoke product_type", ErrProductSpecInvalid)
	}
	if existing.ProductType == product.ProductTypeWheelset && (input.UpdateCarbonRimSpecification || input.CarbonRimSpecification != nil || input.UpdateBicycleTireSpecification || input.BicycleTireSpecification != nil) {
		return fmt.Errorf("%w: wheelset products cannot include carbon rim or bicycle tire specifications", ErrProductSpecInvalid)
	}
	if existing.ProductType == product.ProductTypeSpoke && (input.UpdateCarbonRimSpecification || input.CarbonRimSpecification != nil || input.UpdateBicycleTireSpecification || input.BicycleTireSpecification != nil || input.UpdateWheelsetSpecification || input.WheelsetSpecification != nil) {
		return fmt.Errorf("%w: spoke products cannot include carbon rim, bicycle tire or wheelset specifications", ErrProductSpecInvalid)
	}
	if input.UpdateVariants && existing.ProductType != product.ProductTypeCarbonRim {
		return rejectCarbonRimVariantSpecificationsOutsideCarbonRimProduct(input.Variants)
	}
	return nil
}

func rejectCarbonRimVariantSpecificationsOutsideCarbonRimProduct(inputs []ProductVariantInput) error {
	for _, input := range inputs {
		if input.CarbonRimSpecification != nil {
			return fmt.Errorf("%w: carbon rim SKU specification requires carbon_rim product_type", ErrProductVariantInvalid)
		}
	}
	return nil
}
