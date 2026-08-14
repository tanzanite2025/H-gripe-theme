package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readProductVariantJSONBMigration(t *testing.T, name string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test location")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(data)
}

func TestProductVariantOptionValuesJSONBUpMigrationContract(t *testing.T) {
	sql := readProductVariantJSONBMigration(t, "131_product_variant_option_values_jsonb.up.sql")

	for _, fragment := range []string{
		"ALTER COLUMN option_values TYPE jsonb USING option_values::jsonb",
		"chk_product_variants_option_values_object",
		"USING GIN (option_values jsonb_ops)",
		"WHERE deleted_at IS NULL",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestProductVariantOptionValuesJSONBDownMigrationContract(t *testing.T) {
	sql := readProductVariantJSONBMigration(t, "131_product_variant_option_values_jsonb.down.sql")

	for _, fragment := range []string{
		"ALTER COLUMN option_values TYPE TEXT USING option_values::text",
		"DROP CONSTRAINT IF EXISTS chk_product_variants_option_values_object",
		"USING GIN ((option_values::jsonb) jsonb_ops)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
