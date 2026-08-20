package migrations_test

import (
	"strings"
	"testing"
)

func TestArchiveTerminalHotDataMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "196_archive_terminal_hot_data.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS site_quality_runs_archive",
		"CREATE TABLE IF NOT EXISTS after_sales_case_events_archive",
		"ALTER TABLE after_sales_cases",
		"ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ",
		"SET closed_at = updated_at",
		"DROP CONSTRAINT IF EXISTS fk_site_quality_findings_latest_run",
		"DROP CONSTRAINT IF EXISTS fk_site_quality_finding_events_run",
		"idx_site_quality_runs_archive_target_strategy_created",
		"idx_after_sales_case_events_archive_case_id_created_at",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("hot data archive up migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readMigrationFile(t, "196_archive_terminal_hot_data.down.sql")
	for _, fragment := range []string{
		"INSERT INTO site_quality_runs",
		"INSERT INTO after_sales_case_events",
		"ADD CONSTRAINT fk_site_quality_findings_latest_run",
		"ADD CONSTRAINT fk_site_quality_finding_events_run",
		"DROP TABLE IF EXISTS after_sales_case_events_archive",
		"DROP TABLE IF EXISTS site_quality_runs_archive",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("hot data archive down migration is missing contract fragment %q", fragment)
		}
	}
}
