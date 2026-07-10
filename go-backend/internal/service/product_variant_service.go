package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"
)

type ProductVariantInput struct {
	ID           *uint
	SKU          string
	Title        string
	OptionValues map[string]string
	Price        float64
	SalePrice    *float64
	Stock        int
	Weight       int
	IsDefault    bool
	IsActive     *bool
	SortOrder    int
}

func (s *ProductService) buildSpecValues(productTypeID *uint, values map[string]string) ([]product.ProductSpecValue, error) {
	if productTypeID == nil {
		if len(values) > 0 {
			return nil, fmt.Errorf("%w: product_type_id is required when specs are provided", ErrProductSpecInvalid)
		}
		return nil, nil
	}

	productType, err := s.productRepo.FindProductTypeByID(*productTypeID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductTypeNotFound
		}
		return nil, err
	}

	definitionsBySlug := make(map[string]product.SpecDefinition, len(productType.SpecDefinitions))
	for _, definition := range productType.SpecDefinitions {
		definitionsBySlug[definition.Slug] = definition
	}

	normalizedValues := make(map[string]string, len(values))
	for slug, raw := range values {
		definition, ok := definitionsBySlug[slug]
		if !ok {
			return nil, fmt.Errorf("%w: unknown spec %s", ErrProductSpecInvalid, slug)
		}
		if definition.IsVariantOption {
			return nil, fmt.Errorf("%w: spec %s belongs to product variants", ErrProductSpecInvalid, slug)
		}

		normalized, err := normalizeSpecValue(definition, raw)
		if err != nil {
			return nil, err
		}
		if normalized != "" {
			normalizedValues[slug] = normalized
		}
	}

	specValues := make([]product.ProductSpecValue, 0, len(normalizedValues))
	for _, definition := range productType.SpecDefinitions {
		if definition.IsVariantOption {
			continue
		}
		value := strings.TrimSpace(normalizedValues[definition.Slug])
		if value == "" {
			if definition.IsRequired {
				return nil, fmt.Errorf("%w: required spec %s is missing", ErrProductSpecInvalid, definition.Slug)
			}
			continue
		}

		specValues = append(specValues, product.ProductSpecValue{
			SpecDefinitionID: definition.ID,
			Value:            value,
		})
	}

	return specValues, nil
}

func (s *ProductService) buildVariants(productTypeID *uint, inputs []ProductVariantInput) ([]product.ProductVariant, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%w: at least one variant is required", ErrProductVariantInvalid)
	}

	variantDefinitions, err := s.loadVariantDefinitions(productTypeID)
	if err != nil {
		return nil, err
	}

	variants := make([]product.ProductVariant, 0, len(inputs))
	defaultIndex := -1
	seenSKU := make(map[string]struct{}, len(inputs))
	seenOptions := make(map[string]struct{}, len(inputs))

	for i, input := range inputs {
		input.SKU = strings.TrimSpace(input.SKU)
		if input.SKU == "" {
			return nil, fmt.Errorf("%w: sku is required", ErrProductVariantInvalid)
		}
		if input.Price <= 0 {
			return nil, fmt.Errorf("%w: price must be greater than zero for %s", ErrProductVariantInvalid, input.SKU)
		}
		if input.Stock < 0 {
			return nil, fmt.Errorf("%w: stock cannot be negative for %s", ErrProductVariantInvalid, input.SKU)
		}
		if input.SalePrice != nil && *input.SalePrice < 0 {
			return nil, fmt.Errorf("%w: sale_price cannot be negative for %s", ErrProductVariantInvalid, input.SKU)
		}

		skuKey := strings.ToLower(input.SKU)
		if _, exists := seenSKU[skuKey]; exists {
			return nil, fmt.Errorf("%w: duplicate sku %s", ErrProductVariantInvalid, input.SKU)
		}
		seenSKU[skuKey] = struct{}{}

		optionValues, optionJSON, err := s.normalizeVariantOptions(variantDefinitions, input.OptionValues)
		if err != nil {
			return nil, err
		}
		if _, exists := seenOptions[optionJSON]; exists {
			return nil, fmt.Errorf("%w: duplicate option combination %s", ErrProductVariantInvalid, optionJSON)
		}
		seenOptions[optionJSON] = struct{}{}

		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		if input.IsDefault {
			defaultIndex = i
		}

		variant := product.ProductVariant{
			SKU:          input.SKU,
			Title:        strings.TrimSpace(input.Title),
			OptionValues: optionJSON,
			Price:        input.Price,
			SalePrice:    input.SalePrice,
			Stock:        input.Stock,
			Weight:       input.Weight,
			IsDefault:    input.IsDefault,
			IsActive:     isActive,
			SortOrder:    input.SortOrder,
		}
		if input.ID != nil {
			variant.ID = *input.ID
		}
		if variant.Title == "" {
			variant.Title = variantTitle(optionValues)
		}
		variants = append(variants, variant)
	}

	if defaultIndex == -1 {
		defaultIndex = 0
	}
	for i := range variants {
		variants[i].IsDefault = i == defaultIndex
	}

	return variants, nil
}

