package migrations_test

import (
	"strings"
	"testing"
)

func TestProductCategoryProductForeignKeyRestrictMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "264_product_category_product_fk_restrict.up.sql")
	downSQL := readMigrationFile(t, "264_product_category_product_fk_restrict.down.sql")

	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS fk_products_product_category",
		"ADD CONSTRAINT fk_products_product_category",
		"REFERENCES product_categories(id)",
		"ON DELETE RESTRICT",
	} {
		if !strings.Contains(strings.ToUpper(upSQL), strings.ToUpper(fragment)) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(strings.ToUpper(downSQL), "ON DELETE SET NULL") {
		t.Fatal("down migration must restore the previous nullable delete behavior")
	}
}
