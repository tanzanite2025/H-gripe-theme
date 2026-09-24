package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveCarrierServiceLegacyFuelSurchargeMigrationUsesDecimalOnly(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "339_remove_carrier_service_legacy_fuel_surcharge_percent.up.sql"))
	for _, fragment := range []string{
		"add column if not exists fuel_surcharge_percent_decimal numeric(30,15)",
		"set fuel_surcharge_percent_decimal",
		"round(coalesce(fuel_surcharge_percent, 0)::numeric, 15)",
		"chk_shipping_carrier_services_fuel_surcharge_percent_decimal_range",
		"check (fuel_surcharge_percent_decimal >= 0 and fuel_surcharge_percent_decimal <= 100)",
		"drop column if exists fuel_surcharge_percent",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	if strings.Contains(sql, "add column if not exists fuel_surcharge_percent ") {
		t.Fatal("up migration must not recreate the legacy float percentage")
	}
}

func TestRemoveCarrierServiceLegacyFuelSurchargeMigrationRollbackRestoresLegacyColumn(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "339_remove_carrier_service_legacy_fuel_surcharge_percent.down.sql"))
	for _, fragment := range []string{
		"add column if not exists fuel_surcharge_percent numeric(8,3)",
		"set fuel_surcharge_percent = fuel_surcharge_percent_decimal::numeric(8,3)",
		"drop constraint if exists chk_shipping_carrier_services_fuel_surcharge_percent_decimal_range",
		"drop column if exists fuel_surcharge_percent_decimal",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("rollback migration missing %q", fragment)
		}
	}
}
