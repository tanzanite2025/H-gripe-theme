package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigration319RemovesOrderLegacyMajorAmounts(t *testing.T) {
	root := "319_remove_order_legacy_major_amounts.up.sql"
	data, err := os.ReadFile(filepath.Join(root))
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, column := range []string{"payment_amount", "subtotal_amount", "shipping_fee", "tax_amount", "discount_amount", "total_amount", "points_value"} {
		if !strings.Contains(sql, "drop column if exists "+column) {
			t.Fatalf("migration missing orders column %s", column)
		}
	}
	for _, column := range []string{"price", "subtotal", "tax_amount", "discount", "total"} {
		if !strings.Contains(sql, "drop column if exists "+column) {
			t.Fatalf("migration missing order_items column %s", column)
		}
	}
}
