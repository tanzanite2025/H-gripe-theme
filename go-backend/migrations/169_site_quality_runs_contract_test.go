package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readSiteQualityMigration(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test location")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(data)
}

func TestSiteQualityRunsMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "169_site_quality_runs.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS site_quality_runs",
		"provider VARCHAR(32) NOT NULL DEFAULT 'lighthouse_runner'",
		"canonical_url TEXT NOT NULL DEFAULT ''",
		"CHECK (strategy IN ('mobile', 'desktop'))",
		"CHECK (status IN ('success', 'failed'))",
		"issues_json JSONB NOT NULL DEFAULT '[]'::jsonb",
		"raw_response_json JSONB NOT NULL DEFAULT '{}'::jsonb",
		"idx_site_quality_runs_target_strategy_created",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("Site Quality runs migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "169_site_quality_runs.down.sql")
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS site_quality_runs") {
		t.Fatal("Site Quality runs down migration is missing the drop table contract")
	}
}
