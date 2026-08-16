package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readCustomerServiceStatusVersionMigration(t *testing.T, name string) string {
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

func TestCustomerServiceStatusVersionMigrationContract(t *testing.T) {
	upSQL := readCustomerServiceStatusVersionMigration(t, "165_customer_service_status_versions.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS status_version INTEGER NOT NULL DEFAULT 1",
		"SET status_version = 1",
		"ck_tickets_status_version_positive",
		"CHECK (status_version > 0)",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("customer-service status version migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readCustomerServiceStatusVersionMigration(t, "165_customer_service_status_versions.down.sql")
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS ck_tickets_status_version_positive",
		"DROP COLUMN IF EXISTS status_version",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("customer-service status version down migration is missing contract fragment %q", fragment)
		}
	}
}
