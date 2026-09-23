package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readCustomerServiceInboxArchiveMigration(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test location")
	}
	path := filepath.Join(filepath.Dir(file), name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestCustomerServiceInboxArchiveMigrationContract(t *testing.T) {
	upSQL := readCustomerServiceInboxArchiveMigration(t, "303_customer_service_inbox_archive.up.sql")
	for _, fragment := range []string{
		"ALTER TABLE customer_service_inbox_states",
		"ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL",
		"idx_customer_service_inbox_states_recipient_archived",
		"idx_customer_service_inbox_states_ticket_archived",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}

	downSQL := readCustomerServiceInboxArchiveMigration(t, "303_customer_service_inbox_archive.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_customer_service_inbox_states_ticket_archived",
		"DROP INDEX IF EXISTS idx_customer_service_inbox_states_recipient_archived",
		"DROP COLUMN IF EXISTS archived_at",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
