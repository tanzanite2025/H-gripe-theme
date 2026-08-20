package service

import (
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

var productSpecificationTemplateSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

type ProductSpecificationTemplateInput struct {
	Name            string
	Slug            string
	Description     string
	SortOrder       int
	IsEnabled       bool
	SpecDefinitions []ProductSpecDefinitionInput
}

type ProductSpecDefinitionInput struct {
	ID              uint
	Group           string
	Name            string
	Slug            string
	FieldType       string
	Presentation    string
	Unit            string
	IsRequired      bool
	IsFilterable    bool
	IsVisible       bool
	IsVariantOption bool
	SortOrder       int
	Options         string
	Validation      string
}

func (s *ProductService) GetAttributeByID(id uint) (*product.ProductAttribute, error) {
	return s.productRepo.FindAttributeByID(id)
}

func (s *ProductService) GetAttributeBySlug(slug string) (*product.ProductAttribute, error) {
	return s.productRepo.FindAttributeBySlug(slug)
}

func (s *ProductService) ListAttributes(page, pageSize int) ([]product.ProductAttribute, int64, error) {
	return s.productRepo.FindAllAttributes(page, pageSize)
}

func (s *ProductService) CreateAttribute(attr *product.ProductAttribute) error {
	return s.productRepo.CreateAttribute(attr)
}

func (s *ProductService) UpdateAttribute(attr *product.ProductAttribute) error {
	return s.productRepo.UpdateAttribute(attr)
}

func (s *ProductService) DeleteAttribute(id uint) error {
	return s.productRepo.DeleteAttribute(id)
}

func (s *ProductService) GetFilterableAttributes() ([]product.ProductAttribute, error) {
	return s.productRepo.FindFilterableAttributes()
}

func (s *ProductService) GetAttributeValueByID(id uint) (*product.AttributeValue, error) {
	return s.productRepo.FindAttributeValueByID(id)
}

func (s *ProductService) CreateAttributeValue(val *product.AttributeValue) error {
	return s.productRepo.CreateAttributeValue(val)
}

func (s *ProductService) UpdateAttributeValue(val *product.AttributeValue) error {
	return s.productRepo.UpdateAttributeValue(val)
}

func (s *ProductService) DeleteAttributeValue(id uint) error {
	return s.productRepo.DeleteAttributeValue(id)
}

func (s *ProductService) GetValuesByAttributeID(attrID uint) ([]product.AttributeValue, error) {
	return s.productRepo.FindValuesByAttributeID(attrID)
}

func (s *ProductService) ListProductSpecificationTemplates(includeDisabled bool) ([]product.ProductSpecificationTemplate, error) {
	return s.productRepo.FindAllProductSpecificationTemplates(includeDisabled)
}

func (s *ProductService) ListPublicProductSpecificationTemplates(includeDisabled bool) ([]product.ProductSpecificationTemplate, error) {
	return s.productRepo.FindPublicProductSpecificationTemplates(includeDisabled)
}

func (s *ProductService) GetProductSpecificationTemplate(id uint) (*product.ProductSpecificationTemplate, error) {
	productSpecificationTemplate, err := s.productRepo.FindProductSpecificationTemplateByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}
	return productSpecificationTemplate, nil
}

