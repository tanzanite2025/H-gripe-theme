package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCreateShippingYanwenCountryCatalogEntriesMigrationDefinesEnvironmentScopedOfficialCountries(t *testing.T) {
	contents, err := os.ReadFile("381_create_shipping_yanwen_country_catalog_entries.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, fragment := range []string{
		"create table if not exists shipping_yanwen_country_catalog_entries",
		"environment varchar(16) not null",
		"country_id varchar(80) not null",
		"country_code varchar(16) not null",
		"name_ch varchar(200) not null default ''",
		"name_en varchar(200) not null default ''",
		"last_synced_at timestamptz not null",
		"unique (environment, country_id)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing required fragment %q", fragment)
		}
	}
}

func TestDropShippingYanwenCountryCatalogEntriesMigrationRemovesCountryTable(t *testing.T) {
	contents, err := os.ReadFile("381_create_shipping_yanwen_country_catalog_entries.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(contents)), "drop table if exists shipping_yanwen_country_catalog_entries") {
		t.Fatal("migration rollback is missing country catalog table drop")
	}
}
