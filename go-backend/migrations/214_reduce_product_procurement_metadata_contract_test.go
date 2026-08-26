package migrations_test

import (
	"strings"
	"testing"
)

func TestReducedProductProcurementMetadataKeepsOnlyRequiredSourcingFields(t *testing.T) {
	upSQL := readMigrationFile(t, "214_reduce_product_procurement_metadata.up.sql")
	downSQL := readMigrationFile(t, "214_reduce_product_procurement_metadata.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE product_procurement_records",
		"DROP COLUMN IF EXISTS supplier_address",
		"DROP COLUMN IF EXISTS supplier_product_code",
		"DROP COLUMN IF EXISTS notes",
		"DROP COLUMN IF EXISTS is_enabled",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("procurement metadata migration is missing contract fragment %q", fragment)
		}
	}
	for _, retained := range []string{
		"purchase_price",
		"supplier_name",
		"supplier_contact_name",
		"lead_time_days",
		"minimum_order_quantity",
	} {
		if strings.Contains(strings.ToLower(upSQL), "drop column if exists "+retained) {
			t.Fatalf("procurement metadata migration must retain %q", retained)
		}
	}
	if !strings.Contains(downSQL, "ADD COLUMN IF NOT EXISTS is_enabled") {
		t.Fatal("down migration must restore the removed status column")
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"REFERENCES",
		"FOREIGN KEY",
	} {
		if strings.Contains(strings.ToLower(upSQL), strings.ToLower(forbidden)) {
			t.Fatalf("procurement metadata migration must not reference catalog fragment %q", forbidden)
		}
	}
}
