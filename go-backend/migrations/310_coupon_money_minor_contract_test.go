package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCouponMoneyMinorMigrationContract(t *testing.T) {
	up, err := os.ReadFile("310_coupon_money_minor.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile("310_coupon_money_minor.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	upSQL := strings.ToLower(string(up))
	downSQL := strings.ToLower(string(down))
	for _, fragment := range []string{
		"alter table coupons",
		"value_minor",
		"min_amount_minor",
		"max_discount_minor",
		"alter table coupon_usage",
		"discount_minor",
		"chk_coupons_money_minor_non_negative",
		"chk_coupon_usage_discount_minor_non_negative",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"drop column if exists value_minor",
		"drop column if exists discount_minor",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
