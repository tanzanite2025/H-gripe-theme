package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderCheckoutCartReferenceUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "253_order_checkout_cart_reference.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS checkout_cart_id BIGINT NULL",
		"CREATE INDEX IF NOT EXISTS idx_orders_checkout_cart_id",
		"fk_orders_checkout_cart",
		"REFERENCES carts(id)",
		"ON DELETE SET NULL",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestOrderCheckoutCartReferenceDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "253_order_checkout_cart_reference.down.sql")
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS fk_orders_checkout_cart",
		"DROP INDEX IF EXISTS idx_orders_checkout_cart_id",
		"DROP COLUMN IF EXISTS checkout_cart_id",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
