package migrations_test

import (
	"strings"
	"testing"
)

func TestRefundAmountMinorMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "288_refund_amount_minor.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "288_refund_amount_minor.down.sql"))
	for _, required := range []string{"ADD COLUMN IF NOT EXISTS AMOUNT_MINOR BIGINT", "REQUESTED_AMOUNT_MINOR", "UPDATE REFUNDS", "CHK_REFUNDS_AMOUNT_MINOR_NON_NEGATIVE"} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	for _, required := range []string{"DROP COLUMN IF EXISTS AMOUNT_MINOR", "DROP COLUMN IF EXISTS CURRENCY"} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
