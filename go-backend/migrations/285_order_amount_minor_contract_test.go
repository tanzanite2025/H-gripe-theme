package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderAmountMinorMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "285_order_amount_minor.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "285_order_amount_minor.down.sql"))
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS PAYMENT_AMOUNT_MINOR BIGINT",
		"ADD COLUMN IF NOT EXISTS SUBTOTAL_AMOUNT_MINOR BIGINT",
		"ADD COLUMN IF NOT EXISTS TOTAL_AMOUNT_MINOR BIGINT",
		"UPDATE ORDERS",
		"UPDATE ORDER_ITEMS",
		"CHK_ORDERS_TOTAL_AMOUNT_MINOR_NON_NEGATIVE",
		"CHK_ORDER_ITEMS_TOTAL_MINOR_NON_NEGATIVE",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	for _, required := range []string{
		"DROP COLUMN IF EXISTS PAYMENT_AMOUNT_MINOR",
		"DROP COLUMN IF EXISTS TOTAL_AMOUNT_MINOR",
		"DROP COLUMN IF EXISTS TOTAL_MINOR",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
