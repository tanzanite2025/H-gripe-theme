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
	InvalidateProductCacheByProductSpecificationTemplateIDWithSource(productSpecificationTemplateID uint, source string) (ProductCacheInvalidationResult, error)
	InvalidateProductCacheByBrandIDWithSource(brandID uint, source string) (ProductCacheInvalidationResult, error)
	InvalidateProductCacheByInformationTemplateIDWithSource(templateID uint, source string) (ProductCacheInvalidationResult, error)
}

type ProductCacheInvalidator interface {
	InvalidateProductCacheByIDs(ids []uint)
}

type ProductDependencyCacheInvalidator interface {
	InvalidateProductCacheByProductSpecificationTemplateID(productSpecificationTemplateID uint)
	InvalidateProductCacheByBrandID(brandID uint)
	InvalidateProductCacheByInformationTemplateID(templateID uint)
}

type productDetailCacheProductRepository interface {
	FindProductCacheIdentitiesByIDs(ids []uint) ([]product.Product, error)
	FindProductCacheIdentitiesByProductSpecificationTemplateID(productSpecificationTemplateID uint) ([]product.Product, error)
	FindProductCacheIdentitiesByBrandID(brandID uint) ([]product.Product, error)
	FindProductCacheIdentitiesByInformationTemplateID(templateID uint) ([]product.Product, error)
}

type pagedProductBrandCacheProductRepository interface {
	FindProductCacheIdentitiesByBrandIDPage(brandID, afterID uint, limit int) ([]product.Product, error)
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
	if s == nil {
		return nil
	}
	return enqueueProductCacheInvalidationByIDsWithPublisher(s.productCacheEvents, productIDs, reason)
}

func enqueueProductCacheInvalidationByIDsWithPublisher(publisher ProductCacheEventPublisher, productIDs []uint, reason string) error {
	if publisher == nil {
		return nil
	}
	return publisher.EnqueueProductCacheInvalidateByIDs(productIDs, reason)
}

func (s *ProductService) enqueueProductCacheInvalidationByProductSpecificationTemplateID(productSpecificationTemplateID uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByProductSpecificationTemplateID(productSpecificationTemplateID, reason)
}

func (s *ProductService) enqueueProductCacheInvalidationByInformationTemplateID(templateID uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByInformationTemplateID(templateID, reason)
}

func (s *ProductService) enqueueProductCacheInvalidationByBrandID(brandID uint, reason string) error {
	if s == nil || s.productCacheEvents == nil {
		return nil
	}
	return s.productCacheEvents.EnqueueProductCacheInvalidateByBrandID(brandID, reason)
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

func (s *ProductService) InvalidateProductCacheByProductSpecificationTemplateID(productSpecificationTemplateID uint) {
	_, _ = s.InvalidateProductCacheByProductSpecificationTemplateIDWithSource(productSpecificationTemplateID, productCacheInvalidationSourceDirect)
}

func (s *ProductService) InvalidateProductCacheByProductSpecificationTemplateIDWithSource(productSpecificationTemplateID uint, source string) (ProductCacheInvalidationResult, error) {
	if s == nil || s.productCacheInvalidator == nil {
		return ProductCacheInvalidationResult{}, nil
	}
	return s.productCacheInvalidator.InvalidateProductCacheByProductSpecificationTemplateIDWithSource(productSpecificationTemplateID, source)
}

func (s *ProductService) InvalidateProductCacheByBrandID(brandID uint) {
	_, _ = s.InvalidateProductCacheByBrandIDWithSource(brandID, productCacheInvalidationSourceDirect)
}

func (s *ProductService) InvalidateProductCacheByBrandIDWithSource(brandID uint, source string) (ProductCacheInvalidationResult, error) {
	if s == nil || s.productCacheInvalidator == nil {
		return ProductCacheInvalidationResult{}, nil
	}
	return s.productCacheInvalidator.InvalidateProductCacheByBrandIDWithSource(brandID, source)
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

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByProductSpecificationTemplateID(productSpecificationTemplateID uint) {
	_, _ = i.InvalidateProductCacheByProductSpecificationTemplateIDWithSource(productSpecificationTemplateID, productCacheInvalidationSourceDirect)
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

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByProductSpecificationTemplateIDWithSource(productSpecificationTemplateID uint, source string) (ProductCacheInvalidationResult, error) {
	if i == nil || i.cache == nil || i.productRepo == nil || productSpecificationTemplateID == 0 {
		return ProductCacheInvalidationResult{}, nil
	}
	products, err := i.productRepo.FindProductCacheIdentitiesByProductSpecificationTemplateID(productSpecificationTemplateID)
	if err != nil {
		recordProductCacheInvalidationMetric(source, 0, err)
		return ProductCacheInvalidationResult{}, err
	}
	return i.InvalidateProductCachesWithSource(products, source)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByBrandID(brandID uint) {
	_, _ = i.InvalidateProductCacheByBrandIDWithSource(brandID, productCacheInvalidationSourceDirect)
}

func (i *ProductDetailCacheInvalidator) InvalidateProductCacheByBrandIDWithSource(brandID uint, source string) (ProductCacheInvalidationResult, error) {
	if i == nil || i.cache == nil || i.productRepo == nil || brandID == 0 {
		return ProductCacheInvalidationResult{}, nil
	}
	if pagedRepo, ok := i.productRepo.(pagedProductBrandCacheProductRepository); ok {
		result := ProductCacheInvalidationResult{}
		var afterID uint
		for {
			products, err := pagedRepo.FindProductCacheIdentitiesByBrandIDPage(brandID, afterID, productDependencyBatchSize)
			if err != nil {
				recordProductCacheInvalidationMetric(source, result.Keys, err)
				return result, err
			}
			if len(products) == 0 {
				break
			}
			batchResult, err := i.InvalidateProductCachesWithSource(products, source)
			result.Products += batchResult.Products
			result.Keys += batchResult.Keys
			if err != nil {
				return result, err
			}
			afterID = products[len(products)-1].ID
			if len(products) < productDependencyBatchSize {
				break
			}
		}
		return result, nil
	}
	products, err := i.productRepo.FindProductCacheIdentitiesByBrandID(brandID)
	if err != nil {
		recordProductCacheInvalidationMetric(source, 0, err)
		return ProductCacheInvalidationResult{}, err
	}
	return i.InvalidateProductCachesWithSource(products, source)
}

const productDependencyBatchSize = 500

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
