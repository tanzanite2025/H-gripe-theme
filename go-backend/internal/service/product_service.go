package service

import (
	"errors"
	"fmt"
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
