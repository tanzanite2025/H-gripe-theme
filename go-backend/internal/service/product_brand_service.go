package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

type ProductBrandInput struct {
	Name        string
	Slug        string
	Description string
	LogoURL     string
	WebsiteURL  string
	IsEnabled   bool
	SortOrder   int
}

var (
	ErrProductBrandNotFound   = errors.New("product brand not found")
	ErrProductBrandInvalid    = errors.New("product brand invalid")
	ErrProductBrandSlugExists = errors.New("product brand slug already exists")
	ErrProductBrandInUse      = errors.New("product brand is used by products")
)

type ProductBrandService struct {
	repo           *repository.ProductBrandRepository
	productRepo    *repository.ProductRepository
	cacheEvents    ProductCacheEventPublisher
	merchantEvents MerchantProductEventPublisher
	txManager      *repository.TxManager
	htmlCache      *StorefrontHTMLCacheInvalidator
}

func NewProductBrandService(repo *repository.ProductBrandRepository) *ProductBrandService {
	return &ProductBrandService{repo: repo}
}

func (s *ProductBrandService) ConfigureProductDependencies(
	productRepo *repository.ProductRepository,
	cacheEvents ProductCacheEventPublisher,
	merchantEvents MerchantProductEventPublisher,
) {
	if s == nil {
		return
	}
	s.productRepo = productRepo
	s.cacheEvents = cacheEvents
	s.merchantEvents = merchantEvents
}

func (s *ProductBrandService) ConfigureTxManager(manager *repository.TxManager) {
	if s == nil {
		return
	}
	s.txManager = manager
}

func (s *ProductBrandService) SetStorefrontHTMLCacheInvalidator(invalidator *StorefrontHTMLCacheInvalidator) {
	if s == nil {
		return
	}
	s.htmlCache = invalidator
}

func (s *ProductBrandService) List(includeDisabled bool) ([]product.ProductBrand, error) {
	return s.repo.List(includeDisabled)
}

func (s *ProductBrandService) Get(id uint) (*product.ProductBrand, error) {
	brand, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductBrandNotFound
	}
	return brand, err
}

func (s *ProductBrandService) Create(input ProductBrandInput) (*product.ProductBrand, error) {
	brand, err := normalizeProductBrandInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsBySlug(brand.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductBrandSlugExists
	}
	if s.txManager != nil {
		var created *product.ProductBrand
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.ProductBrand == nil {
				return errors.New("transactional product brand repository is not configured")
			}
			if err := tx.ProductBrand.Create(brand); err != nil {
				return err
			}
			created, err = tx.ProductBrand.FindByID(brand.ID)
			return err
		})
		if err != nil {
			return nil, err
		}
		s.purgeHTMLCache("product brand create")
		return created, nil
	}
	if err := s.repo.Create(brand); err != nil {
		return nil, err
	}
	s.purgeHTMLCache("product brand create")
	return s.Get(brand.ID)
}

func (s *ProductBrandService) Update(id uint, input ProductBrandInput) (*product.ProductBrand, error) {
	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeProductBrandInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsBySlug(normalized.Slug, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductBrandSlugExists
	}
	normalized.ID = existing.ID
	if s.txManager != nil {
		var updated *product.ProductBrand
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.ProductBrand == nil {
				return errors.New("transactional product brand repository is not configured")
			}
			if err := tx.ProductBrand.Update(normalized); err != nil {
				return err
			}
			updated, err = tx.ProductBrand.FindByID(id)
			if err != nil {
				return err
			}
			if !productBrandAffectsProductPayload(existing, updated) {
				return nil
			}
			cacheEvents, merchantEvents, err := s.transactionalBrandPublishers(tx.Outbox)
			if err != nil {
				return err
			}
			return s.notifyBrandDependentsWith(tx.Product, cacheEvents, merchantEvents, id, "product_brand_changed")
		})
		if err != nil {
			return nil, err
		}
		s.purgeHTMLCache("product brand update")
		return updated, nil
	}
	if err := s.repo.Update(normalized); err != nil {
		return nil, err
	}
	if productBrandAffectsProductPayload(existing, normalized) {
		if err := s.notifyBrandDependents(id, "product_brand_changed"); err != nil {
			return nil, fmt.Errorf("product brand %d updated but dependent events were not queued: %w", id, err)
		}
	}
	s.purgeHTMLCache("product brand update")
	return s.Get(id)
}

