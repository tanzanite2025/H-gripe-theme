package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestFpxAPISyncStatsMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("353_fpx_api_sync_stats.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("353_fpx_api_sync_stats.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	down := strings.ToLower(string(downSQL))
	for _, column := range []string{
		"last_sync_scanned",
		"last_sync_added",
		"last_sync_updated",
		"last_sync_preserved_enabled",
	} {
		if !strings.Contains(up, "add column if not exists "+column) {
			t.Fatalf("up migration missing %q", column)
		}
		if !strings.Contains(down, "drop column if exists "+column) {
			t.Fatalf("down migration missing %q", column)
		}
	}
}
