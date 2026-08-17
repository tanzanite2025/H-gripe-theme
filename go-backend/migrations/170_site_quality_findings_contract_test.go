package migrations_test

import (
	"strings"
	"testing"
)

func TestSiteQualityFindingsMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "170_site_quality_findings.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS site_quality_findings",
		"UNIQUE (target_url, strategy, audit_id)",
		"consecutive_clean INTEGER NOT NULL DEFAULT 0",
		"CHECK (state IN ('open', 'acknowledged', 'resolved', 'verified'))",
		"latest_evidence JSONB NOT NULL DEFAULT '{}'::jsonb",
		"CREATE TABLE IF NOT EXISTS site_quality_finding_events",
		"idx_site_quality_finding_events_finding_created",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("Site Quality findings migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "170_site_quality_findings.down.sql")
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS site_quality_finding_events") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS site_quality_findings") {
		t.Fatal("Site Quality findings down migration is missing table drops")
	}
}
