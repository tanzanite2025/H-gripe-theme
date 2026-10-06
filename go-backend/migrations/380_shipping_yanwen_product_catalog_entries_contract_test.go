package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCreateShippingYanwenProductCatalogEntriesMigrationDefinesEnvironmentScopedOfficialProducts(t *testing.T) {
	contents, err := os.ReadFile("380_create_shipping_yanwen_product_catalog_entries.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, fragment := range []string{
		"create table if not exists shipping_yanwen_product_catalog_entries",
		"environment varchar(16) not null",
		"product_id varchar(80) not null",
		"name_ch varchar(200) not null default ''",
		"name_en varchar(200) not null default ''",
		"last_synced_at timestamptz not null",
		"unique (environment, product_id)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing required fragment %q", fragment)
		}
	}
}

