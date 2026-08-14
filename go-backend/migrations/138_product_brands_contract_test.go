package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readProductBrandMigration(t *testing.T, name string) string {
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

func TestProductBrandUpMigrationContract(t *testing.T) {
	sql := readProductBrandMigration(t, "138_product_brands.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS product_brands",
		"slug VARCHAR(160) NOT NULL",
		"ALTER TABLE products",
		"ADD COLUMN IF NOT EXISTS brand_id BIGINT NULL",
		"fk_products_brand",
		"ON DELETE RESTRICT",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestProductBrandDownMigrationContract(t *testing.T) {
	sql := readProductBrandMigration(t, "138_product_brands.down.sql")
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS fk_products_brand",
		"DROP COLUMN IF EXISTS brand_id",
		"DROP TABLE IF EXISTS product_brands",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
