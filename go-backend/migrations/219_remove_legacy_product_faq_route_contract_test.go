package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveLegacyProductFAQRouteKeepsOnlyFormalProductLookup(t *testing.T) {
	upSQL := readMigrationFile(t, "219_remove_legacy_product_faq_route.up.sql")
	downSQL := readMigrationFile(t, "219_remove_legacy_product_faq_route.down.sql")

	for _, fragment := range []string{
		"page_id = 'products-product-detail'",
		"page_id = 'shop-product-detail'",
		"route_path = '/products/:slug'",
		"DELETE FROM faq_categories",
		"DELETE FROM faq_pages",
	} {
		if !strings.Contains(strings.ToLower(upSQL), strings.ToLower(fragment)) {
			t.Fatalf("FAQ route migration is missing contract fragment %q", fragment)
		}
	}

	if strings.Contains(strings.ToLower(upSQL), "legacy product detail faqs") {
		t.Fatal("FAQ route migration must remove the legacy product title from the active product page")
	}
	if !strings.Contains(strings.ToLower(downSQL), "page_id = 'shop-product-detail'") ||
		!strings.Contains(strings.ToLower(downSQL), "route_path, domain, locale") {
		t.Fatal("FAQ route down migration must restore the retired page record")
	}
}
