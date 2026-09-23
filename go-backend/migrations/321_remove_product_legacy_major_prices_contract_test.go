package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration321RemovesProductLegacyMajorPrices(t *testing.T) {
	data, err := os.ReadFile("321_remove_product_legacy_major_prices.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, table := range []string{"products", "product_variants"} {
		for _, column := range []string{"price", "sale_price"} {
			needle := "alter table " + table
			if !strings.Contains(sql, needle) {
				t.Fatalf("migration missing table %s", table)
			}
			if !strings.Contains(sql, "drop column if exists "+column) {
				t.Fatalf("migration missing legacy column %s", column)
			}
		}
	}
}
