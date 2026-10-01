package service

import (
	"fmt"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/locales"
)

func buildManifestRouteCatalogEntries(
	manifest seodomain.StorefrontRouteManifest,
	summary *StorefrontRouteCatalogSyncSummary,
) []seodomain.StorefrontRouteCatalogEntry {
	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0, len(manifest.Routes)*len(locales.EnabledLocaleCodes()))

	for _, declaration := range manifest.Routes {
		path := normalizeCatalogRoutePath(declaration.Path)
		if path == "" {
			continue
		}
		canonicalPath := normalizeCatalogRoutePath(declaration.CanonicalPath)
		if canonicalPath == "" {
			canonicalPath = path
		}

		for _, locale := range manifestRouteLocales(declaration) {
			routePath := seodomain.BuildStaticRoute(locale, path).Path
			routeCanonicalPath := seodomain.BuildStaticRoute(locale, canonicalPath).Path
			sourceType := seodomain.RouteSourceStatic
			entryStatus := seodomain.RouteEntryStatusActive
			if declaration.IsAlias {
				sourceType = seodomain.RouteSourceAlias
				entryStatus = seodomain.RouteEntryStatusAlias
				summary.AliasEntries++
			} else {
				summary.StaticEntries++
			}

			entries = append(entries, seodomain.StorefrontRouteCatalogEntry{
				RouteKey:        fmt.Sprintf("manifest:%s:%s", declaration.Key, locale),
				Path:            routePath,
				Locale:          locale,
				SourceType:      sourceType,
				SourceKey:       declaration.Key,
				Title:           declaration.Label,
				Summary:         declaration.Description,
				CanonicalPath:   routeCanonicalPath,
				IsAlias:         declaration.IsAlias,
				IsSearchable:    declaration.IsSearchable,
				IsCheckable:     declaration.IsCheckable,
				IsIndexable:     declaration.IsIndexable,
				EntryStatus:     entryStatus,
				ManifestVersion: manifest.Version,
			})
		}
	}

	return entries
}

// manifestRouteLocales lets a page opt into sitemap/catalog discovery only for
// locales that have a real page translation. Existing manifest entries omit
// the field and therefore retain the all-enabled-locales behavior.
func manifestRouteLocales(declaration seodomain.StorefrontRouteManifestRoute) []string {
	if len(declaration.SitemapLocales) == 0 {
		return locales.EnabledLocaleCodes()
	}

	enabled := make(map[string]struct{}, len(locales.EnabledLocaleCodes()))
	for _, code := range locales.EnabledLocaleCodes() {
		enabled[code] = struct{}{}
	}

	selected := make([]string, 0, len(declaration.SitemapLocales))
	seen := make(map[string]struct{}, len(declaration.SitemapLocales))
	for _, raw := range declaration.SitemapLocales {
		code := locales.Normalize(raw)
		if _, ok := enabled[code]; !ok {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		selected = append(selected, code)
	}
	return selected
}
