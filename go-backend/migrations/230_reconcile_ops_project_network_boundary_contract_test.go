package migrations_test

import (
	"strings"
	"testing"
)

func TestReconcileOpsProjectNetworkBoundaryUpMigrationContract(t *testing.T) {
	sql := readOpsComposeServicesMigrationFile(t, "230_reconcile_ops_project_network_boundary.up.sql")
	for _, fragment := range []string{
		"UPDATE ops_project_bindings AS project",
		"networks = 'db, cache, app, api_ingress, shared-edge'",
		"project.name = 'commerce-platform'",
		"project.environment = 'production'",
		"updated_at = NOW()",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestReconcileOpsProjectNetworkBoundaryDownMigrationContract(t *testing.T) {
	sql := readOpsComposeServicesMigrationFile(t, "230_reconcile_ops_project_network_boundary.down.sql")
	for _, fragment := range []string{
		"UPDATE ops_project_bindings AS project",
		"networks = 'db, cache, app, shared-edge'",
		"project.name = 'commerce-platform'",
		"project.environment = 'production'",
		"updated_at = NOW()",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
