package service

import (
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/metrics"
	"errors"
	"fmt"
)

const (
	productCacheInvalidationSourceDirect = "direct"
	productCacheInvalidationSourceOutbox = "outbox"
)

type ProductDetailCacheInvalidator struct {
	productRepo productDetailCacheProductRepository
	cache       productDetailCache
}

type ProductCacheInvalidationResult struct {
	Products int
	Keys     int
}

type ProductCacheInvalidationExecutor interface {
	InvalidateProductCacheByIDsWithSource(ids []uint, source string) (ProductCacheInvalidationResult, error)
	InvalidateProductCacheByProductTypeIDWithSource(productTypeID uint, source string) (ProductCacheInvalidationResult, error)
	InvalidateProductCacheByInformationTemplateIDWithSource(templateID uint, source string) (ProductCacheInvalidationResult, error)
}

type ProductCacheInvalidator interface {
	InvalidateProductCacheByIDs(ids []uint)
}

type ProductDependencyCacheInvalidator interface {
	InvalidateProductCacheByProductTypeID(productTypeID uint)
	InvalidateProductCacheByInformationTemplateID(templateID uint)
}

type productDetailCacheProductRepository interface {
	FindProductCacheIdentitiesByIDs(ids []uint) ([]product.Product, error)
	FindProductCacheIdentitiesByProductTypeID(productTypeID uint) ([]product.Product, error)
	FindProductCacheIdentitiesByInformationTemplateID(templateID uint) ([]product.Product, error)
}

type productDetailCache interface {
	Delete(key string) error
}

func productIDCacheKey(id uint) string {
	return fmt.Sprintf("product:%d", id)
}

func productSlugCacheKey(slug, locale string) string {
	return fmt.Sprintf("product:slug:%s:%s", slug, locale)
}

func productCacheLockKey(cacheKey string) string {
	return fmt.Sprintf("lock:%s", cacheKey)
}

