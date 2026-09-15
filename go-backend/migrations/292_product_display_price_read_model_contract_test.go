package migrations_test

import (
	"strings"
	"testing"
)

func TestProductDisplayPriceReadModelMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "292_product_display_price_read_model.up.sql"))
	down := strings.ToUpper(readMigrationFile(t, "292_product_display_price_read_model.down.sql"))
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS PRODUCT_DISPLAY_PRICE_SNAPSHOTS",
		"SCOPE_KEY VARCHAR(64)",
		"PRODUCT_ID BIGINT",
		"VARIANT_ID BIGINT",
		"SOURCE_PRICE_MINOR BIGINT",
		"SOURCE_SALE_PRICE_MINOR BIGINT",
		"DISPLAY_PRICES JSONB",
		"UNIQUE INDEX IF NOT EXISTS IDX_PRODUCT_DISPLAY_PRICE_SNAPSHOTS_PRODUCT_SCOPE",
		"UNIQUE INDEX IF NOT EXISTS IDX_PRODUCT_DISPLAY_PRICE_SNAPSHOTS_VARIANT_SCOPE",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	if !strings.Contains(down, "DROP TABLE IF EXISTS PRODUCT_DISPLAY_PRICE_SNAPSHOTS") {
		t.Fatal("down migration must drop the display price read model table")
	}
}
