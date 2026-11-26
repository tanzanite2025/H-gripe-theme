package migrations_test

import (
	"context"
	"database/sql"
	"testing"
)

func createSpokesProductCategoryMigrationTransaction(t *testing.T, rootEnabled bool) *sql.Tx {
	t.Helper()
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	rootState := "FALSE"
	if rootEnabled {
		rootState = "TRUE"
	}
	_, err := tx.ExecContext(t.Context(), `
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
ALTER TABLE products ADD COLUMN product_type TEXT NOT NULL DEFAULT '', ADD COLUMN product_category_id BIGINT;
INSERT INTO product_categories(id, name, slug, depth, is_enabled) VALUES (317, 'Wheel Components', 'wheel-components', 1, `+rootState+`);
SET LOCAL search_path TO pg_temp;
`)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestSpokesProductCategoryMigrationSeedsTranslationsAndProtectsItsFutureProductBoundary(t *testing.T) {
	tx := createSpokesProductCategoryMigrationTransaction(t, true)

	up := readMigrationFile(t, "444_seed_and_protect_spokes_product_category.up.sql")
	down := readMigrationFile(t, "444_seed_and_protect_spokes_product_category.down.sql")
	if _, err := tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
	var matched int
	for statement, expected := range map[string]int{
		"SELECT COUNT(*) FROM product_categories WHERE slug='spokes' AND parent_id=317 AND depth=2 AND sort_order=25 AND is_enabled":                                                                                           1,
		"SELECT COUNT(*) FROM product_category_translations translation JOIN product_categories category ON category.id=translation.product_category_id WHERE category.slug='spokes' AND translation.locale IN ('en','zh_cn')": 2,
		"SELECT COUNT(*) FROM products WHERE product_category_id=(SELECT id FROM product_categories WHERE slug='spokes')":                                                                                                      0,
	} {
		if err := tx.QueryRowContext(t.Context(), statement).Scan(&matched); err != nil || matched != expected {
			t.Fatalf("unexpected spokes migration result for %s: count=%d err=%v", statement, matched, err)
		}
	}
	for _, statement := range []string{
		"DELETE FROM product_categories WHERE slug='spokes'",
		"UPDATE product_categories SET slug='spokes-renamed' WHERE slug='spokes'",
		"UPDATE product_categories SET id=888 WHERE slug='spokes'",
		"UPDATE product_categories SET is_enabled=FALSE WHERE slug='spokes'",
		"UPDATE product_categories SET parent_id=NULL WHERE slug='spokes'",
		"UPDATE product_categories SET depth=3 WHERE slug='spokes'",
		"INSERT INTO products(product_type, product_category_id) SELECT 'spoke', id FROM product_categories WHERE slug='spokes'",
		"UPDATE products SET product_category_id=(SELECT id FROM product_categories WHERE slug='spokes') WHERE id=1",
	} {
		expectSpokesCategoryMigrationStatementRejection(t, tx, statement)
	}

	if _, err := tx.ExecContext(t.Context(), "INSERT INTO quick_buy_step_product_categories SELECT id FROM product_categories WHERE slug='spokes'"); err != nil {
		t.Fatal(err)
	}
	expectSpokesCategoryMigrationStatementRejection(t, tx, down)
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

func TestSpokesProductCategoryMigrationRequiresTheEnabledWheelComponentsRoot(t *testing.T) {
	tx := createSpokesProductCategoryMigrationTransaction(t, false)
	expectSpokesCategoryMigrationStatementRejection(t, tx, readMigrationFile(t, "444_seed_and_protect_spokes_product_category.up.sql"))
}

func expectSpokesCategoryMigrationStatementRejection(t *testing.T, tx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, statement string) {
	t.Helper()
	const savepointName = "expected_spokes_category_migration_rejection"
	if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), statement); err == nil {
		t.Fatal("expected spokes category migration contract to reject the statement")
	}
	if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
}
