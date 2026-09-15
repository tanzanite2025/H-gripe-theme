package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderItemPricingSnapshotImmutabilityMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "274_order_item_pricing_snapshot_immutable.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "274_order_item_pricing_snapshot_immutable.down.sql"))
	for _, required := range []string{
		"CREATE OR REPLACE FUNCTION PREVENT_ORDER_ITEM_PRICING_SNAPSHOT_MUTATION",
		"OLD.PRICING_SNAPSHOT <> '{}'::JSONB",
		"NEW.PRICING_SNAPSHOT IS DISTINCT FROM OLD.PRICING_SNAPSHOT",
		"CREATE TRIGGER TRIGGER_PREVENT_ORDER_ITEM_PRICING_SNAPSHOT_MUTATION",
		"BEFORE UPDATE OF PRICING_SNAPSHOT ON ORDER_ITEMS",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	if !strings.Contains(downSQL, "DROP TRIGGER IF EXISTS TRIGGER_PREVENT_ORDER_ITEM_PRICING_SNAPSHOT_MUTATION ON ORDER_ITEMS") ||
		!strings.Contains(downSQL, "DROP FUNCTION IF EXISTS PREVENT_ORDER_ITEM_PRICING_SNAPSHOT_MUTATION") {
		t.Fatal("down migration must remove the pricing snapshot immutability trigger and function")
	}
}
