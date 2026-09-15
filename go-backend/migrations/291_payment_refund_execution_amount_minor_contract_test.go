package migrations_test

import (
	"strings"
	"testing"
)

func TestPaymentRefundExecutionAmountMinorMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "291_payment_refund_execution_amount_minor.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "291_payment_refund_execution_amount_minor.down.sql"))
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS AMOUNT_MINOR BIGINT",
		"UPDATE PAYMENT_REFUND_EXECUTIONS",
		"ALTER COLUMN AMOUNT_MINOR SET NOT NULL",
		"CHK_PAYMENT_REFUND_EXECUTIONS_AMOUNT_MINOR_NON_NEGATIVE",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	for _, required := range []string{
		"DROP CONSTRAINT IF EXISTS CHK_PAYMENT_REFUND_EXECUTIONS_AMOUNT_MINOR_NON_NEGATIVE",
		"DROP COLUMN IF EXISTS AMOUNT_MINOR",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
