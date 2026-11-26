package localbackup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpokesSystemCategoryRestoreTriggerContractsMatchMigration444(t *testing.T) {
	migrationSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "444_seed_and_protect_spokes_product_category.up.sql"))
	if err != nil {
		t.Fatalf("read spokes category migration: %v", err)
	}
	sqlText := strings.ReplaceAll(string(migrationSQL), "\r\n", "\n")
	for _, contract := range []struct {
		function string
		body     string
	}{
		{"protect_spokes_system_category_identity_and_tree_position", productCatalogSpokesSystemCategoryIdentityAndTreeProtectionFunctionBody},
		{"reject_product_binding_to_unimplemented_spokes_category", productCatalogRejectProductBindingToUnimplementedSpokesCategoryFunctionBody},
	} {
		t.Run(contract.function, func(t *testing.T) {
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
				t.Fatalf("restore trigger body for %s differs from migration 444", contract.function)
			}
		})
	}
	for _, required := range []string{
		"'spokes'",
		"'辐条'",
		"protect_spokes_system_category_identity_and_tree_position",
		"reject_product_binding_to_unimplemented_spokes_category",
		"wheel-components",
	} {
		if !strings.Contains(sqlText, required) {
			t.Errorf("spokes migration is missing %q", required)
		}
	}
}
