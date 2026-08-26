package migrations_test

import (
	"strings"
	"testing"
)

func TestAddProductProcurementAdditionalCostsMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "216_add_product_procurement_additional_costs.up.sql")
	downSQL := readMigrationFile(t, "216_add_product_procurement_additional_costs.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE product_procurement_records",
		"ADD COLUMN IF NOT EXISTS inbound_shipping_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS customs_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS packaging_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS other_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("additional cost migration is missing contract fragment %q", fragment)
		}
	}

	for _, fragment := range []string{
		"DROP COLUMN IF EXISTS inbound_shipping_unit_cost",
		"DROP COLUMN IF EXISTS customs_unit_cost",
		"DROP COLUMN IF EXISTS packaging_unit_cost",
		"DROP COLUMN IF EXISTS other_unit_cost",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("additional cost down migration is missing contract fragment %q", fragment)
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
			t.Fatalf("additional cost migration must not reference catalog fragment %q", forbidden)
		}
	}
}
