package migrations_test

import (
	"strings"
	"testing"
)

func TestSiteQualityAuditScopeMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "346_site_quality_audit_scope.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS audit_scope VARCHAR(16) NOT NULL DEFAULT 'full'",
		"CHECK (audit_scope IN ('full', 'headings', 'schema', 'link_text'))",
		"idx_site_quality_jobs_audit_scope",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("audit scope migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "346_site_quality_audit_scope.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_site_quality_jobs_audit_scope",
		"DROP CONSTRAINT IF EXISTS ck_site_quality_jobs_audit_scope",
		"DROP COLUMN IF EXISTS audit_scope",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("audit scope down migration is missing contract fragment %q", fragment)
		}
	}
}
