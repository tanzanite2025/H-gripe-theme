package migrations_test

import (
	"strings"
	"testing"
)

func TestWarrantyCheckFAQMigrationUsesOrderLookup(t *testing.T) {
	upSQL := readMigrationFile(t, "213_update_warranty_check_faq_to_order_lookup.up.sql")
	downSQL := readMigrationFile(t, "213_update_warranty_check_faq_to_order_lookup.down.sql")

	for _, fragment := range []string{
		"UPDATE faqs",
		"support-warranty-check",
		"order number",
		"订单号",
		"shipped orders",
	} {
		if !strings.Contains(strings.ToLower(upSQL), strings.ToLower(fragment)) {
			t.Fatalf("warranty FAQ migration is missing contract fragment %q", fragment)
		}
	}

	if strings.Contains(strings.ToLower(downSQL), "product code") ||
		strings.Contains(downSQL, "产品编码") {
		t.Fatal("warranty FAQ down migration must not restore product-code lookup wording")
	}
}
