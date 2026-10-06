package service

import (
	"testing"

	seodomain "commerce-platform/internal/domain/seo"
)

func TestManifestRouteLocalesLimitsPublicDiscoveryLocales(t *testing.T) {
	declaration := seodomain.StorefrontRouteManifestRoute{
		SitemapLocales: []string{"zh_cn", "en", "fr", "zh-CN", "unsupported"},
	}

	locales := manifestRouteLocales(declaration)
	if len(locales) != 3 {
		t.Fatalf("expected three unique enabled locales, got %v", locales)
	}
	if locales[0] != "zh_cn" || locales[1] != "en" || locales[2] != "fr" {
		t.Fatalf("unexpected locale order or normalization: %v", locales)
	}
}

func TestManifestRouteLocalesDefaultsToAllEnabledLocales(t *testing.T) {
	locales := manifestRouteLocales(seodomain.StorefrontRouteManifestRoute{})
	if len(locales) < 2 {
		t.Fatalf("expected the enabled locale registry, got %v", locales)
	}
	if locales[0] != "en" || locales[1] != "zh_cn" {
		t.Fatalf("expected enabled locale registry order to be preserved, got %v", locales[:2])
	}
}

func TestBuildManifestRouteCatalogEntriesHonoursSitemapLocales(t *testing.T) {
	manifest := seodomain.StorefrontRouteManifest{
		Version: "test",
		Routes: []seodomain.StorefrontRouteManifestRoute{
			{
				Key:            "guides-wheelset-buyers-spoke-lacing-topology",
				Path:           "/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference",
				Label:          "Wheelset spoke lacing topology",
				Description:    "Interactive wheelset lacing topology reference",
				SitemapLocales: []string{"en", "zh_cn"},
				IsSearchable:   true,
				IsCheckable:    true,
				IsIndexable:    true,
			},
		},
	}

	summary := &StorefrontRouteCatalogSyncSummary{}
	entries := buildManifestRouteCatalogEntries(manifest, summary)
	if len(entries) != 2 {
		t.Fatalf("expected two public locale entries, got %d", len(entries))
	}
	if entries[0].Locale != "en" || entries[0].Path != "/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference" {
		t.Fatalf("unexpected English route entry: %+v", entries[0])
	}
	if entries[1].Locale != "zh_cn" || entries[1].Path != "/zh_cn/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference" {
		t.Fatalf("unexpected Chinese route entry: %+v", entries[1])
	}
}

func TestFAQRoutePagesHonourManifestSitemapLocales(t *testing.T) {
	manifest := seodomain.StorefrontRouteManifest{
		Version: "test",
		Routes: []seodomain.StorefrontRouteManifestRoute{
			{
				Key:            "guides-wheelset-buyers-spoke-lacing-topology",
				Path:           "/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference",
				Label:          "Wheelset spoke lacing topology",
				Description:    "Interactive wheelset lacing topology reference",
				SitemapLocales: []string{"en", "zh_cn"},
			},
		},
	}

	pages := buildFAQRoutePages(manifest)
	localesByRoute := make(map[string]struct{})
	for _, page := range pages {
		if page.RouteKey == "guides-wheelset-buyers-spoke-lacing-topology" {
			localesByRoute[page.Locale] = struct{}{}
			if page.PageID != "resources-wheelset-lacing-topology" {
				t.Fatalf("FAQ page identity changed during route move: %q", page.PageID)
			}
		}
	}
	if len(localesByRoute) != 2 {
		t.Fatalf("expected FAQ route metadata only for en and zh_cn, got %v", localesByRoute)
	}
	if _, ok := localesByRoute["en"]; !ok {
		t.Fatal("expected English FAQ route metadata")
	}
	if _, ok := localesByRoute["zh_cn"]; !ok {
		t.Fatal("expected Chinese FAQ route metadata")
	}
}
