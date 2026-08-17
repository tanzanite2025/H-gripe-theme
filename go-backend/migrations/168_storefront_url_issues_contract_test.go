package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readStorefrontURLIssueMigration(t *testing.T, name string) string {
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

func TestStorefrontURLIssuesMigrationContract(t *testing.T) {
	upSQL := readStorefrontURLIssueMigration(t, "168_storefront_url_issues.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS storefront_url_issues",
		"UNIQUE (route_entry_id, issue_type)",
		"CHECK (state IN ('open', 'acknowledged', 'resolved', 'verified', 'suppressed'))",
		"CREATE TABLE IF NOT EXISTS storefront_url_issue_events",
		"REFERENCES storefront_url_issues(id)",
		"metadata JSONB NOT NULL DEFAULT '{}'::jsonb",
		"idx_storefront_url_issue_events_issue_created",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("storefront URL issue migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readStorefrontURLIssueMigration(t, "168_storefront_url_issues.down.sql")
	for _, fragment := range []string{
		"DROP TABLE IF EXISTS storefront_url_issue_events",
		"DROP TABLE IF EXISTS storefront_url_issues",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("storefront URL issue down migration is missing contract fragment %q", fragment)
		}
	}
}
