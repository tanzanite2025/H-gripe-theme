package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestPaymentMethodMoneyDecimalMigrationAddsExactFields(t *testing.T) {
	data, err := os.ReadFile("317_payment_method_money_decimal.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"fee_rate_decimal numeric(30,15)",
		"fee_value_minor bigint",
		"chk_payment_methods_fee_value_minor_non_negative",
		"when lower(coalesce(fee_type, 'fixed')) = 'percentage'",
		"fee_value_minor = 0",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	if strings.Contains(sql, "coalesce(fee_value, 0) * 100") {
		t.Fatal("migration must not invent a currency for legacy fixed fees")
	}
}