func (s *ProductService) ensureVariantSKUsAvailable(variants []product.ProductVariant, currentProductID uint) error {
	for _, variant := range variants {
		existingProduct, err := s.productRepo.FindBySKU(variant.SKU)
		if err != nil && !repository.IsRecordNotFound(err) {
			return err
		}
		if err == nil && existingProduct.ID != currentProductID {
			return ErrProductSKUExists
		}

		existingVariant, err := s.productRepo.FindVariantBySKU(variant.SKU)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				continue
			}
			return err
		}
		if currentProductID == 0 || existingVariant.ProductID != currentProductID {
			return ErrProductSKUExists
		}
		if variant.ID != 0 && existingVariant.ID != variant.ID {
			return ErrProductSKUExists
		}
	}

	return nil
}

func defaultVariantSKU(variants []product.ProductVariant) string {
	if len(variants) == 0 {
		return ""
	}
	for _, variant := range variants {
		if variant.IsDefault {
			return variant.SKU
		}
	}
	return variants[0].SKU
}

func (s *ProductService) loadVariantDefinitions(productTypeID *uint) (map[string]product.SpecDefinition, error) {
	if productTypeID == nil {
		return nil, nil
	}

	productType, err := s.productRepo.FindProductTypeByID(*productTypeID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductTypeNotFound
		}
		return nil, err
	}

	definitions := make(map[string]product.SpecDefinition)
	for _, definition := range productType.SpecDefinitions {
		if definition.IsVariantOption {
			definitions[definition.Slug] = definition
		}
	}
	return definitions, nil
}

func (s *ProductService) normalizeVariantOptions(definitions map[string]product.SpecDefinition, rawValues map[string]string) (map[string]string, string, error) {
	if len(rawValues) > 0 && len(definitions) == 0 {
		return nil, "", fmt.Errorf("%w: product_type_id is required when variant options are provided", ErrProductVariantInvalid)
	}

	normalizedValues := make(map[string]string, len(rawValues))
	for slug, raw := range rawValues {
		definition, ok := definitions[slug]
		if !ok {
			return nil, "", fmt.Errorf("%w: unknown variant option %s", ErrProductVariantInvalid, slug)
		}
		normalized, err := normalizeSpecValue(definition, raw)
		if err != nil {
			return nil, "", err
		}
		if normalized != "" {
			normalizedValues[slug] = normalized
		}
	}

	for _, definition := range definitions {
		value := strings.TrimSpace(normalizedValues[definition.Slug])
		if value == "" && definition.IsRequired {
			return nil, "", fmt.Errorf("%w: required variant option %s is missing", ErrProductVariantInvalid, definition.Slug)
		}
	}

	if len(normalizedValues) == 0 {
		return normalizedValues, "{}", nil
	}

	encoded, err := json.Marshal(normalizedValues)
	if err != nil {
		return nil, "", err
	}
	return normalizedValues, string(encoded), nil
}

func variantTitle(values map[string]string) string {
	if len(values) == 0 {
		return "Default"
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, values[key])
	}
	return strings.Join(parts, " / ")
}

func boolPtr(value bool) *bool {
	return &value
}

func normalizeSpecValue(definition product.SpecDefinition, raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	switch strings.ToLower(definition.FieldType) {
	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return "", fmt.Errorf("%w: %s must be a number", ErrProductSpecInvalid, definition.Slug)
		}
		return value, nil
	case "boolean":
		switch strings.ToLower(value) {
		case "true", "1", "yes", "y":
			return "true", nil
		case "false", "0", "no", "n":
			return "false", nil
		default:
			return "", fmt.Errorf("%w: %s must be a boolean", ErrProductSpecInvalid, definition.Slug)
		}
	case "select":
		options := parseSpecOptions(definition.Options)
		if len(options) == 0 {
			return value, nil
		}
		for _, option := range options {
			if value == option {
				return value, nil
			}
		}
		return "", fmt.Errorf("%w: %s is not a valid option for %s", ErrProductSpecInvalid, value, definition.Slug)
	default:
		return value, nil
	}
}

func parseSpecOptions(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var options []string
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return nil
	}
	return options
}
