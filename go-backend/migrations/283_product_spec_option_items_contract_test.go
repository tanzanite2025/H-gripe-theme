package migrations_test

import (
	"strings"
	"testing"
)

func TestProductSpecOptionItemsMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "283_product_spec_option_items.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "283_product_spec_option_items.down.sql"))

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS PRODUCT_SPEC_OPTION_ITEMS",
		"SPEC_DEFINITION_ID BIGINT NOT NULL",
		"REFERENCES PRODUCT_SPEC_DEFINITIONS(ID) ON DELETE CASCADE",
		"VALUE_KEY VARCHAR(160) NOT NULL",
		"DEFAULT_PRICE_DELTA_MINOR",
		"DEFAULT_PRICE_CURRENCY",
		"CONSTRAINT CK_PRODUCT_SPEC_OPTION_ITEM_PRICE_PAIR CHECK",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_PRODUCT_SPEC_OPTION_ITEM_DEFAULT",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS PRODUCT_SPEC_OPTION_ITEMS") {
		t.Fatal("down migration must remove product_spec_option_items")
	}
}
