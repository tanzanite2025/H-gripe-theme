package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveExchangeRateLegacyFloatMigrationUsesRateDecimalOnly(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "337_remove_exchange_rate_legacy_float.up.sql"))
	for _, fragment := range []string{
		"rate_decimal type numeric(30,15)",
		"set rate_decimal",
		"chk_currency_exchange_rates_rate_decimal_positive",
		"drop column if exists rate",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	if strings.Contains(sql, "add column if not exists rate ") {
		t.Fatal("up migration must not recreate legacy rate")
	}
}

func TestRemoveExchangeRateLegacyFloatMigrationRollbackRestoresLegacyRate(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "337_remove_exchange_rate_legacy_float.down.sql"))
	for _, fragment := range []string{
		"add column if not exists rate numeric(20,10)",
		"set rate = rate_decimal::numeric(20,10)",
		"drop column if exists rate_decimal",
		"check (rate > 0)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("rollback migration missing %q", fragment)
		}
	}
}
