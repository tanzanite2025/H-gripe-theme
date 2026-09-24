package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCouponPercentageDecimalMigrationAddsExactRate(t *testing.T) {
	data, err := os.ReadFile("314_coupon_percentage_decimal.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"value_rate_decimal numeric(30,15)",
		"set value_rate_decimal",
		"chk_coupons_value_rate_decimal",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

