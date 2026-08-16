package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readCustomerServiceInboxStateMigration(t *testing.T, name string) string {
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

func TestCustomerServiceInboxStatesMigrationContract(t *testing.T) {
	upSQL := readCustomerServiceInboxStateMigration(t, "164_customer_service_inbox_states.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS customer_service_inbox_states",
		"last_read_message_id BIGINT NOT NULL DEFAULT 0",
		"unread_count INTEGER NOT NULL DEFAULT 0",
		"UNIQUE (recipient_user_id, ticket_id)",
		"idx_customer_service_inbox_states_recipient_unread",
		"INSERT INTO customer_service_inbox_states",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("customer-service inbox migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readCustomerServiceInboxStateMigration(t, "164_customer_service_inbox_states.down.sql")
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS customer_service_inbox_states") {
		t.Fatal("customer-service inbox down migration must remove the state table")
	}
}
