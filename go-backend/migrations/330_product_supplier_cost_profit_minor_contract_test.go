package migrations_test

import (
	"strings"
	"testing"
)

func TestProductSupplierCostProfitMinorMigrationRemovesLegacyMajorColumns(t *testing.T) {
	upSQL := readMigrationFile(t, "330_product_supplier_cost_profit_minor.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS purchase_price_minor BIGINT",
		"ADD COLUMN IF NOT EXISTS list_price_minor BIGINT",
		"ADD COLUMN IF NOT EXISTS gross_profit_minor BIGINT",
		"DROP COLUMN IF EXISTS purchase_price",
		"DROP COLUMN IF EXISTS list_price",
		"DROP COLUMN IF EXISTS gross_profit",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("supplier cost/profit migration is missing contract fragment %q", fragment)
		}
	}
	if strings.Contains(upSQL, "ALTER TABLE product_procurement_records\n    DROP COLUMN IF EXISTS purchase_price") {
		// The migration must add and backfill the canonical columns before drop.
		if !strings.Contains(upSQL, "purchase_price_minor") {
			t.Fatal("supplier cost migration drops legacy purchase_price without minor replacement")
		}
	}
}
