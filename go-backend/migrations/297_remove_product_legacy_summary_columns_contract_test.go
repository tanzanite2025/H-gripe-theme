package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveProductLegacySummaryColumnsMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "297_remove_product_legacy_summary_columns.up.sql"))
	if !strings.Contains(up, "ALTER TABLE PRODUCTS") {
		t.Fatal("up migration must alter products")
	}
	for _, column := range []string{"SKU", "PRICE", "SALE_PRICE", "STOCK"} {
		if !strings.Contains(up, "DROP COLUMN IF EXISTS "+column) {
			t.Fatalf("up migration must drop products.%s", strings.ToLower(column))
		}
	}
	if strings.TrimSpace(readMigrationFile(t, "297_remove_product_legacy_summary_columns.down.sql")) == "" {
		t.Fatal("down migration must document the intentional non-restoration")
	}
}
