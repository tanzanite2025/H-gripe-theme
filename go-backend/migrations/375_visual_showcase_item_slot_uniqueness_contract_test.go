package migrations_test

import (
	"strings"
	"testing"
)

func TestVisualShowcaseItemSlotUniquenessMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "375_visual_showcase_item_slot_uniqueness.up.sql"))
	for _, required := range []string{
		"PARTITION BY SHOWCASE_KEY, LOCALE, DESKTOP_ORDER",
		"WHERE DELETED_AT IS NULL",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_VISUAL_SHOWCASE_ITEMS_ACTIVE_SLOT",
		"ON VISUAL_SHOWCASE_ITEMS (SHOWCASE_KEY, LOCALE, DESKTOP_ORDER)",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("visual showcase slot migration is missing contract fragment %q", required)
		}
	}

	down := strings.ToUpper(readMigrationFile(t, "375_visual_showcase_item_slot_uniqueness.down.sql"))
	if !strings.Contains(down, "DROP INDEX IF EXISTS UQ_VISUAL_SHOWCASE_ITEMS_ACTIVE_SLOT") {
		t.Fatal("visual showcase slot down migration must remove only its own index")
	}
}
