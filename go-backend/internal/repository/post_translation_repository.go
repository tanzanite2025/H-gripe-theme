package repository

import (
	"strings"

	"commerce-platform/internal/domain/post"
)

type postTranslationRouteRecord struct {
	ID                 uint
	TranslationGroupID *uint
	Locale             string
	Slug               string
	Tags               string
	Status             string
}

// FindPublishedTranslationRoutes returns only published routes in the
// current article's translation group. An ungrouped article returns itself.
func (r *PostRepository) FindPublishedTranslationRoutes(postID uint) ([]post.TranslationRoute, error) {
	var current postTranslationRouteRecord
	if err := r.db.Model(&post.Post{}).
		Select("id, translation_group_id, locale, slug, tags, status").
		Where("id = ?", postID).
		First(&current).Error; err != nil {
		return nil, err
	}

	query := r.db.Model(&post.Post{}).
		Select("id, locale, slug, tags, status").
		Where("status = ?", "published")

	if current.TranslationGroupID != nil && *current.TranslationGroupID > 0 {
		query = query.Where("translation_group_id = ?", *current.TranslationGroupID)
	} else {
		query = query.Where("id = ?", current.ID)
	}

	var records []postTranslationRouteRecord
	if err := query.Order("locale ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	routes := make([]post.TranslationRoute, 0, len(records))
	seenLocales := make(map[string]struct{}, len(records))
	for _, record := range records {
		locale := strings.TrimSpace(record.Locale)
		slug := strings.TrimSpace(record.Slug)
		if locale == "" || slug == "" {
			continue
		}
		if _, exists := seenLocales[locale]; exists {
			continue
		}
		seenLocales[locale] = struct{}{}
		routes = append(routes, post.TranslationRoute{
			Locale: locale,
			Slug:   slug,
			Tags:   record.Tags,
		})
	}

	return routes, nil
}
