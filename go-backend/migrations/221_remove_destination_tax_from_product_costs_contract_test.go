package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveDestinationTaxFromProductCostsMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "221_remove_destination_tax_from_product_costs.up.sql")
	downSQL := readMigrationFile(t, "221_remove_destination_tax_from_product_costs.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE product_procurement_records",
		"DROP COLUMN IF EXISTS customs_unit_cost",
		"ALTER TABLE product_profit_calculations",
		"gross-margin-v3-no-customs",
		"gross_margin_bps",
		"purchase_price",
		"inbound_shipping_unit_cost",
		"packaging_unit_cost",
		"other_unit_cost",
		"negative_gross_profit",
		"calculation_status",
		"warnings",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("customs removal migration is missing contract fragment %q", fragment)
		}
	}

	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS customs_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("customs removal down migration is missing contract fragment %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"REFERENCES",
		"FOREIGN KEY",
	} {
		if strings.Contains(strings.ToLower(upSQL), strings.ToLower(forbidden)) {
			t.Fatalf("customs removal migration must not reference catalog fragment %q", forbidden)
		}
	}
}
