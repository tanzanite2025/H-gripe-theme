package service

import "fmt"

func (s *ProductService) rejectRetiredProductSpecificationTemplateUse(templateID, existingTemplateID *uint) error {
	if templateID == nil && existingTemplateID == nil {
		return nil
	}
	return fmt.Errorf("%w: %w: product specification templates are retired; use dedicated product-level fields and manually maintained SKU rows", ErrProductSpecificationTemplateInvalid, ErrProductSpecInvalid)
}
