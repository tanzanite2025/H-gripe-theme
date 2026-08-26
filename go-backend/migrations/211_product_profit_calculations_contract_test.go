package migrations_test

import (
	"strings"
	"testing"
)

func TestProductProfitCalculationsMigrationIsCatalogIndependent(t *testing.T) {
	upSQL := readMigrationFile(t, "211_create_product_profit_calculations.up.sql")
	downSQL := readMigrationFile(t, "211_create_product_profit_calculations.down.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS product_profit_calculations",
		"product_code VARCHAR(160) NOT NULL",
		"gross_margin_bps INTEGER NOT NULL",
		"warnings JSONB NOT NULL DEFAULT '[]'::jsonb",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_product_profit_calculations_product_code",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("profitability migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS product_profit_calculations") {
		t.Fatal("profitability down migration must only drop its own table")
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"REFERENCES",
		"FOREIGN KEY",
	} {
		if strings.Contains(strings.ToLower(upSQL), strings.ToLower(forbidden)) {
			t.Fatalf("profitability migration must not reference catalog fragment %q", forbidden)
		}
	}
}
