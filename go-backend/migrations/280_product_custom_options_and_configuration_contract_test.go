package migrations_test

import (
	"strings"
	"testing"
)

func TestProductCustomOptionsAndConfigurationMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "280_product_custom_options_and_configuration.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "280_product_custom_options_and_configuration.down.sql"))

	for _, required := range []string{
		"ALTER TABLE PRODUCT_SPEC_DEFINITIONS",
		"ADD COLUMN IF NOT EXISTS ROLE VARCHAR(24)",
		"ADD COLUMN IF NOT EXISTS SELECTION_MODE VARCHAR(16)",
		"CREATE TABLE IF NOT EXISTS PRODUCT_CUSTOM_OPTION_POLICIES",
		"PRICE_DELTA_MINOR BIGINT",
		"ADD COLUMN IF NOT EXISTS CONFIGURATION JSONB",
		"ADD COLUMN IF NOT EXISTS CONFIGURATION_HASH CHAR(64)",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_CART_ITEMS_IDENTITY",
		"ADD COLUMN IF NOT EXISTS CONFIGURATION_SNAPSHOT JSONB",
		"CREATE TRIGGER TRIGGER_PREVENT_ORDER_ITEM_CONFIGURATION_SNAPSHOT_MUTATION",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}
	for _, required := range []string{
		"DROP TABLE IF EXISTS PRODUCT_CUSTOM_OPTION_POLICIES",
		"DROP COLUMN IF EXISTS CONFIGURATION_HASH",
		"DROP COLUMN IF EXISTS CONFIGURATION_SNAPSHOT",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
