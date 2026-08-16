package service

import (
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ProductService struct {
	productRepo                    *repository.ProductRepository
	productBrandRepo               *repository.ProductBrandRepository
	productCategoryRepo            *repository.ProductCategoryRepository
	informationTemplateRepo        *repository.ProductInformationTemplateRepository
	customsClassificationRepo      *repository.CustomsClassificationRepository
	mediaService                   mediaAssetDeleter
	currencyPolicy                 *CurrencyPolicyService
	cache                          *cache.RedisCache
	cacheTTL                       time.Duration
	cacheLockTTL                   time.Duration
	cacheGroup                     singleflight.Group
	productCacheInvalidator        *ProductDetailCacheInvalidator
	productCacheEvents             ProductCacheEventPublisher
	storefrontHTMLCacheInvalidator *StorefrontHTMLCacheInvalidator
	merchantEvents                 MerchantProductEventPublisher
	txManager                      *repository.TxManager
}

func NewProductService(productRepo *repository.ProductRepository, cache *cache.RedisCache, cacheTTL int) *ProductService {
	return NewProductServiceWithCacheOptions(productRepo, cache, cacheTTL, 0)
}

func NewProductServiceWithCacheOptions(productRepo *repository.ProductRepository, cache *cache.RedisCache, cacheTTL, cacheLockTTL int) *ProductService {
	lockTTL := time.Duration(cacheLockTTL) * time.Second
	if lockTTL <= 0 {
		lockTTL = 5 * time.Second
	}
	var productCacheInvalidator *ProductDetailCacheInvalidator
	if cache != nil {
		productCacheInvalidator = NewProductDetailCacheInvalidator(productRepo, cache)
	}
	return &ProductService{
		productRepo:             productRepo,
		cache:                   cache,
		cacheTTL:                time.Duration(cacheTTL) * time.Second,
		cacheLockTTL:            lockTTL,
		productCacheInvalidator: productCacheInvalidator,
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

func (s *ProductService) ConfigureProductBrandRepository(repo *repository.ProductBrandRepository) {
	if s == nil {
		return
	}
	s.productBrandRepo = repo
}

func (s *ProductService) ConfigureCustomsClassificationRepository(repo *repository.CustomsClassificationRepository) {
	if s == nil {
		return
	}
	s.customsClassificationRepo = repo
}

func (s *ProductService) ConfigureProductCategoryRepository(repo *repository.ProductCategoryRepository) {
	if s == nil {
		return
	}
	s.productCategoryRepo = repo
}

func (s *ProductService) validateProductBrand(id *uint, allowDisabled bool) error {
	if id == nil || *id == 0 {
		return nil
	}
	if s.productBrandRepo == nil {
		return fmt.Errorf("%w: brand repository is not configured", ErrProductBrandInvalid)
	}
	brand, err := s.productBrandRepo.FindByID(*id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductBrandNotFound
	}
	if err != nil {
		return err
	}
	if !brand.IsEnabled && !allowDisabled {
		return fmt.Errorf("%w: brand is disabled", ErrProductBrandInvalid)
	}
	return nil
}

func (s *ProductService) validateProductCategory(id *uint, allowDisabled bool) error {
	if id == nil || *id == 0 {
		return nil
	}
	if s.productCategoryRepo == nil {
		return fmt.Errorf("%w: category repository is not configured", ErrProductCategoryInvalid)
	}
	category, err := s.productCategoryRepo.FindByID(*id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductCategoryNotFound
	}
	if err != nil {
		return err
	}
	if !category.IsEnabled && !allowDisabled {
		return fmt.Errorf("%w: category is disabled", ErrProductCategoryInvalid)
	}
	return nil
}

func (s *ProductService) ConfigureMediaService(mediaService mediaAssetDeleter) {
	if s == nil {
		return
	}
	s.mediaService = mediaService
}

func (s *ProductService) SetStorefrontHTMLCacheInvalidator(invalidator *StorefrontHTMLCacheInvalidator) {
	s.storefrontHTMLCacheInvalidator = invalidator
}

func (s *ProductService) ConfigureMerchantEventPublisher(publisher MerchantProductEventPublisher) {
	if s == nil {
		return
	}
	s.merchantEvents = publisher
}

func (s *ProductService) ConfigureProductCacheEventPublisher(publisher ProductCacheEventPublisher) {
	if s == nil {
		return
	}
	s.productCacheEvents = publisher
}

func (s *ProductService) ConfigureTxManager(manager *repository.TxManager) {
	if s == nil {
		return
	}
	s.txManager = manager
}

var (
	ErrProductNotFound                                = errors.New("product not found")
	ErrProductSKUExists                               = errors.New("product sku already exists")
	ErrProductSpecificationTemplateNotFound           = errors.New("product specification template not found")
	ErrProductSpecificationTemplateInvalid            = errors.New("product specification template invalid")
	ErrProductSpecificationTemplateSlugExists         = errors.New("product specification template slug already exists")
	ErrProductSpecificationTemplateSystemManaged      = errors.New("system product specification template structure is managed by the platform")
	ErrProductSpecificationTemplateTranslationInvalid = errors.New("product specification template translation invalid")
	ErrProductLocaleImmutable                         = errors.New("product locale cannot be changed after creation")
	ErrProductSpecInvalid                             = errors.New("product spec invalid")
	ErrProductVariantInvalid                          = errors.New("product variant invalid")
	ErrProductMediaInvalid                            = errors.New("product media invalid")
	ErrProductCustomsInfoInvalid                      = errors.New("product customs information invalid")
	ErrProductCustomsProfileNotFound                  = errors.New("product customs classification profile not found")
	ErrProductCustomsProfileInvalid                   = errors.New("product customs classification profile invalid")
	ErrProductTranslationInvalid                      = errors.New("product translation relationship invalid")
)

type ProductSearchInput struct {
	Locale       string
	Status       string
	Keyword      string
	ProductSpecificationTemplateSlug     string
	CategorySlug string
	BrandSlug    string
	PriceMin     *float64
	PriceMax     *float64
	SpecFilters  map[string][]string
	Page         int
	PageSize     int
}

type ProductRecommendationCandidateInput struct {
	ProductSpecificationTemplateID *uint
	Keyword                        string
	ExcludeProductIDs              []uint
	Page                           int
	PageSize                       int
}

func (s *ProductService) GetByID(id uint) (*product.Product, error) {
	return s.GetByIDContext(context.Background(), id)
}

func (s *ProductService) GetByIDContext(ctx context.Context, id uint) (*product.Product, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.loadProduct(ctx, productIDCacheKey(id), func(ctx context.Context) (*product.Product, error) {
		result, err := s.productRepo.FindByIDContext(ctx, id)
		if err != nil {
			return nil, err
		}
		_ = s.productRepo.IncrementViewCountContext(ctx, id)
		return result, nil
	})
}

func (s *ProductService) GetBySlug(slug, locale string) (*product.Product, error) {
	return s.GetBySlugContext(context.Background(), slug, locale)
}

func (s *ProductService) GetBySlugContext(ctx context.Context, slug, locale string) (*product.Product, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.loadProduct(ctx, productSlugCacheKey(slug, locale), func(ctx context.Context) (*product.Product, error) {
		result, err := s.productRepo.FindBySlugContext(ctx, slug, locale)
		if err != nil {
			return nil, err
		}
		_ = s.productRepo.IncrementViewCountContext(ctx, result.ID)
		return result, nil
	})
}

func (s *ProductService) loadProduct(ctx context.Context, cacheKey string, loader func(context.Context) (*product.Product, error)) (*product.Product, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.cache == nil || s.cacheTTL <= 0 {
		resultChan := s.cacheGroup.DoChan(cacheKey, func() (interface{}, error) {
			result, err := loader(ctx)
			if err != nil {
				return nil, err
			}
			return sanitizeProductHTML(result), nil
		})
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultChan:
			if result.Err != nil {
				return nil, result.Err
			}
			return result.Val.(*product.Product), nil
		}
	}

	if result, ok := s.getCachedProduct(ctx, cacheKey); ok {
		return result, nil
	}

	leaderContext := context.WithoutCancel(ctx)
	resultChan := s.cacheGroup.DoChan(cacheKey, func() (interface{}, error) {
		return s.loadProductWithDistributedLock(leaderContext, cacheKey, loader)
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*product.Product), nil
	}
}

func (s *ProductService) loadProductWithDistributedLock(ctx context.Context, cacheKey string, loader func(context.Context) (*product.Product, error)) (*product.Product, error) {
	if result, ok := s.getCachedProduct(ctx, cacheKey); ok {
		return result, nil
	}

	lockKey := productCacheLockKey(cacheKey)
	deadline := time.Now().Add(s.cacheLockTTL)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		lock, acquired, err := s.cache.AcquireLock(ctx, lockKey, s.cacheLockTTL)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return s.loadAndCacheProduct(ctx, cacheKey, loader)
		}
		if acquired {
			defer func() {
				_ = lock.Release(context.Background())
			}()
			if result, ok := s.getCachedProduct(ctx, cacheKey); ok {
				return result, nil
			}
			return s.loadAndCacheProductWithLease(ctx, cacheKey, loader, lock)
		}

		if result, ok := s.waitForCachedProduct(ctx, cacheKey, deadline); ok {
			return result, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if time.Now().After(deadline) {
			deadline = time.Now().Add(s.cacheLockTTL)
		}
	}
}

func (s *ProductService) loadAndCacheProduct(ctx context.Context, cacheKey string, loader func(context.Context) (*product.Product, error)) (*product.Product, error) {
	result, err := loader(ctx)
	if err != nil {
		return nil, err
	}
	result = sanitizeProductHTML(result)
	_ = s.cache.SetContext(ctx, cacheKey, result, s.cacheTTL)
	return result, nil
}

func (s *ProductService) loadAndCacheProductWithLease(ctx context.Context, cacheKey string, loader func(context.Context) (*product.Product, error), lock *cache.RedisLock) (*product.Product, error) {
	loadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	refreshErr := make(chan error, 1)
	refreshDone := make(chan struct{})
	go func() {
		defer close(refreshDone)
		interval := s.cacheLockTTL / 3
		if interval < 50*time.Millisecond {
			interval = 50 * time.Millisecond
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-loadCtx.Done():
				return
			case <-ticker.C:
				if err := lock.Refresh(loadCtx, s.cacheLockTTL); err != nil {
					select {
					case refreshErr <- err:
					default:
					}
					cancel()
					return
				}
			}
		}
	}()

	result, err := loader(loadCtx)
	if err == nil {
		result = sanitizeProductHTML(result)
		if cacheErr := s.cache.SetContext(loadCtx, cacheKey, result, s.cacheTTL); cacheErr != nil {
			// A cache write failure should not turn a successful database read into
			// an application failure; the next request will retry under the lock.
			_ = cacheErr
		}
	}
	cancel()
	<-refreshDone
	if err != nil {
		return nil, err
	}
	select {
	case err := <-refreshErr:
		return nil, err
	default:
		return result, nil
	}
}

