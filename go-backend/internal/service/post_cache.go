package service

import (
	"fmt"
	"tanzanite/internal/domain/post"
)

func postIDCacheKey(id uint) string {
	return fmt.Sprintf("post:%d", id)
}

func postSlugCacheKey(slug, locale string) string {
	return fmt.Sprintf("post:slug:%s:%s", slug, locale)
}

func (s *PostService) clearPostCache(p *post.Post) {
	if s.cache == nil || p == nil {
		return
	}

	_ = s.cache.Delete(postIDCacheKey(p.ID))
	if p.Slug != "" {
		_ = s.cache.Delete(postSlugCacheKey(p.Slug, p.Locale))
	}
}

func (s *PostService) invalidateStorefrontHTMLCache(reason string) {
	if s.storefrontHTMLCacheInvalidator == nil {
		return
	}

	s.storefrontHTMLCacheInvalidator.PurgeAllAsync(reason)
}
