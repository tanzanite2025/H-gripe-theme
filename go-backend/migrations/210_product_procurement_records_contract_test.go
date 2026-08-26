package migrations_test

import (
	"strings"
	"testing"
)

func TestProductProcurementMigrationIsCatalogIndependent(t *testing.T) {
	upSQL := readMigrationFile(t, "210_create_product_procurement_records.up.sql")
	downSQL := readMigrationFile(t, "210_create_product_procurement_records.down.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS product_procurement_records",
		"product_code VARCHAR(160) NOT NULL",
		"supplier_name VARCHAR(255) NOT NULL",
		"lead_time_days INTEGER NOT NULL",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_product_procurement_records_product_code",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("procurement migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS product_procurement_records") {
		t.Fatal("procurement down migration must only drop its own table")
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"REFERENCES",
		"FOREIGN KEY",
	} {
		if strings.Contains(strings.ToLower(upSQL), strings.ToLower(forbidden)) {
			t.Fatalf("procurement migration must not reference catalog fragment %q", forbidden)
		}
	}
}
