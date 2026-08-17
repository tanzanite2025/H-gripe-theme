package migrations_test

import (
	"strings"
	"testing"
)

func TestSiteQualityEngineMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "171_site_quality_engine.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS site_quality_targets",
		"CREATE TABLE IF NOT EXISTS site_quality_jobs",
		"CREATE TABLE IF NOT EXISTS site_quality_provider_slots",
		"CREATE TABLE IF NOT EXISTS site_quality_evaluations",
		"site_quality_runs",
		"uq_site_quality_findings_target_strategy_audit",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("site quality engine migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "171_site_quality_engine.down.sql")
	for _, fragment := range []string{
		"DROP TABLE IF EXISTS site_quality_evaluations",
		"DROP TABLE IF EXISTS site_quality_provider_slots",
		"DROP TABLE IF EXISTS site_quality_jobs",
		"DROP TABLE IF EXISTS site_quality_targets",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("site quality engine down migration is missing contract fragment %q", fragment)
		}
	}
}