func (s *ProductService) CreateProductSpecificationTemplate(input ProductSpecificationTemplateInput) (*product.ProductSpecificationTemplate, error) {
	productSpecificationTemplate, err := normalizeProductSpecificationTemplateInput(input)
	if err != nil {
		return nil, err
	}

	exists, err := s.productRepo.ProductSpecificationTemplateSlugExists(productSpecificationTemplate.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductSpecificationTemplateSlugExists
	}

	if err := s.productRepo.CreateProductSpecificationTemplate(productSpecificationTemplate); err != nil {
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin product specification template create")
	return s.productRepo.FindProductSpecificationTemplateByID(productSpecificationTemplate.ID)
}

func (s *ProductService) UpdateProductSpecificationTemplate(id uint, input ProductSpecificationTemplateInput) (*product.ProductSpecificationTemplate, error) {
	existing, err := s.GetProductSpecificationTemplate(id)
	if err != nil {
		return nil, err
	}

	productSpecificationTemplate, err := normalizeProductSpecificationTemplateInput(input)
	if err != nil {
		return nil, err
	}
	productSpecificationTemplate.ID = id

	if existing.IsSystemManaged {
		if productSpecificationTemplate.Slug != existing.Slug {
			return nil, fmt.Errorf("%w: product specification template slug cannot be changed", ErrProductSpecificationTemplateSystemManaged)
		}
		if err := validateSystemManagedProductSpecificationTemplate(existing, productSpecificationTemplate); err != nil {
			return nil, err
		}
	}

	exists, err := s.productRepo.ProductSpecificationTemplateSlugExists(productSpecificationTemplate.Slug, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductSpecificationTemplateSlugExists
	}

	existingIDs := make(map[uint]struct{}, len(existing.SpecDefinitions))
	for _, definition := range existing.SpecDefinitions {
		existingIDs[definition.ID] = struct{}{}
	}
	retainedIDs := make(map[uint]struct{}, len(productSpecificationTemplate.SpecDefinitions))
	for _, definition := range productSpecificationTemplate.SpecDefinitions {
		if definition.ID == 0 {
			continue
		}
		if _, ok := existingIDs[definition.ID]; !ok {
			return nil, fmt.Errorf("%w: specification does not belong to product specification template", ErrProductSpecInvalid)
		}
		retainedIDs[definition.ID] = struct{}{}
	}

	removedIDs := make([]uint, 0)
	for definitionID := range existingIDs {
		if _, ok := retainedIDs[definitionID]; !ok {
			removedIDs = append(removedIDs, definitionID)
		}
	}
	sort.Slice(removedIDs, func(i, j int) bool { return removedIDs[i] < removedIDs[j] })

	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.UpdateProductSpecificationTemplate(productSpecificationTemplate, removedIDs); err != nil {
				if err == gorm.ErrRecordNotFound {
					return ErrProductSpecificationTemplateNotFound
				}
				return err
			}
			cacheEvents, _, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			return cacheEvents.EnqueueProductCacheInvalidateByProductSpecificationTemplateID(id, "admin product specification template update")
		})
		if err != nil {
			return nil, err
		}
		s.InvalidateProductCacheByProductSpecificationTemplateID(id)
		s.invalidateStorefrontHTMLCache("admin product specification template update")
		return s.productRepo.FindProductSpecificationTemplateByID(id)
	}

	if err := s.productRepo.UpdateProductSpecificationTemplate(productSpecificationTemplate, removedIDs); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}
	s.InvalidateProductCacheByProductSpecificationTemplateID(id)
	if err := s.enqueueProductCacheInvalidationByProductSpecificationTemplateID(id, "admin product specification template update"); err != nil {
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin product specification template update")
	return s.productRepo.FindProductSpecificationTemplateByID(id)
}

func (s *ProductService) DeleteProductSpecificationTemplate(id uint) error {
	existing, err := s.GetProductSpecificationTemplate(id)
	if err != nil {
		return err
	}
	if existing.IsSystemManaged {
		return fmt.Errorf("%w: product specification template cannot be deleted", ErrProductSpecificationTemplateSystemManaged)
	}
	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.DeleteProductSpecificationTemplate(id); err != nil {
				if err == gorm.ErrRecordNotFound {
					return ErrProductSpecificationTemplateNotFound
				}
				return err
			}
			cacheEvents, _, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			return cacheEvents.EnqueueProductCacheInvalidateByProductSpecificationTemplateID(id, "admin product specification template delete")
		})
		if err != nil {
			return err
		}
		s.InvalidateProductCacheByProductSpecificationTemplateID(id)
		s.invalidateStorefrontHTMLCache("admin product specification template delete")
		return nil
	}

	s.InvalidateProductCacheByProductSpecificationTemplateID(id)
	if err := s.enqueueProductCacheInvalidationByProductSpecificationTemplateID(id, "admin product specification template delete"); err != nil {
		return err
	}
	if err := s.productRepo.DeleteProductSpecificationTemplate(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProductSpecificationTemplateNotFound
		}
		return err
	}
	s.invalidateStorefrontHTMLCache("admin product specification template delete")
	return nil
}

func validateSystemManagedProductSpecificationTemplate(existing, next *product.ProductSpecificationTemplate) error {
	if existing == nil || next == nil {
		return fmt.Errorf("%w: product specification template is missing", ErrProductSpecificationTemplateSystemManaged)
	}

	definitionsByID := make(map[uint]product.SpecDefinition, len(next.SpecDefinitions))
	for _, definition := range next.SpecDefinitions {
		if definition.ID == 0 {
			return fmt.Errorf("%w: system product specification template fields cannot be added", ErrProductSpecificationTemplateSystemManaged)
		}
		definitionsByID[definition.ID] = definition
	}
	if len(definitionsByID) != len(existing.SpecDefinitions) {
		return fmt.Errorf("%w: system product specification template fields cannot be added or removed", ErrProductSpecificationTemplateSystemManaged)
	}

	for _, previous := range existing.SpecDefinitions {
		current, ok := definitionsByID[previous.ID]
		if !ok {
			return fmt.Errorf("%w: system product specification template fields cannot be added or removed", ErrProductSpecificationTemplateSystemManaged)
		}
		if previous.Slug != current.Slug ||
			previous.FieldType != current.FieldType ||
			previous.Presentation != current.Presentation ||
			previous.IsFilterable != current.IsFilterable ||
			previous.IsVariantOption != current.IsVariantOption ||
			normalizedSpecOptionsForComparison(previous.Options) != normalizedSpecOptionsForComparison(current.Options) ||
			previous.Validation != current.Validation {
			return fmt.Errorf("%w: field %q structure is immutable", ErrProductSpecificationTemplateSystemManaged, previous.Slug)
		}
	}
	return nil
}

