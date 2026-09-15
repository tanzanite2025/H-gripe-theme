package migrations_test

import (
	"strings"
	"testing"
)

func TestProductOptionVariantRulesMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "287_product_option_variant_rules.up.sql"))
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS PRODUCT_OPTION_GROUP_VARIANT_RULES",
		"UNIQUE (VARIANT_ID, SPEC_DEFINITION_ID)",
		"CREATE TABLE IF NOT EXISTS PRODUCT_OPTION_VALUE_VARIANT_RULES",
		"UNIQUE (VARIANT_ID, PRODUCT_VARIANT_OPTION_VALUE_ID)",
		"PRICE_DELTA_MINOR_OVERRIDE",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("migration must contain %q", required)
		}
	}
}
