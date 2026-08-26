package migrations_test

import (
	"strings"
	"testing"
)

func TestShippingSourceCurrencyMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "224_add_shipping_source_currency.up.sql")
	downSQL := readMigrationFile(t, "224_add_shipping_source_currency.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE shipping_templates",
		"ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD'",
		"ALTER TABLE shipping_rules",
		"idx_shipping_templates_currency",
		"idx_shipping_rules_currency",
		"chk_shipping_templates_currency_iso",
		"chk_shipping_rules_currency_iso",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("shipping currency migration is missing contract fragment %q", fragment)
		}
	}

	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS chk_shipping_rules_currency_iso",
		"DROP CONSTRAINT IF EXISTS chk_shipping_templates_currency_iso",
		"DROP COLUMN IF EXISTS currency",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("shipping currency down migration is missing contract fragment %q", fragment)
		}
	}
}

func TestCurrencyExchangeSyncLeaseMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "225_create_currency_exchange_sync_leases.up.sql")
	downSQL := readMigrationFile(t, "225_create_currency_exchange_sync_leases.down.sql")

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS currency_exchange_sync_leases",
		"lease_key VARCHAR(80) PRIMARY KEY",
		"owner_id VARCHAR(160) NOT NULL",
		"lease_expires_at TIMESTAMPTZ NOT NULL",
		"idx_currency_exchange_sync_leases_expires_at",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("exchange sync lease migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS currency_exchange_sync_leases") {
		t.Fatal("exchange sync lease down migration must drop only the lease table")
	}
}

func TestProductCurrencyMigrationDoesNotReintroduceGlobalDisplayCurrencies(t *testing.T) {
	upSQL := readMigrationFile(t, "083_product_price_currency_and_display_currencies.up.sql")
	if strings.Contains(upSQL, "INSERT INTO settings") && strings.Contains(upSQL, "currency_display_currencies") {
		t.Fatal("product currency migration must not recreate the legacy global display-currency setting")
	}
	if !strings.Contains(upSQL, "DELETE FROM settings") || !strings.Contains(upSQL, "currency_display_currencies") {
		t.Fatal("product currency migration must remove the legacy global display-currency setting")
	}
}