func NewProductDetailCacheInvalidator(productRepo productDetailCacheProductRepository, cache productDetailCache) *ProductDetailCacheInvalidator {
	return &ProductDetailCacheInvalidator{
		productRepo: productRepo,
		cache:       cache,
	}
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCache(p *product.Product) {
	_, _ = i.InvalidateProductCacheWithSource(p, productCacheInvalidationSourceDirect)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByIDs(ids []uint) {
	_, _ = i.InvalidateProductCacheByIDsWithSource(ids, productCacheInvalidationSourceDirect)
}

func (s *ProductService) clearProductCache(p *product.Product) {
	if s == nil || s.productCacheInvalidator == nil {
		return
	}
	s.productCacheInvalidator.InvalidateProductCache(p)
}

func (s *ProductService) enqueueProductCacheInvalidationByIDs(productIDs []uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByIDs(productIDs, reason)
}

func (s *ProductService) enqueueProductCacheInvalidationByProductTypeID(productTypeID uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByProductTypeID(productTypeID, reason)
}

func (s *ProductService) enqueueProductCacheInvalidationByInformationTemplateID(templateID uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByInformationTemplateID(templateID, reason)
}

func (s *ProductService) InvalidateProductCacheByIDs(ids []uint) {
	_, _ = s.InvalidateProductCacheByIDsWithSource(ids, productCacheInvalidationSourceDirect)
}

func (s *ProductService) InvalidateProductCacheByIDsWithSource(ids []uint, source string) (ProductCacheInvalidationResult, error) {
	if s == nil || s.productCacheInvalidator == nil {
		return ProductCacheInvalidationResult{}, nil
	}
	return s.productCacheInvalidator.InvalidateProductCacheByIDsWithSource(ids, source)
}

func (s *ProductService) InvalidateProductCacheByProductTypeID(productTypeID uint) {
	_, _ = s.InvalidateProductCacheByProductTypeIDWithSource(productTypeID, productCacheInvalidationSourceDirect)
}

func (s *ProductService) InvalidateProductCacheByProductTypeIDWithSource(productTypeID uint, source string) (ProductCacheInvalidationResult, error) {
	if s == nil || s.productCacheInvalidator == nil {
		return ProductCacheInvalidationResult{}, nil
	}
	return s.productCacheInvalidator.InvalidateProductCacheByProductTypeIDWithSource(productTypeID, source)
}

func (s *ProductService) InvalidateProductCacheByInformationTemplateID(templateID uint) {
	_, _ = s.InvalidateProductCacheByInformationTemplateIDWithSource(templateID, productCacheInvalidationSourceDirect)
}

func (s *ProductService) InvalidateProductCacheByInformationTemplateIDWithSource(templateID uint, source string) (ProductCacheInvalidationResult, error) {
	if s == nil || s.productCacheInvalidator == nil {
		return ProductCacheInvalidationResult{}, nil
	}
	return s.productCacheInvalidator.InvalidateProductCacheByInformationTemplateIDWithSource(templateID, source)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByProductTypeID(productTypeID uint) {
	_, _ = i.InvalidateProductCacheByProductTypeIDWithSource(productTypeID, productCacheInvalidationSourceDirect)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByInformationTemplateID(templateID uint) {
	_, _ = i.InvalidateProductCacheByInformationTemplateIDWithSource(templateID, productCacheInvalidationSourceDirect)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCaches(products []product.Product) {
	_, _ = i.InvalidateProductCachesWithSource(products, productCacheInvalidationSourceDirect)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheWithSource(p *product.Product, source string) (ProductCacheInvalidationResult, error) {
	if p == nil {
		return ProductCacheInvalidationResult{}, nil
	}
	return i.InvalidateProductCachesWithSource([]product.Product{*p}, source)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByIDsWithSource(ids []uint, source string) (ProductCacheInvalidationResult, error) {
	if i == nil || i.cache == nil || i.productRepo == nil {
		return ProductCacheInvalidationResult{}, nil
	}

	products, err := i.productRepo.FindProductCacheIdentitiesByIDs(uniqueUintIDs(ids))
	if err != nil {
		recordProductCacheInvalidationMetric(source, 0, err)
		return ProductCacheInvalidationResult{}, err
	}
	return i.InvalidateProductCachesWithSource(products, source)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByProductTypeIDWithSource(productTypeID uint, source string) (ProductCacheInvalidationResult, error) {
	if i == nil || i.cache == nil || i.productRepo == nil || productTypeID == 0 {
		return ProductCacheInvalidationResult{}, nil
	}
	products, err := i.productRepo.FindProductCacheIdentitiesByProductTypeID(productTypeID)
	if err != nil {
		recordProductCacheInvalidationMetric(source, 0, err)
		return ProductCacheInvalidationResult{}, err
	}
	return i.InvalidateProductCachesWithSource(products, source)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByInformationTemplateIDWithSource(templateID uint, source string) (ProductCacheInvalidationResult, error) {
	if i == nil || i.cache == nil || i.productRepo == nil || templateID == 0 {
		return ProductCacheInvalidationResult{}, nil
	}
	products, err := i.productRepo.FindProductCacheIdentitiesByInformationTemplateID(templateID)
	if err != nil {
		recordProductCacheInvalidationMetric(source, 0, err)
		return ProductCacheInvalidationResult{}, err
	}
	return i.InvalidateProductCachesWithSource(products, source)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCachesWithSource(products []product.Product, source string) (ProductCacheInvalidationResult, error) {
	if i == nil || i.cache == nil {
		return ProductCacheInvalidationResult{}, nil
	}

	result := ProductCacheInvalidationResult{}
	var err error
	for index := range products {
		keys := productCacheKeys(&products[index])
		if len(keys) == 0 {
			continue
		}
		result.Products++
		for _, key := range keys {
			if deleteErr := i.cache.Delete(key); deleteErr != nil {
				err = errors.Join(err, deleteErr)
			}
			result.Keys++
		}
	}
	recordProductCacheInvalidationMetric(source, result.Keys, err)
	return result, err
}

func (s *ProductService) invalidateStorefrontHTMLCache(reason string) {
	if s.storefrontHTMLCacheInvalidator == nil {
		return
	}

	s.storefrontHTMLCacheInvalidator.PurgeAllAsync(reason)
}

func uniqueUintIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	result := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func productCacheKeys(p *product.Product) []string {
	if p == nil || p.ID == 0 {
		return nil
	}
	keys := []string{productIDCacheKey(p.ID)}
	if p.Slug == "" {
		return keys
	}
	keys = append(keys, productSlugCacheKey(p.Slug, ""))
	keys = append(keys, productSlugCacheKey(p.Slug, p.Locale))
	if normalizedLocale := normalizeLocale(p.Locale); normalizedLocale != "" && normalizedLocale != p.Locale {
		keys = append(keys, productSlugCacheKey(p.Slug, normalizedLocale))
	}
	return keys
}

func recordProductCacheInvalidationMetric(source string, keys int, err error) {
	if source == "" {
		source = productCacheInvalidationSourceDirect
	}
	result := "success"
	if err != nil {
		result = "error"
	}
	metrics.ProductCacheInvalidations.WithLabelValues(source, result).Inc()
	metrics.ProductCacheInvalidationKeys.WithLabelValues(source).Observe(float64(keys))
}
