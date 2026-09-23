package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveProductDisplayPriceCoreColumnsMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "305_remove_product_display_price_core_columns.up.sql"))
	for _, fragment := range []string{
		"ALTER TABLE PRODUCTS",
		"ALTER TABLE PRODUCT_VARIANTS",
		"DROP COLUMN IF EXISTS DISPLAY_PRICES",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	down := strings.ToUpper(readMigrationFile(t, "305_remove_product_display_price_core_columns.down.sql"))
	if !strings.Contains(down, "ADD COLUMN IF NOT EXISTS DISPLAY_PRICES JSONB") {
		t.Fatal("down migration must restore display_prices columns")
	}
}
