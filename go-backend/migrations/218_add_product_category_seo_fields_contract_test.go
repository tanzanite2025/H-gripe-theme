package migrations_test

import (
	"strings"
	"testing"
)

func TestProductCategorySEOMigrationAddsOwnedFields(t *testing.T) {
	upSQL := readMigrationFile(t, "218_add_product_category_seo_fields.up.sql")
	downSQL := readMigrationFile(t, "218_add_product_category_seo_fields.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE product_categories",
		"meta_title",
		"meta_description",
		"seo_intro",
		"ALTER TABLE product_category_translations",
	} {
		if !strings.Contains(strings.ToLower(upSQL), strings.ToLower(fragment)) {
			t.Fatalf("category SEO migration is missing contract fragment %q", fragment)
		}
	}
	for _, fragment := range []string{
		"DROP COLUMN IF EXISTS seo_intro",
		"DROP COLUMN IF EXISTS meta_description",
		"DROP COLUMN IF EXISTS meta_title",
	} {
		if !strings.Contains(strings.ToLower(downSQL), strings.ToLower(fragment)) {
			t.Fatalf("category SEO down migration is missing contract fragment %q", fragment)
		}
	}
}
