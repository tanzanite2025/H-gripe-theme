package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestRemoveTaxRateLegacyFloatMigrationUsesDecimalOnly(t *testing.T) {
	data, err := os.ReadFile("333_remove_tax_rate_legacy_float.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"rate_decimal numeric(30,15)",
		"set rate_decimal",
		"drop column if exists rate",
		"chk_tax_rates_rate_decimal_non_negative",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