func (s *ProductService) getCachedProduct(ctx context.Context, cacheKey string) (*product.Product, bool) {
	var cachedProduct product.Product
	if s.cache != nil && s.cache.GetContext(ctx, cacheKey, &cachedProduct) == nil {
		return sanitizeProductHTML(&cachedProduct), true
	}
	return nil, false
}

func (s *ProductService) waitForCachedProduct(ctx context.Context, cacheKey string, deadline time.Time) (*product.Product, bool) {
	timer := time.NewTimer(25 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-timer.C:
			if result, ok := s.getCachedProduct(ctx, cacheKey); ok {
				return result, true
			}
			if time.Now().After(deadline) {
				return nil, false
			}
			timer.Reset(25 * time.Millisecond)
		}
	}
}

func (s *ProductService) GetPublicByID(id uint) (*product.Product, error) {
	return s.GetPublicByIDContext(context.Background(), id)
}

func (s *ProductService) GetPublicByIDContext(ctx context.Context, id uint) (*product.Product, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	result, err := s.loadProduct(ctx, productIDCacheKey(id), func(ctx context.Context) (*product.Product, error) {
		return s.productRepo.FindByIDContext(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	if result.Status != "active" {
		return nil, ErrProductNotFound
	}
	_ = s.productRepo.IncrementViewCountContext(ctx, id)
	return sanitizeProductHTML(result), nil
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
	result, _, err := s.GetPublicBySlugWithRoutes(slug, locale)
	return result, err
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
	return sanitizeProductSliceHTML(products), total, err
}

func (s *ProductService) ListPublic(locale string, featured bool, page, pageSize int) ([]product.Product, int64, error) {
	return s.List(locale, "active", featured, page, pageSize)
}

func (s *ProductService) ListPublicAvailable(locale string, page, pageSize int) ([]product.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	products, total, err := s.productRepo.ListPublicAvailable(locale, offset, pageSize)
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
		ProductSpecificationTemplateID: input.ProductSpecificationTemplateID,
		Keyword:                        input.Keyword,
		ExcludeProductIDs:              input.ExcludeProductIDs,
		Offset:                         offset,
		Limit:                          pageSize,
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
		Locale:       input.Locale,
		Status:       "active",
		Keyword:      input.Keyword,
		ProductSpecificationTemplateSlug:     input.ProductSpecificationTemplateSlug,
		CategorySlug: input.CategorySlug,
		BrandSlug:    input.BrandSlug,
		PriceMin:     input.PriceMin,
		PriceMax:     input.PriceMax,
		SpecFilters:  input.SpecFilters,
		Offset:       offset,
		Limit:        pageSize,
	}
	products, total, err := s.productRepo.SearchPublic(query)
	return sanitizeProductSliceHTML(products), total, err
}

func (s *ProductService) Create(p *product.Product) error {
	if s.txManager != nil {
		var created *product.Product
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.Create(p); err != nil {
				return err
			}
			createdProduct, err := tx.Product.FindByID(p.ID)
			if err != nil {
				return err
			}
			created = createdProduct
			_, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			return enqueueMerchantProductChangeWithPublisher(merchantEvents, createdProduct, "product_created")
		})
		if err != nil {
			return err
		}
		if created != nil {
			*p = *created
		}
		s.invalidateStorefrontHTMLCache("product create")
		return nil
	}
	if err := s.productRepo.Create(p); err != nil {
		return err
	}
	s.invalidateStorefrontHTMLCache("product create")
	if err := s.enqueueMerchantProductChange(p, "product_created"); err != nil {
		return requireMerchantEvent(err, p, "product_created")
	}
	return nil
}

