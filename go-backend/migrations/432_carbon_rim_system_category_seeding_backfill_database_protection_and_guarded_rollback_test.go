package migrations_test

import (
	"database/sql"
	"testing"
)

func createCarbonRimSystemCategoryMigrationTransaction(t *testing.T) *sql.Tx {
	t.Helper()
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	_, err := tx.ExecContext(t.Context(), `
DO $$
DECLARE isolated_schema TEXT := 'carbon_rim_category_test_' || pg_backend_pid();
BEGIN
    EXECUTE format('CREATE SCHEMA %I', isolated_schema);
    PERFORM set_config('search_path', isolated_schema || ', pg_temp', TRUE);
END $$;
CREATE TEMP TABLE product_categories (
    id BIGSERIAL PRIMARY KEY, parent_id BIGINT REFERENCES product_categories(id),
    name TEXT NOT NULL, slug TEXT UNIQUE NOT NULL, description TEXT NOT NULL DEFAULT '',
    depth INTEGER NOT NULL, sort_order INTEGER NOT NULL DEFAULT 0, is_enabled BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(), updated_at TIMESTAMPTZ DEFAULT NOW()
) ON COMMIT DROP;
CREATE TEMP TABLE product_category_translations (
    product_category_id BIGINT REFERENCES product_categories(id), locale TEXT, name TEXT, description TEXT,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, UNIQUE(product_category_id, locale)
) ON COMMIT DROP;
CREATE TEMP TABLE quick_buy_step_product_categories (product_category_id BIGINT REFERENCES product_categories(id)) ON COMMIT DROP;
ALTER TABLE products ADD COLUMN product_type TEXT NOT NULL DEFAULT '', ADD COLUMN product_category_id BIGINT REFERENCES product_categories(id),
    ADD COLUMN parent_id BIGINT, ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
INSERT INTO product_categories(id, name, slug, depth, is_enabled) VALUES (317, 'Wheel Components', 'wheel-components', 1, TRUE), (621, 'Other', 'other', 1, TRUE);
INSERT INTO products(id, product_type, product_category_id, parent_id, deleted_at) VALUES
    (21, 'carbon_rim', NULL, NULL, NULL), (22, 'carbon_rim', 621, 21, NULL), (23, 'carbon_rim', 621, NULL, NOW());
UPDATE products SET product_category_id = 621 WHERE id = 1;
`)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestCarbonRimSystemCategoryMigrationSeedsAndBackfillsWithoutFixedIDsAndProtectsDirectDatabaseWrites(t *testing.T) {
	tx := createCarbonRimSystemCategoryMigrationTransaction(t)
	up := readMigrationFile(t, "432_seed_and_protect_carbon_rim_system_category_and_require_dedicated_product_binding.up.sql")
	down := readMigrationFile(t, "432_seed_and_protect_carbon_rim_system_category_and_require_dedicated_product_binding.down.sql")
	if _, err := tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
	var matched int
	for statement, expected := range map[string]int{
		"SELECT COUNT(*) FROM product_categories WHERE slug = 'carbon-rim' AND parent_id = 317 AND depth = 2 AND is_enabled":                                                         1,
		"SELECT COUNT(*) FROM product_category_translations t JOIN product_categories c ON c.id = t.product_category_id WHERE c.slug = 'carbon-rim' AND t.locale IN ('en', 'zh_cn')": 2,
		"SELECT COUNT(*) FROM products p JOIN product_categories c ON c.id = p.product_category_id WHERE p.product_type = 'carbon_rim' AND c.slug = 'carbon-rim'":                    3,
		"SELECT COUNT(*) FROM products WHERE id = 1 AND product_type = '' AND product_category_id = 621":                                                                             1,
	} {
		if err := tx.QueryRowContext(t.Context(), statement).Scan(&matched); err != nil || matched != expected {
			t.Fatalf("unexpected migration result for %s: count=%d err=%v", statement, matched, err)
		}
	}
	for _, statement := range []string{
		"DELETE FROM product_categories WHERE slug = 'carbon-rim'",
		"DELETE FROM product_categories WHERE slug = 'wheel-components'",
		"UPDATE product_categories SET slug = 'renamed' WHERE slug = 'carbon-rim'",
		"UPDATE product_categories SET id = 888 WHERE slug = 'carbon-rim'",
		"UPDATE product_categories SET is_enabled = FALSE WHERE slug = 'carbon-rim'",
		"UPDATE product_categories SET parent_id = 621 WHERE slug = 'carbon-rim'",
		"UPDATE product_categories SET depth = 3 WHERE slug = 'carbon-rim'",
		"UPDATE product_categories SET parent_id = 621, depth = 2 WHERE slug = 'wheel-components'",
		"UPDATE product_categories SET is_enabled = FALSE WHERE slug = 'wheel-components'",
		"INSERT INTO products(product_type) VALUES ('carbon_rim')",
		"INSERT INTO products(product_type, product_category_id) VALUES ('carbon_rim', 621)",
		"INSERT INTO products(product_category_id) SELECT id FROM product_categories WHERE slug = 'carbon-rim'",
		"UPDATE products SET product_category_id = NULL WHERE id = 21",
		"UPDATE products SET product_type = '', product_category_id = 621 WHERE id = 21",
		"UPDATE products SET product_type = 'carbon_rim', product_category_id = (SELECT id FROM product_categories WHERE slug = 'carbon-rim') WHERE id = 1",
		down,
	} {
		expectCarbonRimMigrationStatementRejection(t, tx, statement)
	}
	if _, err := tx.ExecContext(t.Context(), "UPDATE product_categories SET name = 'Carbon Fiber Rims', description = 'Editable copy' WHERE slug = 'carbon-rim'; UPDATE products SET updated_at = NOW() WHERE id = 21; DELETE FROM products WHERE product_type = 'carbon_rim'; INSERT INTO quick_buy_step_product_categories SELECT id FROM product_categories WHERE slug = 'carbon-rim'"); err != nil {
		t.Fatal(err)
	}
	expectCarbonRimMigrationStatementRejection(t, tx, down)
	if _, err := tx.ExecContext(t.Context(), "DELETE FROM quick_buy_step_product_categories"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
}

func TestCarbonRimSystemCategoryMigrationPreservesExistingCategoryIDAndTranslation(t *testing.T) {
	tx := createCarbonRimSystemCategoryMigrationTransaction(t)
	if _, err := tx.ExecContext(t.Context(), `INSERT INTO product_categories(id, name, slug, parent_id, depth, is_enabled) VALUES (975, 'Custom rim label', 'carbon-rim', 621, 2, FALSE); INSERT INTO product_category_translations(product_category_id, locale, name) VALUES (975, 'zh_cn', '自定义碳圈名称')`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), readMigrationFile(t, "432_seed_and_protect_carbon_rim_system_category_and_require_dedicated_product_binding.up.sql")); err != nil {
		t.Fatal(err)
	}
	var preserved bool
	if err := tx.QueryRowContext(t.Context(), `SELECT c.id = 975 AND c.name = 'Custom rim label' AND c.parent_id = 317 AND c.is_enabled AND t.name = '自定义碳圈名称' FROM product_categories c JOIN product_category_translations t ON t.product_category_id = c.id AND t.locale = 'zh_cn' WHERE c.slug = 'carbon-rim'`).Scan(&preserved); err != nil || !preserved {
		t.Fatalf("existing category identity or translation was lost: %v", err)
	}
}

func TestCarbonRimSystemCategoryMigrationRejectsCategoryWithGenericProductsAndMissingEnabledRoot(t *testing.T) {
	for _, setup := range []string{
		"UPDATE product_categories SET is_enabled = FALSE WHERE slug = 'wheel-components'",
		"INSERT INTO product_categories(id, name, slug, depth, is_enabled) VALUES (975, 'Carbon Rim', 'carbon-rim', 1, TRUE); UPDATE products SET product_category_id = 975 WHERE id = 1",
	} {
		t.Run(setup, func(t *testing.T) {
			tx := createCarbonRimSystemCategoryMigrationTransaction(t)
			if _, err := tx.ExecContext(t.Context(), setup); err != nil {
				t.Fatal(err)
			}
			expectCarbonRimMigrationStatementRejection(t, tx, readMigrationFile(t, "432_seed_and_protect_carbon_rim_system_category_and_require_dedicated_product_binding.up.sql"))
		})
	}
}
