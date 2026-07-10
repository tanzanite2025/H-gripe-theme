package service

import (
	"errors"
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

func (s *ProductService) Create(p *product.Product) error {
	return s.productRepo.Create(p)
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
