package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsolateBicycleTireProductConstraintsAndProtectCategoryContract(t *testing.T) {
	upSQL, err := os.ReadFile(filepath.Join("440_isolate_bicycle_tire_product_constraints_and_protect_category.up.sql"))
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile(filepath.Join("440_isolate_bicycle_tire_product_constraints_and_protect_category.down.sql"))
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	down := strings.ToLower(string(downSQL))
	carbonConstraintStart := strings.Index(up, "add constraint products_carbon_rim_type_and_specification_check")
	tireConstraintStart := strings.Index(up, "add constraint products_bicycle_tire_type_and_specification_check")
	categoryTriggerStart := strings.Index(up, "create function protect_bicycle_tire_system_category_identity_and_tree_position")
	if carbonConstraintStart < 0 || tireConstraintStart <= carbonConstraintStart || categoryTriggerStart <= tireConstraintStart {
		t.Fatal("up migration must define separate carbon-rim and bicycle-tire checks before protecting the tire category")
	}
	carbonConstraint := up[carbonConstraintStart:tireConstraintStart]
	tireConstraint := up[tireConstraintStart:categoryTriggerStart]
	if strings.Contains(carbonConstraint, "bicycle_tire_specification") {
		t.Fatal("carbon-rim database constraint must not inspect bicycle-tire specification fields")
	}
	if strings.Contains(tireConstraint, "carbon_rim_specification") {
		t.Fatal("bicycle-tire database constraint must not inspect carbon-rim specification fields")
	}
	for _, required := range []string{
		"products_dedicated_product_type_allowed_check",
		"product_type in ('', 'carbon_rim', 'bicycle_tire')",
		"protect_bicycle_tire_system_category_identity_and_tree_position",
		"tire must remain enabled directly under wheel-components",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}
	if !strings.Contains(down, "products_dedicated_product_type_specifications_check") || strings.Contains(down, "drop column bicycle_tire_specification") {
		t.Fatal("down migration must restore the version-439 checks without deleting tire product data")
	}
}
