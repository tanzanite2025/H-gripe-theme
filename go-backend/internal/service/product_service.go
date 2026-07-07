package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/repository"
	"time"
)

type ProductService struct {
	productRepo *repository.ProductRepository
	cache       *cache.RedisCache
	cacheTTL    time.Duration
}

func NewProductService(productRepo *repository.ProductRepository, cache *cache.RedisCache, cacheTTL int) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		cache:       cache,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
	}
}

var (
	ErrProductNotFound       = errors.New("product not found")
	ErrProductSKUExists      = errors.New("product sku already exists")
	ErrProductTypeNotFound   = errors.New("product type not found")
	ErrProductSpecInvalid    = errors.New("product spec invalid")
	ErrProductVariantInvalid = errors.New("product variant invalid")
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

type ProductCreateInput struct {
	ProductTypeID *uint
	Name          string
	Slug          string
	Description   string
	ShortDesc     string
	Weight        int
	Status        string
	Locale        string
	ParentID      *uint
	Featured      bool
	MetaTitle     string
	MetaDesc      string
	SpecValues    map[string]string
	Variants      []ProductVariantInput
}

type ProductUpdateInput struct {
	ProductTypeID       *uint
	UpdateProductTypeID bool
	Name                *string
	Slug                *string
	Description         *string
	ShortDesc           *string
	Weight              *int
	Status              *string
	Locale              *string
	ParentID            *uint
	UpdateParentID      bool
	Featured            *bool
	MetaTitle           *string
	MetaDesc            *string
	SpecValues          map[string]string
	UpdateSpecValues    bool
	Variants            []ProductVariantInput
	UpdateVariants      bool
}

type ProductSearchInput struct {
	Locale      string
	Status      string
	Keyword     string
	TypeSlug    string
	PriceMin    *float64
	PriceMax    *float64
	SpecFilters map[string][]string
	Page        int
	PageSize    int
}

func (s *ProductService) GetByID(id uint) (*product.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	var cachedProduct product.Product
	if s.cache != nil && s.cache.Get(cacheKey, &cachedProduct) == nil {
		return &cachedProduct, nil
	}

	result, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	_ = s.productRepo.IncrementViewCount(id)

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *ProductService) GetBySlug(slug, locale string) (*product.Product, error) {
	cacheKey := fmt.Sprintf("product:slug:%s:%s", slug, locale)

	var cachedProduct product.Product
	if s.cache != nil && s.cache.Get(cacheKey, &cachedProduct) == nil {
		return &cachedProduct, nil
	}

	result, err := s.productRepo.FindBySlug(slug, locale)
	if err != nil {
		return nil, err
	}

	_ = s.productRepo.IncrementViewCount(result.ID)

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *ProductService) List(locale, status string, featured bool, page, pageSize int) ([]product.Product, int64, error) {
	offset := (page - 1) * pageSize
	return s.productRepo.List(locale, status, featured, offset, pageSize)
}

func (s *ProductService) SearchPublic(input ProductSearchInput) ([]product.Product, int64, error) {
	page := input.Page
	pageSize := input.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.productRepo.SearchPublic(repository.ProductSearchQuery{
		Locale:      input.Locale,
		Status:      input.Status,
		Keyword:     input.Keyword,
		TypeSlug:    input.TypeSlug,
		PriceMin:    input.PriceMin,
		PriceMax:    input.PriceMax,
		SpecFilters: input.SpecFilters,
		Offset:      offset,
		Limit:       pageSize,
	})
}

func (s *ProductService) ListAdmin(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	return s.productRepo.FindAllWithFilters(page, pageSize, status, locale, search, featured)
}

func (s *ProductService) GetAdminProduct(id uint) (*product.Product, error) {
	return s.findProduct(id)
}

func (s *ProductService) GetStats() (map[string]interface{}, error) {
	return s.productRepo.GetStats()
}

func (s *ProductService) Create(p *product.Product) error {
	return s.productRepo.Create(p)
}

func (s *ProductService) CreateAdminProduct(input ProductCreateInput) (*product.Product, error) {
	specValues, err := s.buildSpecValues(input.ProductTypeID, input.SpecValues)
	if err != nil {
		return nil, err
	}

	variants, err := s.buildVariants(input.ProductTypeID, input.Variants)
	if err != nil {
		return nil, err
	}
	if err := s.ensureVariantSKUsAvailable(variants, 0); err != nil {
		return nil, err
	}

	newProduct := &product.Product{
		ProductTypeID: input.ProductTypeID,
		SKU:           defaultVariantSKU(variants),
		Name:          input.Name,
		Slug:          input.Slug,
		Description:   input.Description,
		ShortDesc:     input.ShortDesc,
		Weight:        input.Weight,
		Status:        input.Status,
		Locale:        input.Locale,
		ParentID:      input.ParentID,
		Featured:      input.Featured,
		MetaTitle:     input.MetaTitle,
		MetaDesc:      input.MetaDesc,
	}

	if err := s.productRepo.CreateWithSpecValuesAndVariants(newProduct, specValues, variants); err != nil {
		return nil, err
	}

	return s.findProduct(newProduct.ID)
}

func (s *ProductService) Update(p *product.Product) error {
	previousProduct, err := s.findProduct(p.ID)
	if err != nil {
		return err
	}

	if err := s.productRepo.Update(p); err != nil {
		return err
	}

	s.clearProductCache(previousProduct)
	s.clearProductCache(p)

	return nil
}

func (s *ProductService) UpdateAdminProduct(id uint, input ProductUpdateInput) (*product.Product, error) {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}

	previousProduct := *existingProduct

	if input.UpdateProductTypeID {
		existingProduct.ProductTypeID = input.ProductTypeID
	}
	if input.Name != nil {
		existingProduct.Name = *input.Name
	}
	if input.Slug != nil {
		existingProduct.Slug = *input.Slug
	}
	if input.Description != nil {
		existingProduct.Description = *input.Description
	}
	if input.ShortDesc != nil {
		existingProduct.ShortDesc = *input.ShortDesc
	}
	if input.Weight != nil {
		existingProduct.Weight = *input.Weight
	}
	if input.Status != nil {
		existingProduct.Status = *input.Status
	}
	if input.Locale != nil {
		existingProduct.Locale = *input.Locale
	}
	if input.UpdateParentID {
		existingProduct.ParentID = input.ParentID
	}
	if input.Featured != nil {
		existingProduct.Featured = *input.Featured
	}
	if input.MetaTitle != nil {
		existingProduct.MetaTitle = *input.MetaTitle
	}
	if input.MetaDesc != nil {
		existingProduct.MetaDesc = *input.MetaDesc
	}

	var specValues []product.ProductSpecValue
	if input.UpdateSpecValues {
		specValues, err = s.buildSpecValues(existingProduct.ProductTypeID, input.SpecValues)
		if err != nil {
			return nil, err
		}
	}

	var variants []product.ProductVariant
	if input.UpdateVariants {
		variants, err = s.buildVariants(existingProduct.ProductTypeID, input.Variants)
		if err != nil {
			return nil, err
		}
		if err := s.ensureVariantSKUsAvailable(variants, existingProduct.ID); err != nil {
			return nil, err
		}
	}

	if err := s.productRepo.UpdateWithSpecValuesAndVariants(existingProduct, specValues, input.UpdateSpecValues, variants, input.UpdateVariants); err != nil {
		return nil, err
	}

	s.clearProductCache(&previousProduct)
	s.clearProductCache(existingProduct)

	return s.findProduct(existingProduct.ID)
}

func (s *ProductService) Delete(id uint) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)

	return nil
}

func (s *ProductService) UpdateStatus(id uint, status string) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)

	return nil
}

func (s *ProductService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.UpdateStatus(id, status); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}

	return updated, nil
}

func (s *ProductService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}

func (s *ProductService) findProduct(id uint) (*product.Product, error) {
	result, err := s.productRepo.FindByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return result, nil
}

func (s *ProductService) clearProductCache(p *product.Product) {
	if s.cache == nil || p == nil {
		return
	}

	_ = s.cache.Delete(fmt.Sprintf("product:%d", p.ID))
	if p.Slug != "" {
		_ = s.cache.Delete(fmt.Sprintf("product:slug:%s:%s", p.Slug, p.Locale))
	}
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
