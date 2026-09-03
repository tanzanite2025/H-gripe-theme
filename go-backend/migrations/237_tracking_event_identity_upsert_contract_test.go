package migrations_test

import (
	"strings"
	"testing"
)

func TestTrackingEventIdentityUpsertMigrationContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "237_tracking_event_identity_upsert.up.sql"))
	for _, fragment := range []string{
		"row_number() over",
		"partition by order_id, tracking_number, event_time, status",
		"duplicate_rank > 1",
		"create unique index if not exists idx_tracking_events_identity",
		"on tracking_events (order_id, tracking_number, event_time, status)",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("tracking event identity migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "237_tracking_event_identity_upsert.down.sql"))
	if !strings.Contains(downSQL, "drop index if exists idx_tracking_events_identity") {
		t.Fatal("tracking event identity rollback is missing unique index cleanup")
	}
}
