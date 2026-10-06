package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/faq"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/locales"
)

const (
	FAQRouteStatusCurrent = "current"
	FAQRouteStatusStale   = "stale"
	FAQRouteStatusMissing = "missing"
	FAQRouteStatusAlias   = "alias"
)

// FAQRouteSyncSummary is returned by startup, admin and URL-catalog syncs.
// Reconciliation is metadata-only for existing FAQ pages and never edits FAQ
// answers or editorial page text.
type FAQRouteSyncSummary struct {
	ManifestVersion string `json:"manifest_version"`
	Total           int    `json:"total"`
	Created         int    `json:"created"`
	Updated         int    `json:"updated"`
	Stale           int    `json:"stale"`
}

// ReconcileStorefrontRoutes makes faq_pages follow the current Nuxt route
// manifest. The manifest is the route identity source; page_id remains the
// stable content identity used by FAQ answers and automatic-reply references.
func (s *FAQService) ReconcileStorefrontRoutes(ctx context.Context, manifest seodomain.StorefrontRouteManifest) (FAQRouteSyncSummary, error) {
	var summary FAQRouteSyncSummary
	if s == nil || s.faqRepo == nil {
		return summary, fmt.Errorf("FAQ service is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return summary, fmt.Errorf("storefront route manifest version is missing")
	}
	select {
	case <-ctx.Done():
		return summary, ctx.Err()
	default:
	}

	desired := buildFAQRoutePages(manifest)
	stats, err := s.faqRepo.ReconcileRoutePages(ctx, desired, time.Now().UTC())
	if err != nil {
		return summary, fmt.Errorf("reconcile FAQ route pages: %w", err)
	}
	summary = FAQRouteSyncSummary{
		ManifestVersion: manifest.Version,
		Total:           stats.Total,
		Created:         stats.Created,
		Updated:         stats.Updated,
		Stale:           stats.Stale,
	}
	if stats.Created > 0 || stats.Updated > 0 || stats.Stale > 0 {
		s.notifyStorefrontContentChange("FAQ route reconciliation")
	}
	return summary, nil
}

func buildFAQRoutePages(manifest seodomain.StorefrontRouteManifest) []faq.FAQPage {
	routes := make([]seodomain.StorefrontRouteManifestRoute, 0, len(manifest.Routes)+1)
	routes = append(routes, manifest.Routes...)
	// Dynamic product pages are intentionally not emitted by the static Nuxt
	// manifest. They still have one stable FAQ page so every concrete slug can
	// resolve to the same content page.
	routes = append(routes, seodomain.StorefrontRouteManifestRoute{
		Key:          "products-product-detail",
		Path:         "/products/:slug",
		Label:        "Product Detail",
		Description:  "Common questions shown on individual product detail pages",
		IsSearchable: true,
		IsCheckable:  true,
		IsIndexable:  false,
	})

	localesList := locales.EnabledLocaleCodes()
	seenAt := time.Now().UTC()
	pages := make([]faq.FAQPage, 0, len(routes)*len(localesList))
	for index, declaration := range routes {
		path := normalizeFAQManifestPath(declaration.Path)
		key := strings.TrimSpace(declaration.Key)
		if path == "" || key == "" {
			continue
		}
		pageID := faqPageIDForRouteKey(key)
		routeStatus := FAQRouteStatusCurrent
		if declaration.IsAlias {
			routeStatus = FAQRouteStatusAlias
		}
		domain := faqRouteDomain(path)
		sortOrder := 100 + index*10
		for _, locale := range manifestRouteLocales(declaration) {
			pages = append(pages, faq.FAQPage{
				PageID:          pageID,
				RoutePath:       path,
				RouteKey:        key,
				ManifestVersion: manifest.Version,
				RouteStatus:     routeStatus,
				LastSeenAt:      &seenAt,
				Domain:          domain,
				Locale:          locale,
				Title:           faqRouteTitle(declaration.Label, declaration.IsAlias),
				Subtitle:        strings.TrimSpace(declaration.Description),
				SortOrder:       sortOrder,
				Status:          "active",
			})
		}
	}
	return pages
}

func faqPageIDForRouteKey(routeKey string) string {
	switch routeKey {
	case "resources-spoke-calculator":
		return "products-spoke-calculator"
	case "resources-membershipandpoints":
		return "company-membership"
	case "resources-blog":
		return "blog"
	case "resources-picture-warehouse":
		return "picture-warehouse"
	case "guides-tireguides-schwalbe-tire-selector":
		return "guides-schwalbe-tire-selector"
	case "guides-wheelset-buyers-spoke-lacing-topology":
		// Keep the FAQ content identity stable while the public route moves
		// from Resources into the Wheelset Guide route family.
		return "resources-wheelset-lacing-topology"
	case "legacy-faq":
		return "faq"
	default:
		return routeKey
	}
}

func faqRouteTitle(label string, alias bool) string {
	label = strings.TrimSpace(label)
	if label == "" {
		label = "Storefront"
	}
	if alias {
		return label + " FAQs (Alias)"
	}
	return label + " FAQs"
}

func faqRouteDomain(path string) string {
	path = normalizeFAQManifestPath(path)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "general"
	}
	if parts[0] == "resources" && len(parts) > 1 {
		return "resources"
	}
	return parts[0]
}

func normalizeFAQManifestPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		return "/"
	}
	return path
}
