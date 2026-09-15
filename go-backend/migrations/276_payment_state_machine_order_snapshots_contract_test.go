package migrations_test

import (
	"strings"
	"testing"
)

func TestPaymentStateMachineOrderSnapshotMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "276_payment_state_machine_order_snapshots.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "276_payment_state_machine_order_snapshots.down.sql"))
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS PAYMENT_CURRENCY VARCHAR(3)",
		"ADD COLUMN IF NOT EXISTS PAYMENT_AMOUNT NUMERIC(18,2)",
		"ADD COLUMN IF NOT EXISTS DISPUTE_PREVIOUS_STATUS VARCHAR(32)",
		"ADD COLUMN IF NOT EXISTS DISPUTE_PREVIOUS_HOLD BOOLEAN",
		"CREATE INDEX IF NOT EXISTS IDX_ORDERS_PAYMENT_CURRENCY",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	for _, required := range []string{
		"DROP COLUMN IF EXISTS PAYMENT_CURRENCY",
		"DROP COLUMN IF EXISTS PAYMENT_AMOUNT",
		"DROP COLUMN IF EXISTS DISPUTE_PREVIOUS_STATUS",
		"DROP COLUMN IF EXISTS DISPUTE_PREVIOUS_HOLD",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
