package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestYanwenTrackingPollingScheduleMigrationSeparatesOfficialFactsFromWorkerState(t *testing.T) {
	upSQL, err := os.ReadFile("394_add_yanwen_tracking_polling_schedule.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.ToLower(string(upSQL))
	for _, requiredFragment := range []string{
		"add column if not exists has_official_result boolean",
		"add column if not exists polling_state varchar(16)",
		"add column if not exists next_sync_at timestamptz",
		"add column if not exists polling_lease_token varchar(64)",
		"add column if not exists polling_lease_until timestamptz",
		"add column if not exists last_polling_error varchar(1000)",
		"insert into shipping_yanwen_tracking_snapshots",
		"from shipping_yanwen_waybills as waybill",
		"waybill.environment = 'production'",
		"on conflict (environment, tracking_number) do nothing",
	} {
		if !strings.Contains(up, requiredFragment) {
			t.Fatalf("polling schedule migration is missing %q", requiredFragment)
		}
	}

	downSQL, err := os.ReadFile("394_add_yanwen_tracking_polling_schedule.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	down := strings.ToLower(string(downSQL))
	for _, requiredFragment := range []string{
		"drop column if exists has_official_result",
		"drop column if exists polling_state",
		"drop column if exists next_sync_at",
		"drop column if exists polling_lease_token",
		"drop column if exists polling_lease_until",
		"drop column if exists last_polling_error",
	} {
		if !strings.Contains(down, requiredFragment) {
			t.Fatalf("polling schedule rollback is missing %q", requiredFragment)
		}
	}
}
