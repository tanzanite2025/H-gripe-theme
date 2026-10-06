package migrations_test

import (
	"strings"
	"testing"
)

func TestShippingServiceCollectionCountryScopesMigrationContract(t *testing.T) {
	up := strings.ToLower(readMigrationFile(t, "376_add_shipping_service_collection_country_scopes.up.sql"))
	for _, required := range []string{
		"alter table shipping_fpx_channels",
		"alter table shipping_yanwen_published_channels",
		"add column if not exists countries text not null default '[]'",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("country scope migration is missing contract fragment %q", required)
		}
	}

	down := strings.ToLower(readMigrationFile(t, "376_add_shipping_service_collection_country_scopes.down.sql"))
	for _, required := range []string{
		"alter table shipping_yanwen_published_channels",
		"alter table shipping_fpx_channels",
		"drop column if exists countries",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("country scope rollback is missing contract fragment %q", required)
		}
	}
}
