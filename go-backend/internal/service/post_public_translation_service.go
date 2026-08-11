package service

import (
	"commerce-platform/internal/domain/post"
)

// GetPublicBySlugWithRoutes resolves one published localized article and the
// published translation routes that belong to the same translation group.
func (s *PostService) GetPublicBySlugWithRoutes(slug, locale string) (*post.Post, []post.TranslationRoute, error) {
	lookupLocale := normalizeLocale(locale)

	result, err := s.postRepo.FindBySlug(slug, lookupLocale)
	if err != nil {
		return nil, nil, err
	}
	if result.Status != "published" {
		return nil, nil, ErrPostNotFound
	}

	routes, err := s.postRepo.FindPublishedTranslationRoutes(result.ID)
	if err != nil {
		return nil, nil, err
	}

	result = sanitizePostHTML(result)
	_ = s.postRepo.IncrementViewCount(result.ID)
	return result, routes, nil
}

func (s *PostService) GetPublicByIDWithRoutes(id uint) (*post.Post, []post.TranslationRoute, error) {
	result, err := s.GetPublicByID(id)
	if err != nil {
		return nil, nil, err
	}

	routes, err := s.postRepo.FindPublishedTranslationRoutes(result.ID)
	if err != nil {
		return nil, nil, err
	}
	return result, routes, nil
}
