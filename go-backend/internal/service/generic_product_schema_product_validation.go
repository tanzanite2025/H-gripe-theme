package service

import (
	"fmt"

	"commerce-platform/internal/domain/product"
)

func validateGenericProductSchemaCreation(input ProductCreateInput) error {
	kind := dedicatedProductTypeValue(input.ProductType)
	if input.GenericProductSchema != nil && kind != "" {
		return fmt.Errorf("%w: generic product schema is only available for ordinary products", ErrProductSpecInvalid)
	}
	if err := product.ValidateGenericProductSchema(input.GenericProductSchema); err != nil {
		return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
	}
	return nil
}

func validateGenericProductSchemaUpdate(existing *product.Product, input *ProductUpdateInput) error {
	if existing == nil || input == nil {
		return nil
	}
	if input.GenericProductSchema != nil && existing.ProductType != "" {
		return fmt.Errorf("%w: generic product schema is only available for ordinary products", ErrProductSpecInvalid)
	}
	if input.GenericProductSchema != nil && !input.UpdateGenericProductSchema {
		return fmt.Errorf("%w: generic product schema update marker is required", ErrProductSpecInvalid)
	}
	if input.UpdateGenericProductSchema && existing.ProductType != "" {
		return fmt.Errorf("%w: generic product schema is only available for ordinary products", ErrProductSpecInvalid)
	}
	if input.UpdateGenericProductSchema {
		if err := product.ValidateGenericProductSchema(input.GenericProductSchema); err != nil {
			return fmt.Errorf("%w: %v", ErrProductSpecInvalid, err)
		}
	}
	return nil
}
