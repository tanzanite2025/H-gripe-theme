package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingMoneyMinorMigrationAddsCanonicalFields(t *testing.T) {
	data, err := os.ReadFile("313_shipping_money_minor.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"free_threshold_minor bigint",
		"default_fee_minor bigint",
		"fee_minor bigint",
		"chk_shipping_rules_money_minor_non_negative",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

