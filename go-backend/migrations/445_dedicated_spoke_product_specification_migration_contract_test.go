package migrations_test

import (
	"strings"
	"testing"
)

func TestDedicatedSpokeProductMigrationKeepsTypeCategoryAndRollbackBoundaries(t *testing.T) {
	up := readMigrationFile(t, "445_add_dedicated_spoke_product_specification.up.sql")
	down := readMigrationFile(t, "445_add_dedicated_spoke_product_specification.down.sql")

	for _, contract := range []string{
		"ADD COLUMN spoke_specification JSONB",
		"WHERE slug = 'spokes' AND is_enabled",
		"product_type IN ('', 'carbon_rim', 'bicycle_tire', 'wheelset', 'spoke')",
		"product_type = 'spoke'",
		"spoke_specification IS NOT NULL",
		"product_specification_template_id IS NULL",
		"NULLIF(BTRIM(spoke_specification->>'model_name'), '') IS NOT NULL",
		"spokes is reserved for dedicated spoke products",
		"spoke products must belong to the enabled spokes system category",
		"Dedicated spoke product type is immutable",
		"require_spoke_product_type_and_system_category_binding",
	} {
		if !strings.Contains(up, contract) {
			t.Errorf("up migration is missing contract %q", contract)
		}
	}
	if strings.Index(up, "The enabled spokes system category is required") > strings.Index(up, "CREATE TRIGGER require_spoke_product_type_and_system_category_binding") {
		t.Fatal("spoke category must be verified before allowing the dedicated product binding")
	}

	for _, contract := range []string{
		"Cannot remove spoke product support while spoke products or specifications exist",
		"DROP COLUMN spoke_specification",
		"product_type IN ('', 'carbon_rim', 'bicycle_tire', 'wheelset')",
		"reject_product_binding_to_unimplemented_spokes_category",
		"spokes products require the dedicated spoke product type, which is not implemented yet",
	} {
		if !strings.Contains(down, contract) {
			t.Errorf("down migration is missing rollback contract %q", contract)
		}
	}
	if strings.Index(down, "Cannot remove spoke product support") > strings.Index(down, "DROP COLUMN spoke_specification") {
		t.Fatal("rollback must reject existing spoke products before dropping their specification data")
	}
}
