package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderFulfillmentHoldMigrationDefinesBackwardCompatibleDefaults(t *testing.T) {
	upSQL := readMigrationFile(t, "256_order_fulfillment_hold.up.sql")
	downSQL := readMigrationFile(t, "256_order_fulfillment_hold.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE orders",
		"ADD COLUMN IF NOT EXISTS fulfillment_hold BOOLEAN NOT NULL DEFAULT FALSE",
		"CREATE INDEX IF NOT EXISTS idx_orders_fulfillment_hold",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("fulfillment hold migration is missing contract fragment %q", fragment)
		}
	}
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_orders_fulfillment_hold",
		"DROP COLUMN IF EXISTS fulfillment_hold",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("fulfillment hold rollback is missing contract fragment %q", fragment)
		}
	}
}
