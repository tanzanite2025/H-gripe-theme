package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderSignaturePolicySnapshotMigrationDefinesBackwardCompatibleDefaults(t *testing.T) {
	upSQL := readMigrationFile(t, "241_order_signature_policy_snapshot.up.sql")
	downSQL := readMigrationFile(t, "241_order_signature_policy_snapshot.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE orders",
		"ADD COLUMN IF NOT EXISTS signature_required BOOLEAN NOT NULL DEFAULT FALSE",
		"CREATE INDEX IF NOT EXISTS idx_orders_signature_required",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("order signature policy migration is missing contract fragment %q", fragment)
		}
	}

	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_orders_signature_required",
		"DROP COLUMN IF EXISTS signature_required",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("order signature policy rollback is missing contract fragment %q", fragment)
		}
	}
}
