package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readOpsComposeServicesMigrationFile(t *testing.T, name string) string {
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

func TestReconcileOpsProjectComposeServicesUpMigrationContract(t *testing.T) {
	sql := readOpsComposeServicesMigrationFile(t, "158_reconcile_ops_project_compose_services.up.sql")
	for _, fragment := range []string{
		"UPDATE ops_project_bindings AS project",
		"project.name = 'commerce-platform'",
		"project.environment = 'production'",
		"btrim(service.name) = 'edge-config'",
		"array_to_string",
		"updated_at = NOW()",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestReconcileOpsProjectComposeServicesDownMigrationContract(t *testing.T) {
	sql := readOpsComposeServicesMigrationFile(t, "158_reconcile_ops_project_compose_services.down.sql")
	for _, fragment := range []string{
		"UPDATE ops_project_bindings AS project",
		"btrim(service.name) <> 'edge-config'",
		"project.name = 'commerce-platform'",
		"project.environment = 'production'",
		"updated_at = NOW()",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
