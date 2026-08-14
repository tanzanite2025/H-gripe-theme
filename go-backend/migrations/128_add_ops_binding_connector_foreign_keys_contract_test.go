package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readOpsBindingConnectorMigrationFile(t *testing.T, name string) string {
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

func TestAddOpsBindingConnectorForeignKeysMigrationContract(t *testing.T) {
	sql := readOpsBindingConnectorMigrationFile(t, "128_add_ops_binding_connector_foreign_keys.up.sql")
	for _, fragment := range []string{
		"SET connector_id = NULL",
		"fk_ops_vps_binding_connector",
		"fk_ops_project_binding_connector",
		"conrelid = 'ops_vps_bindings'::regclass",
		"conrelid = 'ops_project_bindings'::regclass",
		"ON DELETE SET NULL",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestDropOpsBindingConnectorForeignKeysMigrationContract(t *testing.T) {
	sql := readOpsBindingConnectorMigrationFile(t, "128_add_ops_binding_connector_foreign_keys.down.sql")
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS fk_ops_project_binding_connector",
		"DROP CONSTRAINT IF EXISTS fk_ops_vps_binding_connector",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
