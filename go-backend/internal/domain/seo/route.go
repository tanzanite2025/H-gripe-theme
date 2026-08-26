package seo

import (
	"net/url"
	"strings"

	"commerce-platform/internal/pkg/locales"
)

// PageRoute is the read-only storefront route projection exposed to admin
// resource lists. SEO does not own the resource; it owns the route display
// contract used by the SEO control surface.
type PageRoute struct {
	Path string `json:"path"`
}

func BuildProductRoute(locale, slug string) PageRoute {
	return PageRoute{Path: buildLocalizedPath(locale, "/products", slug)}
}

func BuildCategoryRoute(locale string, slugs ...string) PageRoute {
	segments := make([]string, 0, len(slugs))
	for _, rawSlug := range slugs {
		slug := strings.TrimSpace(rawSlug)
		if slug == "" {
			continue
		}
		segments = append(segments, url.PathEscape(slug))
	}
	if len(segments) == 0 {
		return PageRoute{}
	}
	return PageRoute{Path: buildLocalizedStaticPath(locale, "/shop/"+strings.Join(segments, "/"))}
}

func BuildStaticRoute(locale, path string) PageRoute {
	return PageRoute{Path: buildLocalizedStaticPath(locale, path)}
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

	base := "/resources/blog"
	if category != "" {
		base += "/" + category
	}
	return PageRoute{Path: buildLocalizedPath(locale, base, slug)}
}

// IsProductRoute verifies the complete product route contract, including the
// flat /products path and the locale prefix used by the storefront.
func IsProductRoute(locale, path, slug string) bool {
	expected := BuildProductRoute(locale, slug).Path
	return expected != "" && strings.TrimSpace(path) == expected
}

// IsProductRoutePath accepts only one product slug segment. It deliberately
// does not accept category-shaped paths such as /shop/product-slug.
func IsProductRoutePath(locale, path string) bool {
	normalized := strings.TrimSpace(path)
	prefix := localizedRoutePrefix(locale, "/products/")
	if normalized == "" || !strings.HasPrefix(normalized, prefix) {
		return false
	}

	slug := strings.TrimPrefix(normalized, prefix)
	return slug != "" && !strings.Contains(slug, "/") && !strings.ContainsAny(slug, "?#")
}

// IsCategoryRoutePath accepts a real category path below /shop. The /shop
// landing route itself is owned by the static route manifest.
func IsCategoryRoutePath(locale, path string) bool {
	normalized := strings.TrimSpace(path)
	prefix := localizedRoutePrefix(locale, "/shop/")
	if normalized == "" || !strings.HasPrefix(normalized, prefix) {
		return false
	}

	segments := strings.Split(strings.TrimPrefix(normalized, prefix), "/")
	if len(segments) == 0 {
		return false
	}
	for _, segment := range segments {
		if segment == "" || strings.ContainsAny(segment, "?#") {
			return false
		}
	}
	return true
}

func buildLocalizedPath(locale, basePath, slug string) string {
	cleanSlug := strings.TrimSpace(slug)
	if cleanSlug == "" {
		return ""
	}

	return buildLocalizedStaticPath(locale, basePath+"/"+url.PathEscape(cleanSlug))
}

func buildLocalizedStaticPath(locale, path string) string {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" || !strings.HasPrefix(cleanPath, "/") {
		return ""
	}
	cleanPath = "/" + strings.Trim(cleanPath, "/")
	if cleanPath == "//" {
		cleanPath = "/"
	}

	prefix := ""
	if normalizedLocale := locales.Normalize(locale); normalizedLocale != "en" {
		prefix = "/" + normalizedLocale
	}

	if cleanPath == "/" {
		return prefix + "/"
	}
	return prefix + cleanPath
}

func localizedRoutePrefix(locale, basePath string) string {
	localePath := BuildStaticRoute(locale, "/").Path
	if localePath == "/" {
		return basePath
	}
	return strings.TrimSuffix(localePath, "/") + basePath
}
