package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveOrderItemLegacyAttributesMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "298_remove_order_item_legacy_attributes.up.sql"))
	if !strings.Contains(up, "ALTER TABLE ORDER_ITEMS") {
		t.Fatal("up migration must alter order_items")
	}
	if !strings.Contains(up, "DROP COLUMN IF EXISTS ATTRIBUTES") {
		t.Fatal("up migration must drop order_items.attributes")
	}
	if strings.TrimSpace(readMigrationFile(t, "298_remove_order_item_legacy_attributes.down.sql")) == "" {
		t.Fatal("down migration must document intentional non-restoration")
	}
}
