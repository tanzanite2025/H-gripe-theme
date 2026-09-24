package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration320RemovesCouponLegacyMajorAmounts(t *testing.T) {
	data, err := os.ReadFile("320_remove_coupon_legacy_major_amounts.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, column := range []string{"value", "min_amount", "max_discount"} {
		if !strings.Contains(sql, "drop column if exists "+column) {
			t.Fatalf("migration missing coupons column %s", column)
		}
	}
	if !strings.Contains(sql, "drop column if exists discount") {
		t.Fatal("migration missing coupon_usage discount column")
	}
}
