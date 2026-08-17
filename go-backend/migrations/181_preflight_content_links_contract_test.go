package migrations_test

import (
	"strings"
	"testing"
)

func TestPreflightContentLinksMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "181_preflight_content_links.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS preflight_content_link_runs",
		"CREATE TABLE IF NOT EXISTS preflight_content_link_issues",
		"UNIQUE (issue_key)",
		"latest_evidence JSONB NOT NULL DEFAULT '{}'::jsonb",
		"CHECK (state IN ('open', 'resolved', 'verified', 'ignored'))",
		"CHECK (fix_status IN ('not_fixable', 'pending', 'applied', 'failed'))",
		"CREATE TABLE IF NOT EXISTS preflight_content_link_issue_events",
		"idx_preflight_content_link_issue_events_issue_created",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("content link preflight migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "181_preflight_content_links.down.sql")
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS preflight_content_link_issue_events") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS preflight_content_link_issues") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS preflight_content_link_runs") {
		t.Fatal("content link preflight down migration is missing table drops")
	}
}
