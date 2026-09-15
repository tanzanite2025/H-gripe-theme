package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveLegacyProductSpecDefinitionFieldsMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "293_remove_legacy_product_spec_definition_fields.up.sql"))
	if !strings.Contains(up, "ALTER TABLE PRODUCT_SPEC_DEFINITIONS") {
		t.Fatal("up migration must alter product_spec_definitions")
	}
	for _, column := range []string{"DROP COLUMN IF EXISTS IS_VARIANT_OPTION", "DROP COLUMN IF EXISTS OPTIONS"} {
		if !strings.Contains(up, column) {
			t.Fatalf("up migration must contain %q", column)
		}
	}
	if strings.TrimSpace(readMigrationFile(t, "293_remove_legacy_product_spec_definition_fields.down.sql")) == "" {
		t.Fatal("down migration must document the intentional non-restoration")
	}
}
