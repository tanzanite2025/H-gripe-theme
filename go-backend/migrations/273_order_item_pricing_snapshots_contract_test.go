package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderItemPricingSnapshotsMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "273_order_item_pricing_snapshots.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "273_order_item_pricing_snapshots.down.sql"))
	if !strings.Contains(upSQL, "ALTER TABLE ORDER_ITEMS") ||
		!strings.Contains(upSQL, "ADD COLUMN IF NOT EXISTS PRICING_SNAPSHOT JSONB NOT NULL") {
		t.Fatal("up migration must add the immutable order-item pricing snapshot column")
	}
	if !strings.Contains(downSQL, "ALTER TABLE ORDER_ITEMS") ||
		!strings.Contains(downSQL, "DROP COLUMN IF EXISTS PRICING_SNAPSHOT") {
		t.Fatal("down migration must remove the order-item pricing snapshot column")
	}
}
