package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingCarrierSurchargeMinorMigration(t *testing.T) {
	data, err := os.ReadFile("323_shipping_carrier_surcharge_minor.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"add column if not exists remote_surcharge_minor bigint",
		"chk_shipping_carrier_services_remote_surcharge_minor_non_negative",
		"drop column if exists remote_surcharge",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
