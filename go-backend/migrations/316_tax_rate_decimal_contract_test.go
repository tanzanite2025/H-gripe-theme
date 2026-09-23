package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestTaxRateDecimalMigrationAddsExactRate(t *testing.T) {
	data, err := os.ReadFile("316_tax_rate_decimal.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"rate_decimal numeric(30,15)",
		"set rate_decimal",
		"chk_tax_rates_rate_decimal",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

