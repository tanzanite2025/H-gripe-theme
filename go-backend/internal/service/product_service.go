package service

import (
	"encoding/json"
	"errors"
	"fmt"
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
	ErrProductNotFound     = errors.New("product not found")
	ErrProductSKUExists    = errors.New("product sku already exists")
	ErrProductTypeNotFound = errors.New("product type not found")
	ErrProductSpecInvalid  = errors.New("product spec invalid")
)

type ProductCreateInput struct {
	ProductTypeID *uint
	SKU           string
	Name          string
	Slug          string
	Description   string
	ShortDesc     string
	Price         float64
	SalePrice     *float64
	Stock         int
	Weight        int
	Status        string
	Locale        string
	ParentID      *uint
	Featured      bool
	MetaTitle     string
	MetaDesc      string
	SpecValues    map[string]string
}

type ProductUpdateInput struct {
	ProductTypeID       *uint
	UpdateProductTypeID bool
	SKU                 *string
	Name                *string
	Slug                *string
	Description         *string
	ShortDesc           *string
	Price               *float64
	SalePrice           *float64
	UpdateSalePrice     bool
	Stock               *int
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
	if err := s.ensureSKUAvailable(input.SKU, 0); err != nil {
		return nil, err
	}

	specValues, err := s.buildSpecValues(input.ProductTypeID, input.SpecValues)
	if err != nil {
		return nil, err
	}

	newProduct := &product.Product{
		ProductTypeID: input.ProductTypeID,
		SKU:           input.SKU,
		Name:          input.Name,
		Slug:          input.Slug,
		Description:   input.Description,
		ShortDesc:     input.ShortDesc,
		Price:         input.Price,
		SalePrice:     input.SalePrice,
		Stock:         input.Stock,
		Weight:        input.Weight,
		Status:        input.Status,
		Locale:        input.Locale,
		ParentID:      input.ParentID,
		Featured:      input.Featured,
		MetaTitle:     input.MetaTitle,
		MetaDesc:      input.MetaDesc,
	}

	if err := s.productRepo.CreateWithSpecValues(newProduct, specValues); err != nil {
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
	if input.SKU != nil && *input.SKU != existingProduct.SKU {
		if err := s.ensureSKUAvailable(*input.SKU, existingProduct.ID); err != nil {
			return nil, err
		}
		existingProduct.SKU = *input.SKU
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
	if input.Price != nil {
		existingProduct.Price = *input.Price
	}
	if input.UpdateSalePrice {
		existingProduct.SalePrice = input.SalePrice
	}
	if input.Stock != nil {
		existingProduct.Stock = *input.Stock
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

	if err := s.productRepo.UpdateWithSpecValues(existingProduct, specValues, input.UpdateSpecValues); err != nil {
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

func (s *ProductService) UpdateStock(id uint, quantity int) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.UpdateStock(id, quantity); err != nil {
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

func (s *ProductService) ensureSKUAvailable(sku string, currentProductID uint) error {
	existingProduct, err := s.productRepo.FindBySKU(sku)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	if existingProduct != nil && existingProduct.ID != currentProductID {
		return ErrProductSKUExists
	}

	return nil
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
