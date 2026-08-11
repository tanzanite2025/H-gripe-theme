package service

import (
	"commerce-platform/internal/domain/product"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

var productTypeSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

type ProductTypeInput struct {
	Name               string
	Slug               string
	Description        string
	SortOrder          int
	IsEnabled          bool
	Translations       []ProductTypeTranslationInput
	UpdateTranslations bool
	SpecDefinitions    []ProductSpecDefinitionInput
}

type ProductTypeTranslationInput struct {
	ID          uint
	Locale      string
	Name        string
	Description string
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

func (s *ProductService) ListProductTypes(includeDisabled bool) ([]product.ProductType, error) {
	return s.productRepo.FindAllProductTypes(includeDisabled)
}

func (s *ProductService) ListPublicProductTypes(includeDisabled bool) ([]product.ProductType, error) {
	return s.productRepo.FindPublicProductTypes(includeDisabled)
}

func (s *ProductService) GetProductType(id uint) (*product.ProductType, error) {
	productType, err := s.productRepo.FindProductTypeByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProductTypeNotFound
		}
		return nil, err
	}
	return productType, nil
}

func (s *ProductService) UpdateProductTypeImage(id uint, mediaAssetID *uint, imageURL string) (*product.ProductType, error) {
	existing, err := s.GetProductType(id)
	if err != nil {
		return nil, err
	}
	previousAssetID := existing.ImageMediaAssetID

	imageURL = strings.TrimSpace(imageURL)
	if mediaAssetID != nil && *mediaAssetID == 0 {
		mediaAssetID = nil
	}
	if mediaAssetID == nil && imageURL != "" {
		return nil, fmt.Errorf("%w: image asset is required when an image URL is provided", ErrProductTypeInvalid)
	}
	if mediaAssetID != nil && imageURL == "" {
		return nil, fmt.Errorf("%w: image URL is required when an image asset is selected", ErrProductTypeInvalid)
	}

	if err := s.productRepo.UpdateProductTypeImage(id, mediaAssetID, imageURL); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProductTypeNotFound
		}
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin product type image update")
	if previousAssetID != nil && (mediaAssetID == nil || *mediaAssetID != *previousAssetID) {
		s.cleanupProductTypeImageAsset(*previousAssetID, "product type image replacement or removal")
	}
	return s.productRepo.FindProductTypeByID(id)
}

func (s *ProductService) CreateProductType(input ProductTypeInput) (*product.ProductType, error) {
	productType, err := normalizeProductTypeInput(input)
	if err != nil {
		return nil, err
	}

	exists, err := s.productRepo.ProductTypeSlugExists(productType.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductTypeSlugExists
	}

	if err := s.productRepo.CreateProductType(productType); err != nil {
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin product type create")
	return s.productRepo.FindProductTypeByID(productType.ID)
}

func (s *ProductService) UpdateProductType(id uint, input ProductTypeInput) (*product.ProductType, error) {
	existing, err := s.GetProductType(id)
	if err != nil {
		return nil, err
	}

	productType, err := normalizeProductTypeInput(input)
	if err != nil {
		return nil, err
	}
	productType.ID = id

	exists, err := s.productRepo.ProductTypeSlugExists(productType.Slug, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductTypeSlugExists
	}

	existingIDs := make(map[uint]struct{}, len(existing.SpecDefinitions))
	for _, definition := range existing.SpecDefinitions {
		existingIDs[definition.ID] = struct{}{}
	}
	retainedIDs := make(map[uint]struct{}, len(productType.SpecDefinitions))
	for _, definition := range productType.SpecDefinitions {
		if definition.ID == 0 {
			continue
		}
		if _, ok := existingIDs[definition.ID]; !ok {
			return nil, fmt.Errorf("%w: specification does not belong to product type", ErrProductSpecInvalid)
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

	if err := s.productRepo.UpdateProductType(productType, removedIDs, input.UpdateTranslations); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProductTypeNotFound
		}
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin product type update")
	return s.productRepo.FindProductTypeByID(id)
}

func (s *ProductService) DeleteProductType(id uint) error {
	existing, err := s.GetProductType(id)
	if err != nil {
		return err
	}
	if err := s.productRepo.DeleteProductType(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProductTypeNotFound
		}
		return err
	}
	s.invalidateStorefrontHTMLCache("admin product type delete")
	if existing.ImageMediaAssetID != nil {
		s.cleanupProductTypeImageAsset(*existing.ImageMediaAssetID, "product type deletion")
	}
	return nil
}

func normalizeProductTypeInput(input ProductTypeInput) (*product.ProductType, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if name == "" || slug == "" {
		return nil, fmt.Errorf("%w: name and slug are required", ErrProductTypeInvalid)
	}
	if len(name) > 120 || len(slug) > 120 || !productTypeSlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("%w: slug must use lowercase letters, numbers, dashes, or underscores", ErrProductTypeInvalid)
	}

	translations, err := normalizeProductTypeTranslations(input.Translations)
	if err != nil {
		return nil, err
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

	return &product.ProductType{
		Name:            name,
		Slug:            slug,
		Description:     strings.TrimSpace(input.Description),
		SortOrder:       input.SortOrder,
		IsEnabled:       input.IsEnabled,
		Translations:    translations,
		SpecDefinitions: definitions,
	}, nil
}

func normalizeProductTypeTranslations(input []ProductTypeTranslationInput) ([]product.ProductTypeTranslation, error) {
	if input == nil {
		return nil, nil
	}

	result := make([]product.ProductTypeTranslation, 0, len(input))
	seenLocales := make(map[string]struct{}, len(input))
	for index, item := range input {
		locale, err := requireSupportedLocale(item.Locale)
		if err != nil {
			return nil, fmt.Errorf("%w: translation %d has invalid locale", ErrProductTypeTranslationInvalid, index+1)
		}

		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: translation %d requires a name", ErrProductTypeTranslationInvalid, index+1)
		}
		if len(name) > 120 {
			return nil, fmt.Errorf("%w: translation %d name is too long", ErrProductTypeTranslationInvalid, index+1)
		}
		if _, exists := seenLocales[locale]; exists {
			return nil, fmt.Errorf("%w: duplicate locale %q", ErrProductTypeTranslationInvalid, locale)
		}
		seenLocales[locale] = struct{}{}

		result = append(result, product.ProductTypeTranslation{
			ID:          item.ID,
			Locale:      locale,
			Name:        name,
			Description: strings.TrimSpace(item.Description),
		})
	}

	return result, nil
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
	if len(name) > 120 || len(slug) > 120 || !productTypeSlugPattern.MatchString(slug) {
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
		if len(cleaned) == 0 {
			if input.IsVariantOption && presentation != "text" {
				options = "[]"
			} else {
				return product.SpecDefinition{}, fmt.Errorf("%w: select specification %q requires options", ErrProductSpecInvalid, slug)
			}
		} else {
			encoded, _ := json.Marshal(cleaned)
			options = string(encoded)
		}
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
