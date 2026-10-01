package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestFAQRouteReconciliationMigrationContract(t *testing.T) {
	root := "363_reconcile_faq_route_paths.up.sql"
	up, err := os.ReadFile(root)
	if err != nil {
		t.Fatalf("read FAQ route reconciliation migration: %v", err)
	}
	source := string(up)
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS route_key",
		"ADD COLUMN IF NOT EXISTS manifest_version",
		"ADD COLUMN IF NOT EXISTS route_status",
		"ADD COLUMN IF NOT EXISTS last_seen_at",
		"/resources/spoke-calculator",
		"/resources/membershipandpoints",
		"/resources/blog",
		"/resources/picture-warehouse",
		"/policies/refund-cancellation",
		"products-product-detail",
		"route_status = 'stale'",
		"FAQ answers and page IDs are deliberately preserved",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("FAQ route reconciliation migration is missing %q", fragment)
		}
	}

	down, err := os.ReadFile("363_reconcile_faq_route_paths.down.sql")
	if err != nil {
		t.Fatalf("read FAQ route reconciliation down migration: %v", err)
	}
	downSource := string(down)
	if !strings.Contains(downSource, "intentionally keeps repaired paths") {
		t.Fatal("down migration must document the non-destructive data policy")
	}
	if strings.Contains(strings.ToLower(downSource), "delete from faqs") {
		t.Fatal("down migration must not delete FAQ answers")
	}
}
