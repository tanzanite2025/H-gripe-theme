package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCreateShippingYanwenWarehouseCatalogEntriesMigrationDefinesEnvironmentScopedOfficialWarehouses(t *testing.T) {
	contents, err := os.ReadFile("382_create_shipping_yanwen_warehouse_catalog_entries.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, fragment := range []string{
		"create table if not exists shipping_yanwen_warehouse_catalog_entries",
		"environment varchar(16) not null",
		"warehouse_code varchar(80) not null",
		"name varchar(200) not null",
		"area varchar(200) not null default ''",
		"last_synced_at timestamptz not null",
		"unique (environment, warehouse_code)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing required fragment %q", fragment)
		}
	}
}

func TestDropShippingYanwenWarehouseCatalogEntriesMigrationRemovesWarehouseTable(t *testing.T) {
	contents, err := os.ReadFile("382_create_shipping_yanwen_warehouse_catalog_entries.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(contents)), "drop table if exists shipping_yanwen_warehouse_catalog_entries") {
		t.Fatal("migration rollback is missing warehouse catalog table drop")
	}
}
