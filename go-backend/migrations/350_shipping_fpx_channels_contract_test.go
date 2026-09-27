package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingFpxChannelsMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("350_shipping_fpx_channels.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("350_shipping_fpx_channels.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"create table if not exists shipping_fpx_channels",
		"service_code varchar(80)",
		"display_name varchar(160)",
		"max_weight_grams",
		"volumetric_divisor",
		"where deleted_at is null",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}
	if !strings.Contains(strings.ToLower(string(downSQL)), "drop table if exists shipping_fpx_channels") {
		t.Fatal("down migration must drop shipping_fpx_channels")
	}
}
