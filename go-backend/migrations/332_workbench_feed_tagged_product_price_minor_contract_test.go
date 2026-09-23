package migrations_test

import (
	"strings"
	"testing"
)

func TestWorkbenchFeedTaggedProductPriceMinorMigrationRemovesLegacyMajorColumn(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "332_workbench_feed_tagged_product_price_minor.up.sql"))
	for _, fragment := range []string{
		"add column if not exists price_minor bigint",
		"drop column if exists price",
		"chk_workbench_feed_tagged_products_price_minor_non_negative",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("workbench feed tagged-product migration missing contract fragment %q", fragment)
		}
	}
}
