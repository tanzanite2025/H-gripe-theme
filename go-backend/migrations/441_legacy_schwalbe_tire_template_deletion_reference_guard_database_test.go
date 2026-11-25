package migrations_test

import (
	"database/sql"
	"testing"
)

func TestLegacySchwalbeTireTemplateDeletionGuardsReferencesAndBlocksRecreation(t *testing.T) {
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	_, err := tx.ExecContext(t.Context(), `
CREATE TEMP TABLE product_specification_templates(id BIGINT PRIMARY KEY, slug TEXT);
CREATE TEMP TABLE product_spec_definitions(id BIGINT PRIMARY KEY, product_specification_template_id BIGINT);
CREATE TEMP TABLE product_spec_option_items(id BIGINT PRIMARY KEY, spec_definition_id BIGINT);
CREATE TEMP TABLE product_spec_values(spec_definition_id BIGINT, value TEXT);
CREATE TEMP TABLE product_variant_option_values(spec_definition_id BIGINT, template_option_item_id BIGINT);
CREATE TEMP TABLE product_option_group_variant_rules(spec_definition_id BIGINT);
CREATE TEMP TABLE quick_buy_step_filters(spec_definition_id BIGINT);
CREATE TEMP TABLE quick_buy_step_product_specification_templates(product_specification_template_id BIGINT);
INSERT INTO product_specification_templates VALUES (77, 'schwalbe_tire'), (78, 'wheelset');
INSERT INTO product_spec_definitions VALUES (101, 77), (102, 78);
INSERT INTO product_spec_option_items VALUES (201, 101), (202, 102);
CREATE UNIQUE INDEX uq_schwalbe_product_article_no ON product_spec_values (lower(btrim(value))) WHERE btrim(value) <> '';
SET LOCAL search_path TO pg_temp;`)
	if err != nil {
		t.Fatal(err)
	}

	up := readMigrationFile(t, "441_physically_delete_unused_legacy_schwalbe_tire_template_and_block_recreation.up.sql")
	expectLegacySchwalbeTireTemplateMigrationStatementRejection(t, tx, up)

	if _, err = tx.ExecContext(t.Context(), `UPDATE products SET product_specification_template_id = NULL WHERE id = 1;
INSERT INTO product_spec_values VALUES (101, 'LEGACY-ARTICLE');
INSERT INTO product_variant_option_values VALUES (101);
INSERT INTO product_option_group_variant_rules VALUES (101);
	INSERT INTO quick_buy_step_filters VALUES (101);
	INSERT INTO quick_buy_step_product_specification_templates VALUES (77);`); err != nil {
		t.Fatal(err)
	}
	expectLegacySchwalbeTireTemplateMigrationStatementRejection(t, tx, up)

	if _, err = tx.ExecContext(t.Context(), `DELETE FROM product_spec_values;
DELETE FROM product_variant_option_values;
DELETE FROM product_option_group_variant_rules;
DELETE FROM quick_buy_step_filters;
DELETE FROM quick_buy_step_product_specification_templates;`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(t.Context(), `INSERT INTO product_variant_option_values VALUES (102, 201)`); err != nil {
		t.Fatal(err)
	}
	expectLegacySchwalbeTireTemplateMigrationStatementRejection(t, tx, up)
	if _, err = tx.ExecContext(t.Context(), `DELETE FROM product_variant_option_values`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}

	for table, expectedCount := range map[string]int{"product_specification_templates": 1, "product_spec_definitions": 1, "product_spec_option_items": 1} {
		var count int
		if err = tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != expectedCount {
			t.Fatalf("wrong retained rows in %s: %d %v", table, count, err)
		}
	}
	var oldIndexExists bool
	if err = tx.QueryRowContext(t.Context(), `SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE schemaname = current_schema() AND indexname = 'uq_schwalbe_product_article_no')`).Scan(&oldIndexExists); err != nil || oldIndexExists {
		t.Fatalf("old Article No. index remains: exists=%t err=%v", oldIndexExists, err)
	}
	for _, slug := range []string{"schwalbe_tire", "SCHWALBE_TIRE", " schwalbe_tire "} {
		expectLegacySchwalbeTireTemplateMigrationStatementRejection(t, tx, "INSERT INTO product_specification_templates VALUES (79, '"+slug+"')")
	}
	expectLegacySchwalbeTireTemplateMigrationStatementRejection(t, tx, `UPDATE product_specification_templates SET slug = 'schwalbe_tire' WHERE id = 78`)

	if _, err = tx.ExecContext(t.Context(), readMigrationFile(t, "441_physically_delete_unused_legacy_schwalbe_tire_template_and_block_recreation.down.sql")); err != nil {
		t.Fatal(err)
	}
	var deletedTemplateCount int
	if err = tx.QueryRowContext(t.Context(), `SELECT count(*) FROM product_specification_templates WHERE lower(btrim(slug)) = 'schwalbe_tire'`).Scan(&deletedTemplateCount); err != nil || deletedTemplateCount != 0 {
		t.Fatalf("rollback unexpectedly restored the deleted template: count=%d err=%v", deletedTemplateCount, err)
	}
}

func expectLegacySchwalbeTireTemplateMigrationStatementRejection(t *testing.T, tx *sql.Tx, statement string) {
	t.Helper()
	const savepointName = "expected_legacy_schwalbe_template_rejection"
	if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), statement); err == nil {
		t.Fatal("expected statement to reject an unsafe legacy Schwalbe template state")
	}
	if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
}
