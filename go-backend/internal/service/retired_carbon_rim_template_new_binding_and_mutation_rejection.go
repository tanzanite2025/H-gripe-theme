package service

import (
	"commerce-platform/internal/domain/product"
	"fmt"
)

func (s *ProductService) rejectNewBindingToRetiredCarbonRimTemplate(templateID, existingTemplateID *uint) error {
	if templateID == nil || (existingTemplateID != nil && *templateID == *existingTemplateID) {
		return nil
	}
	template, err := s.GetProductSpecificationTemplate(*templateID)
	if err != nil {
		return err
	}
	if product.IsRetiredCarbonRimSpecificationTemplateSlug(template.Slug) {
		return fmt.Errorf("%w: rim templates are retired; create a dedicated carbon_rim product", ErrProductSpecificationTemplateInvalid)
	}
	return nil
}
