package localbackup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBicycleTireRestoreTriggerContractsMatchTheirMigrations(t *testing.T) {
	for _, contract := range []struct {
		migration string
		function  string
		body      string
	}{
		{"439_add_dedicated_bicycle_tire_product_specification.up.sql", "require_bicycle_tire_product_type_and_tire_category", productCatalogBicycleTireProductTypeAndCategoryBindingFunctionBody},
		{"440_isolate_bicycle_tire_product_constraints_and_protect_category.up.sql", "protect_bicycle_tire_system_category_identity_and_tree_position", productCatalogBicycleTireSystemCategoryIdentityAndTreeProtectionFunctionBody},
	} {
		t.Run(contract.function, func(t *testing.T) {
			migrationSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", contract.migration))
			if err != nil {
				t.Fatalf("read %s: %v", contract.migration, err)
			}
			sqlText := strings.ReplaceAll(string(migrationSQL), "\r\n", "\n")
			functionStart := strings.Index(sqlText, "CREATE FUNCTION "+contract.function+"()")
			if functionStart < 0 {
				t.Fatalf("migration %s does not define %s", contract.migration, contract.function)
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
				t.Fatalf("restore trigger body for %s differs from migration %s", contract.function, contract.migration)
			}
		})
	}
}
