package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration326UsesMinorRefundReviewAmount(t *testing.T) {
	data, err := os.ReadFile("326_after_sales_refund_review_amount_minor.up.sql")
	if err != nil { t.Fatal(err) }
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{"proposed_amount_minor bigint", "drop column if exists proposed_amount"} {
		if !strings.Contains(sql, fragment) { t.Fatalf("migration missing %q", fragment) }
	}
}
