package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveMemberLevelLegacyDiscountRateMigrationUsesDecimalOnly(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "338_remove_member_level_legacy_discount_rate.up.sql"))
	for _, fragment := range []string{
		"add column if not exists discount_rate_decimal numeric(30,15)",
		"set discount_rate_decimal",
		"round(coalesce(discount_rate, 0)::numeric, 15)",
		"chk_member_levels_discount_rate_decimal_range",
		"check (discount_rate_decimal >= 0 and discount_rate_decimal <= 100)",
		"drop column if exists discount_rate",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

func TestRemoveMemberLevelLegacyDiscountRateMigrationRollbackRestoresLegacyColumn(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "338_remove_member_level_legacy_discount_rate.down.sql"))
	for _, fragment := range []string{
		"add column if not exists discount_rate numeric",
		"set discount_rate = discount_rate_decimal::numeric",
		"drop constraint if exists chk_member_levels_discount_rate_decimal_range",
		"drop column if exists discount_rate_decimal",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("rollback migration missing %q", fragment)
		}
	}
}
