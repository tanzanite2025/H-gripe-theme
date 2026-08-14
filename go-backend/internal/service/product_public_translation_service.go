package service

import (
	"context"
	"strings"

	"commerce-platform/internal/domain/product"
)

// GetPublicBySlugWithRoutes resolves one public localized product and the
// active routes that belong to its translation group.
func (s *ProductService) GetPublicBySlugWithRoutes(slug, locale string) (*product.Product, []product.ProductTranslationRoute, error) {
	return s.GetPublicBySlugWithRoutesContext(context.Background(), slug, locale)
}

func (s *ProductService) GetPublicBySlugWithRoutesContext(ctx context.Context, slug, locale string) (*product.Product, []product.ProductTranslationRoute, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	lookupLocale := ""
	if strings.TrimSpace(locale) != "" {
		lookupLocale = normalizeLocale(locale)
	}

	result, err := s.loadProduct(ctx, productSlugCacheKey(slug, lookupLocale), func(ctx context.Context) (*product.Product, error) {
		return s.productRepo.FindBySlugContext(ctx, slug, lookupLocale)
	})
	if err != nil {
		return nil, nil, err
	}
	if result.Status != "active" {
		return nil, nil, ErrProductNotFound
	}
	_ = s.productRepo.IncrementViewCountContext(ctx, result.ID)

	routes, err := s.productRepo.FindPublicTranslationRoutesContext(ctx, result.ID)
	if err != nil {
		return nil, nil, err
	}

	result = sanitizeProductHTML(result)
	return result, routes, nil
}