func normalizedSpecOptionsForComparison(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "[]"
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return raw
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return raw
	}
	return string(encoded)
}

func normalizeProductSpecificationTemplateInput(input ProductSpecificationTemplateInput) (*product.ProductSpecificationTemplate, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if name == "" || slug == "" {
		return nil, fmt.Errorf("%w: name and slug are required", ErrProductSpecificationTemplateInvalid)
	}
	if len(name) > 120 || len(slug) > 120 || !productSpecificationTemplateSlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("%w: slug must use lowercase letters, numbers, dashes, or underscores", ErrProductSpecificationTemplateInvalid)
	}

	definitions := make([]product.SpecDefinition, 0, len(input.SpecDefinitions))
	seenIDs := make(map[uint]struct{}, len(input.SpecDefinitions))
	seenSlugs := make(map[string]struct{}, len(input.SpecDefinitions))
	for index, item := range input.SpecDefinitions {
		definition, err := normalizeSpecDefinition(item, index)
		if err != nil {
			return nil, err
		}
		if definition.ID > 0 {
			if _, exists := seenIDs[definition.ID]; exists {
				return nil, fmt.Errorf("%w: duplicate specification id", ErrProductSpecInvalid)
			}
			seenIDs[definition.ID] = struct{}{}
		}
		if _, exists := seenSlugs[definition.Slug]; exists {
			return nil, fmt.Errorf("%w: duplicate specification slug %q", ErrProductSpecInvalid, definition.Slug)
		}
		seenSlugs[definition.Slug] = struct{}{}
		definitions = append(definitions, definition)
	}

	return &product.ProductSpecificationTemplate{
		Name:            name,
		Slug:            slug,
		Description:     strings.TrimSpace(input.Description),
		SortOrder:       input.SortOrder,
		IsEnabled:       input.IsEnabled,
		SpecDefinitions: definitions,
	}, nil
}

func normalizeSpecDefinition(input ProductSpecDefinitionInput, index int) (product.SpecDefinition, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	fieldType := strings.ToLower(strings.TrimSpace(input.FieldType))
	if fieldType == "" {
		fieldType = "text"
	}
	if name == "" || slug == "" {
		return product.SpecDefinition{}, fmt.Errorf("%w: specification %d requires name and slug", ErrProductSpecInvalid, index+1)
	}
	if len(name) > 120 || len(slug) > 120 || !productSpecificationTemplateSlugPattern.MatchString(slug) {
		return product.SpecDefinition{}, fmt.Errorf("%w: invalid specification slug %q", ErrProductSpecInvalid, slug)
	}
	if fieldType != "text" && fieldType != "number" && fieldType != "select" && fieldType != "boolean" {
		return product.SpecDefinition{}, fmt.Errorf("%w: unsupported field type %q", ErrProductSpecInvalid, fieldType)
	}

	presentation := strings.ToLower(strings.TrimSpace(input.Presentation))
	if presentation == "" {
		presentation = "text"
	}
	if presentation != "text" && presentation != "color" && presentation != "image" {
		return product.SpecDefinition{}, fmt.Errorf("%w: unsupported presentation %q", ErrProductSpecInvalid, presentation)
	}
	if presentation != "text" && (!input.IsVariantOption || fieldType != "select") {
		return product.SpecDefinition{}, fmt.Errorf("%w: color/image presentation requires a select SKU option", ErrProductSpecInvalid)
	}

	options := strings.TrimSpace(input.Options)
	if fieldType == "select" {
		var values []string
		if options != "" {
			if err := json.Unmarshal([]byte(options), &values); err != nil {
				return product.SpecDefinition{}, fmt.Errorf("%w: select specification %q requires valid options", ErrProductSpecInvalid, slug)
			}
		}
		cleaned := make([]string, 0, len(values))
		seen := make(map[string]struct{}, len(values))
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			cleaned = append(cleaned, value)
		}
		encoded, _ := json.Marshal(cleaned)
		options = string(encoded)
	} else {
		options = ""
	}

	validation := strings.TrimSpace(input.Validation)
	if validation != "" && !json.Valid([]byte(validation)) {
		return product.SpecDefinition{}, fmt.Errorf("%w: validation for %q must be valid JSON", ErrProductSpecInvalid, slug)
	}

	group := strings.TrimSpace(input.Group)
	if group == "" {
		group = "规格"
	}
	return product.SpecDefinition{
		ID:              input.ID,
		Group:           group,
		Name:            name,
		Slug:            slug,
		FieldType:       fieldType,
		Presentation:    presentation,
		Unit:            strings.TrimSpace(input.Unit),
		IsRequired:      input.IsRequired,
		IsFilterable:    input.IsFilterable,
		IsVisible:       input.IsVisible,
		IsVariantOption: input.IsVariantOption,
		SortOrder:       input.SortOrder,
		Options:         options,
		Validation:      validation,
	}, nil
}
