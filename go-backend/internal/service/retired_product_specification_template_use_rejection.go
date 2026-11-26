package service

import (
	"fmt"

	"commerce-platform/internal/domain/product"
)

func (s *ProductService) rejectRetiredProductSpecificationTemplateUse(templateID, existingTemplateID *uint) error {
	if existingTemplateID != nil {
		existingTemplate, err := s.GetProductSpecificationTemplate(*existingTemplateID)
		if err != nil {
			return err
		}
		if product.IsRetiredWheelsetSpecificationTemplateSlug(existingTemplate.Slug) {
			return fmt.Errorf("%w: wheelset template bindings are retired; wheelset products use dedicated product-level specifications", ErrProductSpecificationTemplateInvalid)
		}
	}
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
	if product.IsRetiredWheelsetSpecificationTemplateSlug(template.Slug) {
		return fmt.Errorf("%w: wheelset templates are retired; create a dedicated wheelset product", ErrProductSpecificationTemplateInvalid)
	}
	return nil
}
