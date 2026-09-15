package migrations_test

import (
	"strings"
	"testing"
)

func TestRefundLineItemAmountMinorMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "290_refund_line_item_amount_minor.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "290_refund_line_item_amount_minor.down.sql"))
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS CURRENCY VARCHAR(3)",
		"ADD COLUMN IF NOT EXISTS UNIT_PRICE_MINOR BIGINT",
		"LINE_SUBTOTAL_MINOR",
		"LINE_TAX_MINOR",
		"LINE_DISCOUNT_MINOR",
		"LINE_TOTAL_MINOR",
		"UPDATE REFUND_LINE_ITEMS",
		"CHK_REFUND_LINE_ITEMS_TOTAL_MINOR_NON_NEGATIVE",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	for _, required := range []string{
		"DROP COLUMN IF EXISTS UNIT_PRICE_MINOR",
		"DROP COLUMN IF EXISTS CURRENCY",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
