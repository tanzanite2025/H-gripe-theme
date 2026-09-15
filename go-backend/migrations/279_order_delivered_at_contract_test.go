package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderDeliveredAtMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "279_order_delivered_at.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "279_order_delivered_at.down.sql"))

	for _, required := range []string{
		"ALTER TABLE ORDERS",
		"ADD COLUMN IF NOT EXISTS DELIVERED_AT TIMESTAMPTZ",
		"CREATE INDEX IF NOT EXISTS IDX_ORDERS_DELIVERED_AT",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	if !strings.Contains(downSQL, "ALTER TABLE ORDERS DROP COLUMN IF EXISTS DELIVERED_AT") {
		t.Fatal("down migration must remove orders.delivered_at")
	}
}
