package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/safehtml"
	"tanzanite/internal/repository"
	"time"

	"gorm.io/gorm"
)

type ProductService struct {
	productRepo                    *repository.ProductRepository
	informationTemplateRepo        *repository.ProductInformationTemplateRepository
	currencyPolicy                 *CurrencyPolicyService
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

func (s *ProductService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
	if s == nil {
		return
	}
	s.currencyPolicy = policy
}

func (s *ProductService) ConfigureInformationTemplateRepository(repo *repository.ProductInformationTemplateRepository) {
	if s == nil {
		return
	}
	s.informationTemplateRepo = repo
}

func (s *ProductService) SetStorefrontHTMLCacheInvalidator(invalidator *StorefrontHTMLCacheInvalidator) {
	s.storefrontHTMLCacheInvalidator = invalidator
}

var (
	ErrProductNotFound               = errors.New("product not found")
	ErrProductSKUExists              = errors.New("product sku already exists")
	ErrProductTypeNotFound           = errors.New("product type not found")
	ErrProductTypeInvalid            = errors.New("product type invalid")
	ErrProductTypeSlugExists         = errors.New("product type slug already exists")
	ErrProductTypeTranslationInvalid = errors.New("product type translation invalid")
	ErrProductLocaleImmutable        = errors.New("product locale cannot be changed after creation")
	ErrProductSpecInvalid            = errors.New("product spec invalid")
	ErrProductVariantInvalid         = errors.New("product variant invalid")
	ErrProductMediaInvalid           = errors.New("product media invalid")
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

type ProductRecommendationCandidateInput struct {
	ProductTypeID     *uint
	Keyword           string
	ExcludeProductIDs []uint
	Page              int
	PageSize          int
}

func (s *ProductService) GetByID(id uint) (*product.Product, error) {
	cacheKey := productIDCacheKey(id)

	var cachedProduct product.Product
	if s.cache != nil && s.cache.Get(cacheKey, &cachedProduct) == nil {
		return sanitizeProductHTML(&cachedProduct), nil
	}

	result, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	result = sanitizeProductHTML(result)

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
		return sanitizeProductHTML(&cachedProduct), nil
	}

	result, err := s.productRepo.FindBySlug(slug, locale)
	if err != nil {
		return nil, err
	}
	result = sanitizeProductHTML(result)

	_ = s.productRepo.IncrementViewCount(result.ID)

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *ProductService) GetPublicByID(id uint) (*product.Product, error) {
	result, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if result.Status != "active" {
		return nil, ErrProductNotFound
	}
	result = sanitizeProductHTML(result)
	_ = s.productRepo.IncrementViewCount(id)
	return result, nil
}

func (s *ProductService) GetRecommendationContextProduct(id uint) (*product.Product, error) {
	result, err := s.productRepo.FindByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if result.Status != "active" || result.TotalVariantStock() <= 0 {
		return nil, ErrProductNotFound
	}
	return sanitizeProductHTML(result), nil
}

func (s *ProductService) GetPublicBySlug(slug, locale string) (*product.Product, error) {
	result, err := s.productRepo.FindBySlug(slug, "")
	if err != nil {
		return nil, err
	}
	if result.Status != "active" {
		return nil, ErrProductNotFound
	}
	result = sanitizeProductHTML(result)
	_ = s.productRepo.IncrementViewCount(result.ID)
	return result, nil
}

func (s *ProductService) validateInformationTemplate(id *uint, expectedKind, locale string, allowDisabled bool) error {
	if id == nil || *id == 0 {
		return nil
	}
	if s.informationTemplateRepo == nil {
		return fmt.Errorf("%w: template repository is not configured", ErrProductInformationTemplateInvalid)
	}
	template, err := s.informationTemplateRepo.FindByID(*id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: template not found", ErrProductInformationTemplateInvalid)
	}
	if err != nil {
		return err
	}
	if template.Kind != expectedKind {
		return fmt.Errorf("%w: template kind is invalid", ErrProductInformationTemplateInvalid)
	}
	if !template.IsEnabled && !allowDisabled {
		return fmt.Errorf("%w: template status is invalid", ErrProductInformationTemplateInvalid)
	}
	if template.Locale != "" && template.Locale != "en" && template.Locale != locale {
		return fmt.Errorf("%w: template locale does not match product locale", ErrProductInformationTemplateInvalid)
	}
	return nil
}

func (s *ProductService) List(locale, status string, featured bool, page, pageSize int) ([]product.Product, int64, error) {
	offset := (page - 1) * pageSize
	products, total, err := s.productRepo.List(locale, status, featured, offset, pageSize)
	if err == nil && total == 0 && locale != "" && locale != "en" {
		products, total, err = s.productRepo.List("en", status, featured, offset, pageSize)
	}
	return sanitizeProductSliceHTML(products), total, err
}

func (s *ProductService) ListPublic(locale string, featured bool, page, pageSize int) ([]product.Product, int64, error) {
	return s.List("", "active", featured, page, pageSize)
}

func (s *ProductService) ListPublicAvailable(locale string, page, pageSize int) ([]product.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	products, total, err := s.productRepo.ListPublicAvailable("", offset, pageSize)
	return sanitizeProductSliceHTML(products), total, err
}

func (s *ProductService) ListRecommendationCandidates(input ProductRecommendationCandidateInput) ([]product.Product, int64, error) {
	page := input.Page
	pageSize := input.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	products, total, err := s.productRepo.ListRecommendationCandidates(repository.ProductRecommendationQuery{
		ProductTypeID:     input.ProductTypeID,
		Keyword:           input.Keyword,
		ExcludeProductIDs: input.ExcludeProductIDs,
		Offset:            offset,
		Limit:             pageSize,
	})
	return sanitizeProductSliceHTML(products), total, err
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
		Locale:      "",
		Status:      "active",
		Keyword:     input.Keyword,
		TypeSlug:    input.TypeSlug,
		PriceMin:    input.PriceMin,
		PriceMax:    input.PriceMax,
		SpecFilters: input.SpecFilters,
		Offset:      offset,
		Limit:       pageSize,
	}
	products, total, err := s.productRepo.SearchPublic(query)
	return sanitizeProductSliceHTML(products), total, err
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

func sanitizeProductHTML(p *product.Product) *product.Product {
	if p == nil {
		return nil
	}
	if sanitized, err := safehtml.Sanitize(p.Description); err == nil {
		p.Description = sanitized
	}
	if sanitized, err := safehtml.Sanitize(p.ShortDesc); err == nil {
		p.ShortDesc = sanitized
	}
	if p.AfterSalesTemplate != nil {
		if sanitized, err := safehtml.Sanitize(p.AfterSalesTemplate.Content); err == nil {
			p.AfterSalesTemplate.Content = sanitized
		}
	}
	if p.PackagingTemplate != nil {
		if sanitized, err := safehtml.Sanitize(p.PackagingTemplate.Content); err == nil {
			p.PackagingTemplate.Content = sanitized
		}
	}
	return p
}

func sanitizeProductSliceHTML(products []product.Product) []product.Product {
	for index := range products {
		if sanitized, err := safehtml.Sanitize(products[index].Description); err == nil {
			products[index].Description = sanitized
		}
		if sanitized, err := safehtml.Sanitize(products[index].ShortDesc); err == nil {
			products[index].ShortDesc = sanitized
		}
		if products[index].AfterSalesTemplate != nil {
			if sanitized, err := safehtml.Sanitize(products[index].AfterSalesTemplate.Content); err == nil {
				products[index].AfterSalesTemplate.Content = sanitized
			}
		}
		if products[index].PackagingTemplate != nil {
			if sanitized, err := safehtml.Sanitize(products[index].PackagingTemplate.Content); err == nil {
				products[index].PackagingTemplate.Content = sanitized
			}
		}
	}
	return products
}