func (s *ProductBrandService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	count, err := s.repo.CountProducts(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: %d products still reference this brand", ErrProductBrandInUse, count)
	}
	if s.txManager != nil {
		err := s.txManager.WithinTx(func(tx repository.TxRepositories) error {
			if tx.ProductBrand == nil {
				return errors.New("transactional product brand repository is not configured")
			}
			count, err := tx.ProductBrand.CountProducts(id)
			if err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("%w: %d products still reference this brand", ErrProductBrandInUse, count)
			}
			return tx.ProductBrand.Delete(id)
		})
		if err != nil {
			return err
		}
		s.purgeHTMLCache("product brand delete")
		return nil
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.purgeHTMLCache("product brand delete")
	return nil
}

func (s *ProductBrandService) notifyBrandDependents(brandID uint, reason string) error {
	return s.notifyBrandDependentsWith(s.productRepo, s.cacheEvents, s.merchantEvents, brandID, reason)
}

func (s *ProductBrandService) notifyBrandDependentsWith(
	productRepo *repository.ProductRepository,
	cacheEvents ProductCacheEventPublisher,
	merchantEvents MerchantProductEventPublisher,
	brandID uint,
	reason string,
) error {
	if brandID == 0 {
		return nil
	}
	if cacheEvents != nil {
		if err := cacheEvents.EnqueueProductCacheInvalidateByBrandID(brandID, reason); err != nil {
			return fmt.Errorf("enqueue product cache invalidation: %w", err)
		}
	}
	if productRepo == nil || merchantEvents == nil {
		return nil
	}

	var eventErr error
	var afterID uint
	for {
		products, err := productRepo.FindProductSyncIdentitiesByBrandIDPage(brandID, afterID, productDependencyBatchSize)
		if err != nil {
			return fmt.Errorf("load products for Merchant refresh: %w", err)
		}
		if len(products) == 0 {
			break
		}
		for index := range products {
			item := &products[index]
			if item.Status == "active" {
				eventErr = errors.Join(eventErr, merchantEvents.EnqueueProductUpsert(item.ID, reason))
				continue
			}
			eventErr = errors.Join(eventErr, merchantEvents.EnqueueProductWithdraw(item.ID, reason))
		}
		afterID = products[len(products)-1].ID
		if len(products) < productDependencyBatchSize {
			break
		}
	}
	if eventErr != nil {
		return fmt.Errorf("enqueue Merchant product refresh: %w", eventErr)
	}
	return nil
}

func (s *ProductBrandService) transactionalBrandPublishers(outboxRepo *repository.OutboxRepository) (ProductCacheEventPublisher, MerchantProductEventPublisher, error) {
	return newTransactionalProductPublishers(outboxRepo)
}

func (s *ProductBrandService) purgeHTMLCache(reason string) {
	if s == nil || s.htmlCache == nil {
		return
	}
	s.htmlCache.PurgeAllAsync(reason)
}

func productBrandAffectsProductPayload(previous, next *product.ProductBrand) bool {
	if previous == nil || next == nil {
		return true
	}
	return previous.Name != next.Name ||
		previous.Slug != next.Slug ||
		previous.Description != next.Description ||
		previous.LogoURL != next.LogoURL ||
		previous.WebsiteURL != next.WebsiteURL ||
		previous.IsEnabled != next.IsEnabled
}

func normalizeProductBrandInput(input ProductBrandInput) (*product.ProductBrand, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if name == "" || slug == "" {
		return nil, fmt.Errorf("%w: name and slug are required", ErrProductBrandInvalid)
	}
	if strings.ContainsAny(slug, " /?#%") {
		return nil, fmt.Errorf("%w: slug contains unsupported characters", ErrProductBrandInvalid)
	}
	for _, value := range []struct {
		name string
		raw  string
	}{
		{name: "logo_url", raw: input.LogoURL},
		{name: "website_url", raw: input.WebsiteURL},
	} {
		if strings.TrimSpace(value.raw) == "" {
			continue
		}
		parsed, err := url.Parse(strings.TrimSpace(value.raw))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("%w: %s must be an absolute URL", ErrProductBrandInvalid, value.name)
		}
	}

	return &product.ProductBrand{
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
		LogoURL:     strings.TrimSpace(input.LogoURL),
		WebsiteURL:  strings.TrimSpace(input.WebsiteURL),
		IsEnabled:   input.IsEnabled,
		SortOrder:   input.SortOrder,
	}, nil
}
