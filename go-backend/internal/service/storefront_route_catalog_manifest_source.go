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

		for _, locale := range locales.EnabledLocaleCodes() {
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
