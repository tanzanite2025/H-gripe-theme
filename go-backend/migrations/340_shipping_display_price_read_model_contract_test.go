package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingDisplayPriceReadModelMigrationContract(t *testing.T) {
	upBytes, err := os.ReadFile("340_shipping_display_price_read_model.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.ToLower(string(upBytes))
	for _, fragment := range []string{
		"create table if not exists shipping_display_price_snapshots",
		"references shipping_templates(id)",
		"references shipping_rules(id)",
		"insert into shipping_display_price_snapshots",
		"drop column if exists display_price_snapshots",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	downBytes, err := os.ReadFile("340_shipping_display_price_read_model.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	down := strings.ToLower(string(downBytes))
	for _, fragment := range []string{
		"add column if not exists display_price_snapshots jsonb",
		"update shipping_templates",
		"update shipping_rules",
		"drop table if exists shipping_display_price_snapshots",
	} {
		if !strings.Contains(down, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
