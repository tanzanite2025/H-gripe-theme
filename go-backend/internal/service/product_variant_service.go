package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ProductVariantInput struct {
	ID                 *uint
	ShippingTemplateID *uint
	SKU                string
	Title              string
	OptionValues       map[string]string
	Currency           string
	PriceMinor         int64
	SalePriceMinor     *int64
	Price              float64
	SalePrice          *float64
	DisplayPrices      []currency.DisplayPriceSnapshot
	Stock              int
	Weight             int
	IsDefault          bool
	IsActive           *bool
	SortOrder          int
	OptionGroupRules   []product.ProductOptionGroupVariantRule
	OptionValueRules   []product.ProductOptionValueVariantRule
}

func (s *ProductService) buildSpecValues(productSpecificationTemplateID *uint, values map[string]string) ([]product.ProductSpecValue, error) {
	if productSpecificationTemplateID == nil {
		if len(values) > 0 {
			return nil, fmt.Errorf("%w: product_specification_template_id is required when specs are provided", ErrProductSpecInvalid)
		}
		return nil, nil
	}

	productSpecificationTemplate, err := s.productRepo.FindProductSpecificationTemplateByID(*productSpecificationTemplateID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}

	definitionsBySlug := make(map[string]product.SpecDefinition, len(productSpecificationTemplate.SpecDefinitions))
	for _, definition := range productSpecificationTemplate.SpecDefinitions {
		definitionsBySlug[definition.Slug] = definition
	}

	normalizedValues := make(map[string]string, len(values))
	for slug, raw := range values {
		definition, ok := definitionsBySlug[slug]
		if !ok {
			return nil, fmt.Errorf("%w: unknown spec %s", ErrProductSpecInvalid, slug)
		}
		if definition.RuntimeRole() != "attribute" {
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
	for _, definition := range productSpecificationTemplate.SpecDefinitions {
		if definition.RuntimeRole() != "attribute" {
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

func (s *ProductService) buildVariants(productSpecificationTemplateID *uint, inputs []ProductVariantInput, productCurrency string, optionDisplayValues []product.ProductVariantOptionValue) ([]product.ProductVariant, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%w: at least one variant is required", ErrProductVariantInvalid)
	}

	variantDefinitions, err := s.loadVariantDefinitions(productSpecificationTemplateID)
	if err != nil {
		return nil, err
	}
	configuredOptionValues := variantOptionValueKeysBySlug(variantDefinitions, optionDisplayValues)
	customDefinitionIDs, err := s.loadCustomOptionDefinitionIDs(productSpecificationTemplateID)
	if err != nil {
		return nil, err
	}
	optionValuesByID := make(map[uint]product.ProductVariantOptionValue, len(optionDisplayValues))
	for _, optionValue := range optionDisplayValues {
		optionValuesByID[optionValue.ID] = optionValue
	}

	variants := make([]product.ProductVariant, 0, len(inputs))
	defaultIndex := -1
	seenSKU := make(map[string]struct{}, len(inputs))
	seenOptions := make(map[string]struct{}, len(inputs))
	productCurrency = normalizeStoredProductPriceCurrency(productCurrency)
	if !currency.IsCatalogCode(productCurrency) {
		return nil, fmt.Errorf("%w: unsupported product currency", ErrProductVariantInvalid)
	}

	for i, input := range inputs {
		variantCurrency := currency.NormalizeCode(input.Currency)
		if variantCurrency == "" {
			variantCurrency = productCurrency
		}
		if variantCurrency != productCurrency {
			return nil, fmt.Errorf("%w: all variants must use product currency %s", ErrProductVariantInvalid, productCurrency)
		}
		input.SKU = strings.TrimSpace(input.SKU)
		if input.SKU == "" {
			return nil, fmt.Errorf("%w: sku is required", ErrProductVariantInvalid)
		}
		if input.PriceMinor == 0 && input.Price != 0 {
			priceMoney, conversionErr := domainmoney.FromMajorFloat(input.Price, variantCurrency)
			if conversionErr != nil {
				return nil, fmt.Errorf("%w: invalid price for %s", ErrProductVariantInvalid, input.SKU)
			}
			input.PriceMinor = priceMoney.AmountMinor()
		}
		if input.PriceMinor <= 0 {
			return nil, fmt.Errorf("%w: price must be greater than zero for %s", ErrProductVariantInvalid, input.SKU)
		}
		if input.Stock < 0 {
			return nil, fmt.Errorf("%w: stock cannot be negative for %s", ErrProductVariantInvalid, input.SKU)
		}
		if input.SalePriceMinor == nil && input.SalePrice != nil {
			saleMoney, conversionErr := domainmoney.FromMajorFloat(*input.SalePrice, variantCurrency)
			if conversionErr != nil {
				return nil, fmt.Errorf("%w: invalid sale_price for %s", ErrProductVariantInvalid, input.SKU)
			}
			saleMinor := saleMoney.AmountMinor()
			input.SalePriceMinor = &saleMinor
		}
		if input.SalePriceMinor != nil && *input.SalePriceMinor < 0 {
			return nil, fmt.Errorf("%w: sale_price cannot be negative for %s", ErrProductVariantInvalid, input.SKU)
		}
		priceMajor, conversionErr := domainmoney.New(input.PriceMinor, variantCurrency)
		if conversionErr != nil {
			return nil, fmt.Errorf("%w: invalid price for %s", ErrProductVariantInvalid, input.SKU)
		}
		priceFloat, conversionErr := priceMajor.MajorFloat()
		if conversionErr != nil {
			return nil, fmt.Errorf("%w: invalid price for %s", ErrProductVariantInvalid, input.SKU)
		}
		input.Price = priceFloat
		if input.SalePriceMinor != nil {
			saleMoney, saleErr := domainmoney.New(*input.SalePriceMinor, variantCurrency)
			if saleErr != nil {
				return nil, fmt.Errorf("%w: invalid sale_price for %s", ErrProductVariantInvalid, input.SKU)
			}
			saleFloat, saleErr := saleMoney.MajorFloat()
			if saleErr != nil {
				return nil, fmt.Errorf("%w: invalid sale_price for %s", ErrProductVariantInvalid, input.SKU)
			}
			input.SalePrice = &saleFloat
		}
		skuKey := strings.ToLower(input.SKU)
		if _, exists := seenSKU[skuKey]; exists {
			return nil, fmt.Errorf("%w: duplicate sku %s", ErrProductVariantInvalid, input.SKU)
		}
		seenSKU[skuKey] = struct{}{}

		optionValues, optionJSON, err := s.normalizeVariantOptions(variantDefinitions, input.OptionValues, configuredOptionValues)
		if err != nil {
			return nil, err
		}
		if err := validateVariantOptionRules(input.OptionGroupRules, input.OptionValueRules, customDefinitionIDs, optionValuesByID); err != nil {
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
			ShippingTemplateID: input.ShippingTemplateID,
			SKU:                input.SKU,
			Title:              strings.TrimSpace(input.Title),
			OptionValues:       optionJSON,
			Currency:           variantCurrency,
			PriceMinor:         input.PriceMinor,
			SalePriceMinor:     input.SalePriceMinor,
			Price:              input.Price,
			SalePrice:          input.SalePrice,
			DisplayPriceData:   currency.DisplayPriceSnapshotsJSON(input.DisplayPrices, variantCurrency),
			Stock:              input.Stock,
			Weight:             input.Weight,
			IsDefault:          input.IsDefault,
			IsActive:           isActive,
			SortOrder:          input.SortOrder,
			OptionGroupRules:   input.OptionGroupRules,
			OptionValueRules:   input.OptionValueRules,
		}
		if input.ID != nil {
			variant.ID = *input.ID
		}
		if variant.Title == "" {
			variant.Title = variantTitle(optionValues)
		}
		variants = append(variants, variant)
	}

	activeIndex := -1
	for i := range variants {
		if variants[i].IsActive {
			activeIndex = i
			break
		}
	}
	if activeIndex == -1 {
		return nil, fmt.Errorf("%w: at least one sku variant must be enabled", ErrProductVariantInvalid)
	}
	if defaultIndex == -1 || !variants[defaultIndex].IsActive {
		defaultIndex = activeIndex
	}
	for i := range variants {
		variants[i].IsDefault = i == defaultIndex
	}

	return variants, nil
}

func (s *ProductService) loadCustomOptionDefinitionIDs(productSpecificationTemplateID *uint) (map[uint]struct{}, error) {
	result := make(map[uint]struct{})
	if productSpecificationTemplateID == nil {
		return result, nil
	}
	template, err := s.productRepo.FindProductSpecificationTemplateByID(*productSpecificationTemplateID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}
	for _, definition := range template.SpecDefinitions {
		if definition.RuntimeRole() == "custom_option" {
			result[definition.ID] = struct{}{}
		}
	}
	return result, nil
}

func validateVariantOptionRules(
	groupRules []product.ProductOptionGroupVariantRule,
	valueRules []product.ProductOptionValueVariantRule,
	customDefinitionIDs map[uint]struct{},
	optionValuesByID map[uint]product.ProductVariantOptionValue,
) error {
	seenGroups := make(map[uint]struct{}, len(groupRules))
	for _, rule := range groupRules {
		if rule.SpecDefinitionID == 0 {
			return fmt.Errorf("%w: option group rule definition is required", ErrProductVariantInvalid)
		}
		if _, ok := customDefinitionIDs[rule.SpecDefinitionID]; !ok {
			return fmt.Errorf("%w: option group rule references a non-custom definition", ErrProductVariantInvalid)
		}
		if _, ok := seenGroups[rule.SpecDefinitionID]; ok {
			return fmt.Errorf("%w: duplicate option group rule", ErrProductVariantInvalid)
		}
		seenGroups[rule.SpecDefinitionID] = struct{}{}
		if rule.MinSelectionsOverride != nil && *rule.MinSelectionsOverride < 0 {
			return fmt.Errorf("%w: option group minimum cannot be negative", ErrProductVariantInvalid)
		}
		if rule.MaxSelectionsOverride != nil && *rule.MaxSelectionsOverride < 0 {
			return fmt.Errorf("%w: option group maximum cannot be negative", ErrProductVariantInvalid)
		}
		if rule.MinSelectionsOverride != nil && rule.MaxSelectionsOverride != nil && *rule.MaxSelectionsOverride < *rule.MinSelectionsOverride {
			return fmt.Errorf("%w: option group maximum cannot be below minimum", ErrProductVariantInvalid)
		}
	}
	seenValues := make(map[uint]struct{}, len(valueRules))
	for _, rule := range valueRules {
		if rule.ProductVariantOptionValueID == 0 {
			return fmt.Errorf("%w: option value rule value is required", ErrProductVariantInvalid)
		}
		optionValue, ok := optionValuesByID[rule.ProductVariantOptionValueID]
		if !ok {
			return fmt.Errorf("%w: option value rule references an unknown product option", ErrProductVariantInvalid)
		}
		if _, ok := customDefinitionIDs[optionValue.SpecDefinitionID]; !ok {
			return fmt.Errorf("%w: option value rule references a non-custom option", ErrProductVariantInvalid)
		}
		if _, ok := seenValues[rule.ProductVariantOptionValueID]; ok {
			return fmt.Errorf("%w: duplicate option value rule", ErrProductVariantInvalid)
		}
		seenValues[rule.ProductVariantOptionValueID] = struct{}{}
		if rule.PriceDeltaMinorOverride != nil && *rule.PriceDeltaMinorOverride < 0 {
			return fmt.Errorf("%w: option price override cannot be negative", ErrProductVariantInvalid)
		}
	}
	return nil
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

func (s *ProductService) loadVariantDefinitions(productSpecificationTemplateID *uint) (map[string]product.SpecDefinition, error) {
	if productSpecificationTemplateID == nil {
		return nil, nil
	}

	productSpecificationTemplate, err := s.productRepo.FindProductSpecificationTemplateByID(*productSpecificationTemplateID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductSpecificationTemplateNotFound
		}
		return nil, err
	}

	definitions := make(map[string]product.SpecDefinition)
	for _, definition := range productSpecificationTemplate.SpecDefinitions {
		if definition.RuntimeRole() == "variant" {
			definitions[definition.Slug] = definition
		}
	}
	return definitions, nil
}

func (s *ProductService) normalizeVariantOptions(definitions map[string]product.SpecDefinition, rawValues map[string]string, configuredOptionValues map[string]map[string]struct{}) (map[string]string, string, error) {
	if len(rawValues) > 0 && len(definitions) == 0 {
		return nil, "", fmt.Errorf("%w: product_specification_template_id is required when variant options are provided", ErrProductVariantInvalid)
	}

	normalizedValues := make(map[string]string, len(rawValues))
	for slug, raw := range rawValues {
		definition, ok := definitions[slug]
		if !ok {
			return nil, "", fmt.Errorf("%w: unknown variant option %s", ErrProductVariantInvalid, slug)
		}
		if allowedValues := configuredOptionValues[slug]; len(allowedValues) > 0 {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			if _, ok := allowedValues[value]; !ok {
				return nil, "", fmt.Errorf("%w: %s is not a configured option display value for %s", ErrProductVariantInvalid, value, slug)
			}
			normalizedValues[slug] = value
			continue
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

func variantOptionValueKeysBySlug(definitions map[string]product.SpecDefinition, values []product.ProductVariantOptionValue) map[string]map[string]struct{} {
	if len(definitions) == 0 || len(values) == 0 {
		return nil
	}

	definitionsByID := make(map[uint]product.SpecDefinition, len(definitions))
	for _, definition := range definitions {
		definitionsByID[definition.ID] = definition
	}

	result := make(map[string]map[string]struct{})
	for _, value := range values {
		definition, ok := definitionsByID[value.SpecDefinitionID]
		if !ok {
			continue
		}
		valueKey := strings.TrimSpace(value.ValueKey)
		if valueKey == "" {
			continue
		}
		if _, ok := result[definition.Slug]; !ok {
			result[definition.Slug] = make(map[string]struct{})
		}
		result[definition.Slug][valueKey] = struct{}{}
	}
	return result
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
		options := make([]string, 0, len(definition.OptionItems))
		for _, item := range definition.OptionItems {
			if key := strings.TrimSpace(item.ValueKey); key != "" {
				options = append(options, key)
			}
		}
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
