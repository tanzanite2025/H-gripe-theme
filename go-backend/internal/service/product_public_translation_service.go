package service

import (
	"strings"

	"commerce-platform/internal/domain/product"
)

// GetPublicBySlugWithRoutes resolves one public localized product and the
// active routes that belong to its translation group.
func (s *ProductService) GetPublicBySlugWithRoutes(slug, locale string) (*product.Product, []product.ProductTranslationRoute, error) {
	lookupLocale := ""
	if strings.TrimSpace(locale) != "" {
		lookupLocale = normalizeLocale(locale)
	}

	result, err := s.productRepo.FindBySlug(slug, lookupLocale)
	if err != nil {
		return nil, nil, err
	}
	if result.Status != "active" {
		return nil, nil, ErrProductNotFound
	}

	routes, err := s.productRepo.FindPublicTranslationRoutes(result.ID)
	if err != nil {
		return nil, nil, err
	}

	result = sanitizeProductHTML(result)
	_ = s.productRepo.IncrementViewCount(result.ID)
	return result, routes, nil
}
