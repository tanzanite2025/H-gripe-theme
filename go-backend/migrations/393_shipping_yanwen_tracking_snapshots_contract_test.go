package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCreateYanwenTrackingSnapshotsMigrationDefinesIsolatedLatestResultStore(t *testing.T) {
	upSQL, err := os.ReadFile("393_create_shipping_yanwen_tracking_snapshots.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	contents := strings.ToLower(string(upSQL))
	for _, requiredFragment := range []string{
		"create table if not exists shipping_yanwen_tracking_snapshots",
		"tracking_number varchar(120) not null",
		"checkpoints_data jsonb not null",
		"response_data jsonb not null",
		"check (environment = 'production')",
		"unique (environment, tracking_number)",
	} {
		if !strings.Contains(contents, requiredFragment) {
			t.Fatalf("migration is missing required fragment %q", requiredFragment)
		}
	}

	downSQL, err := os.ReadFile("393_create_shipping_yanwen_tracking_snapshots.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(downSQL)), "drop table if exists shipping_yanwen_tracking_snapshots") {
		t.Fatal("down migration must drop the Yanwen tracking snapshot table")
	}
}
