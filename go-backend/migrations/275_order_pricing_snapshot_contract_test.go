package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderPricingSnapshotMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "275_order_pricing_snapshot.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "275_order_pricing_snapshot.down.sql"))
	for _, required := range []string{
		"ALTER TABLE ORDERS",
		"ADD COLUMN IF NOT EXISTS PRICING_SNAPSHOT JSONB NOT NULL",
		"CREATE OR REPLACE FUNCTION PREVENT_ORDER_PRICING_SNAPSHOT_MUTATION",
		"BEFORE UPDATE OF PRICING_SNAPSHOT ON ORDERS",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	if !strings.Contains(downSQL, "ALTER TABLE ORDERS DROP COLUMN IF EXISTS PRICING_SNAPSHOT") {
		t.Fatal("down migration must remove the order pricing snapshot column")
	}
}
