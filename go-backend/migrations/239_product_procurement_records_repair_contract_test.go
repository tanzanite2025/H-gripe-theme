package migrations_test

import (
	"strings"
	"testing"
)

func TestProductProcurementRepairMigrationIsIdempotent(t *testing.T) {
	upSQL := readMigrationFile(t, "239_repair_product_procurement_records.up.sql")
	downSQL := readMigrationFile(t, "239_repair_product_procurement_records.down.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS product_procurement_records",
		"ADD COLUMN IF NOT EXISTS inbound_shipping_unit_cost",
		"ADD COLUMN IF NOT EXISTS packaging_unit_cost",
		"ADD COLUMN IF NOT EXISTS other_unit_cost",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_product_procurement_records_product_code",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("procurement repair migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "SELECT 1") {
		t.Fatal("procurement repair down migration must be a non-destructive no-op")
	}
	if strings.Contains(strings.ToLower(downSQL), "drop table") {
		t.Fatal("procurement repair down migration must not drop the table")
	}
}
