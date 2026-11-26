package localbackup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWheelsetRestoreTriggerContractsMatchTheirMigrationBodies(t *testing.T) {
	for _, contract := range []struct {
		function string
		body     string
	}{
		{"require_wheelset_product_type_and_system_category_binding", productCatalogWheelsetProductTypeAndCategoryBindingFunctionBody},
		{"protect_wheelset_system_category_identity", productCatalogWheelsetSystemCategoryIdentityProtectionFunctionBody},
	} {
		t.Run(contract.function, func(t *testing.T) {
			migrationSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "442_add_dedicated_wheelset_product_specification.up.sql"))
			if err != nil {
				t.Fatalf("read WHEELSET migration: %v", err)
			}
			sqlText := strings.ReplaceAll(string(migrationSQL), "\r\n", "\n")
			functionStart := strings.Index(sqlText, "CREATE FUNCTION "+contract.function+"()")
			if functionStart < 0 {
				t.Fatalf("migration does not define %s", contract.function)
			}
			bodyStartOffset := strings.Index(sqlText[functionStart:], "AS $$")
			if bodyStartOffset < 0 {
				t.Fatalf("migration function %s has no dollar-quoted body", contract.function)
			}
			bodyStart := functionStart + bodyStartOffset + len("AS $$")
			bodyEndOffset := strings.Index(sqlText[bodyStart:], "$$;")
			if bodyEndOffset < 0 {
				t.Fatalf("migration function %s has no closing dollar quote", contract.function)
			}
			actualBody := strings.TrimSpace(sqlText[bodyStart : bodyStart+bodyEndOffset])
			if actualBody != contract.body {
				t.Fatalf("restore trigger body for %s differs from the migration", contract.function)
			}
		})
	}
}

func TestWheelsetMigrationDefinesIndependentVersionedProductSpecification(t *testing.T) {
	upSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "442_add_dedicated_wheelset_product_specification.up.sql"))
	if err != nil {
		t.Fatalf("read WHEELSET up migration: %v", err)
	}
	downSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "442_add_dedicated_wheelset_product_specification.down.sql"))
	if err != nil {
		t.Fatalf("read WHEELSET down migration: %v", err)
	}
	up, down := strings.ToLower(string(upSQL)), strings.ToLower(string(downSQL))
	for _, required := range []string{
		"add column wheelset_specification jsonb",
		"product_type in ('', 'carbon_rim', 'bicycle_tire', 'wheelset')",
		"products_wheelset_type_and_specification_check",
		"front_wheel",
		"rear_wheel",
		"spoke_hole_counts",
		"lacing_patterns",
		"require_wheelset_product_type_and_system_category_binding",
		"protect_wheelset_system_category_identity",
	} {
		if !strings.Contains(up, required) {
			t.Errorf("up migration missing %q", required)
		}
	}
	if strings.Contains(up, "wheelset.depth") || strings.Contains(up, "wheelset.parent_id") {
		t.Fatal("WHEELSET category binding must not assume a fixed root position")
	}
	if !strings.Contains(down, "cannot remove wheelset product support while wheelset products exist") || !strings.Contains(down, "drop column wheelset_specification") {
		t.Fatal("down migration must refuse data loss before removing the WHEELSET field")
	}
}
