package service

import (
	"errors"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/repository"
	"time"
)

type ProductService struct {
	productRepo                    *repository.ProductRepository
	cache                          *cache.RedisCache
	cacheTTL                       time.Duration
	storefrontHTMLCacheInvalidator *StorefrontHTMLCacheInvalidator
}

func NewProductService(productRepo *repository.ProductRepository, cache *cache.RedisCache, cacheTTL int) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		cache:       cache,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
	}
}

func (s *ProductService) SetStorefrontHTMLCacheInvalidator(invalidator *StorefrontHTMLCacheInvalidator) {
	s.storefrontHTMLCacheInvalidator = invalidator
}

var (
	ErrProductNotFound       = errors.New("product not found")
	ErrProductSKUExists      = errors.New("product sku already exists")
	ErrProductTypeNotFound   = errors.New("product type not found")
	ErrProductTypeInvalid    = errors.New("product type invalid")
	ErrProductTypeSlugExists = errors.New("product type slug already exists")
	ErrProductSpecInvalid    = errors.New("product spec invalid")
	ErrProductVariantInvalid = errors.New("product variant invalid")
	ErrProductMediaInvalid   = errors.New("product media invalid")
)

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
	cacheKey := productIDCacheKey(id)

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
	cacheKey := productSlugCacheKey(slug, locale)

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
	products, total, err := s.productRepo.List(locale, status, featured, offset, pageSize)
	if err == nil && total == 0 && locale != "" && locale != "en" {
		return s.productRepo.List("en", status, featured, offset, pageSize)
	}
	return products, total, err
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
	query := repository.ProductSearchQuery{
		Locale:      input.Locale,
		Status:      input.Status,
		Keyword:     input.Keyword,
		TypeSlug:    input.TypeSlug,
		PriceMin:    input.PriceMin,
		PriceMax:    input.PriceMax,
		SpecFilters: input.SpecFilters,
		Offset:      offset,
		Limit:       pageSize,
	}
	products, total, err := s.productRepo.SearchPublic(query)
	if err == nil && total == 0 && input.Locale != "" && input.Locale != "en" {
		query.Locale = "en"
		return s.productRepo.SearchPublic(query)
	}
	return products, total, err
}

func (s *ProductService) Create(p *product.Product) error {
	if err := s.productRepo.Create(p); err != nil {
		return err
	}
	s.invalidateStorefrontHTMLCache("product create")
	return nil
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
	s.invalidateStorefrontHTMLCache("product update")

	return nil
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