func (s *ProductService) Update(p *product.Product) error {
	previousProduct, err := s.findProduct(p.ID)
	if err != nil {
		return err
	}

	if s.txManager != nil {
		var updated *product.Product
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.Product == nil {
				return errors.New("transactional product repository is not configured")
			}
			if err := tx.Product.Update(p); err != nil {
				return err
			}
			updatedProduct, err := tx.Product.FindByID(p.ID)
			if err != nil {
				return err
			}
			updated = updatedProduct
			cacheEvents, merchantEvents, err := s.transactionalProductPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			if err := enqueueProductCacheInvalidationByIDsWithPublisher(cacheEvents, []uint{previousProduct.ID, updatedProduct.ID}, "product update"); err != nil {
				return err
			}
			if merchantProductCoreChanged(previousProduct, updatedProduct) {
				if err := enqueueMerchantProductChangeWithPublisher(merchantEvents, updatedProduct, "product_source_changed"); err != nil {
					return requireMerchantEvent(err, updatedProduct, "product_source_changed")
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		s.clearProductCache(previousProduct)
		s.clearProductCache(updated)
		s.invalidateStorefrontHTMLCache("product update")
		if updated != nil {
			*p = *updated
		}
		return nil
	}

	if err := s.productRepo.Update(p); err != nil {
		return err
	}

	s.clearProductCache(previousProduct)
	s.clearProductCache(p)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{previousProduct.ID, p.ID}, "product update"); err != nil {
		return err
	}
	s.invalidateStorefrontHTMLCache("product update")

	if merchantProductCoreChanged(previousProduct, p) {
		reason := "product_source_changed"
		if err := s.enqueueMerchantProductChange(p, reason); err != nil {
			return requireMerchantEvent(err, p, reason)
		}
	}
	return nil
}

func (s *ProductService) transactionalProductPublishers(outboxRepo *repository.OutboxRepository) (ProductCacheEventPublisher, MerchantProductEventPublisher, error) {
	return newTransactionalProductPublishers(outboxRepo)
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
