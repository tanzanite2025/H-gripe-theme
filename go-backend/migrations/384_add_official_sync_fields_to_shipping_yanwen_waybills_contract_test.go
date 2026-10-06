package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestYanwenWaybillOfficialSyncMigrationAddsOfficialFields(t *testing.T) {
	contents, err := os.ReadFile("384_add_official_sync_fields_to_shipping_yanwen_waybills.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(contents))
	for _, required := range []string{
		"alter table shipping_yanwen_waybills",
		"reference_number varchar(100)",
		"official_status integer",
		"is_printed boolean",
		"last_official_synced_at timestamptz",
		"idx_shipping_yanwen_waybills_official_status",
	} {
		if !strings.Contains(lower, required) {
			t.Fatalf("Yanwen official sync migration is missing %q", required)
		}
	}
}

func TestYanwenWaybillOfficialSyncMigrationCanBeRolledBack(t *testing.T) {
	contents, err := os.ReadFile("384_add_official_sync_fields_to_shipping_yanwen_waybills.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(contents))
	for _, required := range []string{
		"drop index if exists idx_shipping_yanwen_waybills_official_status",
		"drop column if exists reference_number",
		"drop column if exists official_status",
		"drop column if exists is_printed",
		"drop column if exists last_official_synced_at",
	} {
		if !strings.Contains(lower, required) {
			t.Fatalf("Yanwen official sync rollback is missing %q", required)
		}
	}
}
