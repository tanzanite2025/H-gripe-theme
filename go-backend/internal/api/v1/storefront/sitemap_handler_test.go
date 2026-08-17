package storefront

import (
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/seo"
)

func TestSitemapRouteForEntryUsesSitemapLastModFormat(t *testing.T) {
	updatedAt := time.Date(2026, time.August, 17, 10, 20, 30, 0, time.FixedZone("CST", 8*60*60))
	route := sitemapRouteForEntry(seo.StorefrontRouteCatalogEntry{
		Path:      "/shop/example",
		UpdatedAt: updatedAt,
	})

	if route.Loc != "/shop/example" {
		t.Fatalf("loc = %q, want /shop/example", route.Loc)
	}
	if strings.Contains(route.LastMod, ",") || strings.Contains(route.LastMod, "GMT") {
		t.Fatalf("lastmod = %q, must be ISO 8601 rather than HTTP date", route.LastMod)
	}
	if route.LastMod != "2026-08-17T02:20:30Z" {
		t.Fatalf("lastmod = %q, want 2026-08-17T02:20:30Z", route.LastMod)
	}
}
