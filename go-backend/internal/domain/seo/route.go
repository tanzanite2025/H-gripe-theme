package seo

import (
	"net/url"
	"strings"

	"tanzanite/internal/pkg/locales"
)

// PageRoute is the read-only storefront route projection exposed to admin
// resource lists. SEO does not own the resource; it owns the route display
// contract used by the SEO control surface.
type PageRoute struct {
	Path string `json:"path"`
}

func BuildProductRoute(locale, slug string) PageRoute {
	return PageRoute{Path: buildLocalizedPath(locale, "/shop", slug)}
}

func BuildArticleRoute(locale, slug, tags string) PageRoute {
	hasNews := false
	hasWheelsbuild := false
	for _, rawTag := range strings.Split(tags, ",") {
		normalizedTag := strings.ToLower(strings.TrimSpace(rawTag))
		switch normalizedTag {
		case "news":
			hasNews = true
		case "wheelsbuild":
			hasWheelsbuild = true
		}
	}

	category := ""
	switch {
	case hasNews:
		category = "news"
	case hasWheelsbuild:
		category = "wheelsbuild"
	}

	base := "/blog"
	if category != "" {
		base += "/" + category
	}
	return PageRoute{Path: buildLocalizedPath(locale, base, slug)}
}

func buildLocalizedPath(locale, basePath, slug string) string {
	cleanSlug := strings.TrimSpace(slug)
	if cleanSlug == "" {
		return ""
	}

	prefix := ""
	if normalizedLocale := locales.Normalize(locale); normalizedLocale != "en" {
		prefix = "/" + normalizedLocale
	}

	return prefix + basePath + "/" + url.PathEscape(cleanSlug)
}
