package migrations_test

import (
	"strings"
	"testing"
)

func TestOptionalGenericProductSchemaMigrationDefinesOrdinaryProductBoundary(t *testing.T) {
	up := readMigrationFile(t, "448_add_optional_generic_product_schema.up.sql")
	for _, required := range []string{
		"ADD COLUMN generic_product_schema JSONB",
		"products_generic_product_schema_check",
		"product_type = ''",
		"(generic_product_schema->>'schema_version') IS NOT DISTINCT FROM '1'",
		"product_type <> ''",
		"generic_product_schema IS NULL",
	} {
		if !strings.Contains(up, required) {
			t.Errorf("migration 448 is missing generic schema contract %q", required)
		}
	}
}

func TestOptionalGenericProductSchemaMigrationDownRemovesOnlyGenericSchema(t *testing.T) {
	down := readMigrationFile(t, "448_add_optional_generic_product_schema.down.sql")
	if !strings.Contains(down, "generic_product_schema IS NOT NULL") || !strings.Contains(down, "RAISE EXCEPTION") || !strings.Contains(down, "DROP COLUMN IF EXISTS generic_product_schema") || !strings.Contains(down, "DROP CONSTRAINT IF EXISTS products_generic_product_schema_check") {
		t.Fatal("migration 448 down script must remove the generic product schema column and constraint")
	}
}
