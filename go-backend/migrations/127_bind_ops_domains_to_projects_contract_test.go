package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readMigrationFile(t *testing.T, name string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(data)
}

func TestBindOpsDomainsToProjectsUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "127_bind_ops_domains_to_projects.up.sql")

	requiredFragments := []string{
		"ADD COLUMN IF NOT EXISTS project_binding_id BIGINT NULL",
		"CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_project",
		"conname = 'fk_ops_domain_binding_project'",
		"conrelid = 'ops_domain_bindings'::regclass",
		"ON DELETE SET NULL",
		"domain.project_binding_id IS NULL",
		"project.name = 'commerce-platform'",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestBindOpsDomainsToProjectsDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "127_bind_ops_domains_to_projects.down.sql")

	requiredFragments := []string{
		"DROP CONSTRAINT IF EXISTS fk_ops_domain_binding_project",
		"DROP INDEX IF EXISTS idx_ops_domain_binding_project",
		"DROP COLUMN IF EXISTS project_binding_id",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
