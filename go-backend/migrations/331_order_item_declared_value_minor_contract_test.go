package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderItemDeclaredValueMinorMigrationRemovesLegacyMajorColumn(t *testing.T) {
	upSQL := readMigrationFile(t, "331_order_item_declared_value_minor.up.sql")
	lower := strings.ToLower(upSQL)
	for _, fragment := range []string{
		"add column if not exists declared_value_minor bigint",
		"drop column if exists declared_value",
		"chk_order_items_declared_value_minor_non_negative",
	} {
		if !strings.Contains(lower, fragment) {
			t.Fatalf("declared value migration missing contract fragment %q", fragment)
		}
	}
}
