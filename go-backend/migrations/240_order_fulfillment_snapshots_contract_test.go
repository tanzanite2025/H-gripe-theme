package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderFulfillmentSnapshotMigrationDefinesBackwardCompatibleDefaults(t *testing.T) {
	upSQL := readMigrationFile(t, "240_order_fulfillment_snapshots.up.sql")
	downSQL := readMigrationFile(t, "240_order_fulfillment_snapshots.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE products",
		"ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(20) NOT NULL DEFAULT 'stock'",
		"ALTER TABLE orders",
		"ADD COLUMN IF NOT EXISTS production_status VARCHAR(20) NOT NULL DEFAULT 'not_applicable'",
		"ALTER TABLE order_items",
		"CREATE INDEX IF NOT EXISTS idx_orders_production_status",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("fulfillment snapshot migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP COLUMN IF EXISTS") {
		t.Fatal("fulfillment snapshot down migration must remove only its columns")
	}
}
