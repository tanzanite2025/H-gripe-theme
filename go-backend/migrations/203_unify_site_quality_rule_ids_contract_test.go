package migrations_test

import (
	"strings"
	"testing"
)

func TestUnifiedSiteQualityRuleIDsMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "203_unify_site_quality_rule_ids.up.sql")
	for _, fragment := range []string{
		"ALTER TABLE site_quality_findings",
		"ADD COLUMN IF NOT EXISTS rule_id VARCHAR(128) NOT NULL DEFAULT ''",
		"ADD COLUMN IF NOT EXISTS provider_audit_id VARCHAR(128) NOT NULL DEFAULT ''",
		"link_descriptive_text",
		"jsonb_set",
		"latest_evidence = jsonb_set",
		"CREATE INDEX IF NOT EXISTS idx_site_quality_findings_rule_id",
		"ALTER TABLE preflight_content_link_issues",
		"CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_rule_id",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("unified rule ID migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "203_unify_site_quality_rule_ids.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_preflight_content_link_issues_rule_id",
		"DROP INDEX IF EXISTS idx_site_quality_findings_rule_id",
		"DROP COLUMN IF EXISTS provider_audit_id",
		"DROP COLUMN IF EXISTS rule_id",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("unified rule ID down migration is missing contract fragment %q", fragment)
		}
	}
}

func TestContentLinkRunRuleIdentityMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "204_add_content_link_run_rule_identity.up.sql")
	for _, fragment := range []string{
		"ALTER TABLE preflight_content_link_runs",
		"ADD COLUMN IF NOT EXISTS rule_id VARCHAR(128) NOT NULL DEFAULT 'link_descriptive_text'",
		"ADD COLUMN IF NOT EXISTS provider_audit_id VARCHAR(128) NOT NULL DEFAULT 'link-text'",
		"CREATE INDEX IF NOT EXISTS idx_preflight_content_link_runs_rule_id",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("content link run identity migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "204_add_content_link_run_rule_identity.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_preflight_content_link_runs_rule_id",
		"DROP COLUMN IF EXISTS provider_audit_id",
		"DROP COLUMN IF EXISTS rule_id",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("content link run identity down migration is missing contract fragment %q", fragment)
		}
	}
}

func TestContentLinkIssueKeyNamespaceMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "206_namespace_content_link_issue_keys.up.sql")
	for _, fragment := range []string{
		"ALTER TABLE preflight_content_link_issues",
		"ALTER COLUMN issue_key TYPE VARCHAR(256)",
		"'content-link:' || rule_id || ':'",
		"idx_preflight_content_link_issues_rule_key",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("content link issue key migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "206_namespace_content_link_issue_keys.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_preflight_content_link_issues_rule_key",
		"ALTER COLUMN issue_key TYPE VARCHAR(128)",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("content link issue key down migration is missing contract fragment %q", fragment)
		}
	}
}
