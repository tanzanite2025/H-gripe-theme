package migrations_test

import (
	"strings"
	"testing"
)

func TestPaymentOperationIdempotenciesMigrationDefinesDurableUniqueFence(t *testing.T) {
	sql := readMigrationFile(t, "261_payment_operation_idempotencies.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS payment_operation_idempotencies",
		"request_hash VARCHAR(64) NOT NULL",
		"status VARCHAR(16) NOT NULL",
		"response_body TEXT NOT NULL",
		"UNIQUE (user_id, scope, idempotency_key)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestPaymentOperationIdempotenciesDownMigrationDropsTable(t *testing.T) {
	sql := readMigrationFile(t, "261_payment_operation_idempotencies.down.sql")
	if !strings.Contains(sql, "DROP TABLE IF EXISTS payment_operation_idempotencies") {
		t.Fatal("down migration must drop payment_operation_idempotencies")
	}
}
