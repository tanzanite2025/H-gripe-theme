package service

import (
	"fmt"
	"tanzanite/internal/domain/product"
)

func productIDCacheKey(id uint) string {
	return fmt.Sprintf("product:%d", id)
}

func productSlugCacheKey(slug, locale string) string {
	return fmt.Sprintf("product:slug:%s:%s", slug, locale)
}

func (s *ProductService) clearProductCache(p *product.Product) {
	if s.cache == nil || p == nil {
		return
	}

	_ = s.cache.Delete(productIDCacheKey(p.ID))
	if p.Slug != "" {
		_ = s.cache.Delete(productSlugCacheKey(p.Slug, p.Locale))
	}
}
