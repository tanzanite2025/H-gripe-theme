package content

import (
	"time"

	postdomain "tanzanite/internal/domain/post"
	seodomain "tanzanite/internal/domain/seo"
)

type PublicPost struct {
	ID                 uint                         `json:"id"`
	Title              string                       `json:"title"`
	Slug               string                       `json:"slug"`
	Content            string                       `json:"content"`
	Excerpt            string                       `json:"excerpt"`
	Status             string                       `json:"status"`
	AuthorID           uint                         `json:"author_id"`
	Locale             string                       `json:"locale"`
	TranslationGroupID *uint                        `json:"translation_group_id,omitempty"`
	FeaturedImg        string                       `json:"featured_image"`
	ViewCount          int                          `json:"view_count"`
	MetaTitle          string                       `json:"meta_title"`
	MetaDescription    string                       `json:"meta_description"`
	CanonicalURL       string                       `json:"canonical_url"`
	Tags               string                       `json:"tags"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
	PublishedAt        *time.Time                   `json:"published_at"`
	LocalizedRoutes    []PublicPostTranslationRoute `json:"localized_routes,omitempty"`
}

type PublicPostTranslationRoute struct {
	Locale string `json:"locale"`
	Slug   string `json:"slug"`
	Path   string `json:"path"`
}

func PublicPostFromDomain(item postdomain.Post) PublicPost {
	return PublicPostFromDomainWithRoutes(item, nil)
}

func PublicPostFromDomainWithRoutes(item postdomain.Post, translationRoutes []postdomain.TranslationRoute) PublicPost {
	return PublicPost{
		ID:                 item.ID,
		Title:              item.Title,
		Slug:               item.Slug,
		Content:            item.Content,
		Excerpt:            item.Excerpt,
		Status:             item.Status,
		AuthorID:           item.AuthorID,
		Locale:             item.Locale,
		TranslationGroupID: item.TranslationGroupID,
		FeaturedImg:        item.FeaturedImg,
		ViewCount:          item.ViewCount,
		MetaTitle:          item.MetaTitle,
		MetaDescription:    item.MetaDesc,
		CanonicalURL:       item.CanonicalURL,
		Tags:               item.Tags,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
		PublishedAt:        item.PublishedAt,
		LocalizedRoutes:    publicPostTranslationRoutesFromDomain(translationRoutes),
	}
}

func publicPostTranslationRoutesFromDomain(items []postdomain.TranslationRoute) []PublicPostTranslationRoute {
	if len(items) == 0 {
		return nil
	}

	routes := make([]PublicPostTranslationRoute, 0, len(items))
	for _, item := range items {
		route := seodomain.BuildArticleRoute(item.Locale, item.Slug, item.Tags)
		if route.Path == "" {
			continue
		}
		routes = append(routes, PublicPostTranslationRoute{
			Locale: item.Locale,
			Slug:   item.Slug,
			Path:   route.Path,
		})
	}
	if len(routes) == 0 {
		return nil
	}
	return routes
}

func PublicPostsFromDomain(items []postdomain.Post) []PublicPost {
	if len(items) == 0 {
		return []PublicPost{}
	}

	result := make([]PublicPost, 0, len(items))
	for _, item := range items {
		result = append(result, PublicPostFromDomain(item))
	}
	return result
}
