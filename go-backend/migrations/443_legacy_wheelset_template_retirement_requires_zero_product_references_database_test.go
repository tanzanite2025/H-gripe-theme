package migrations_test

import (
	"database/sql"
	"testing"
)

func TestLegacyWheelsetTemplateRetirementRequiresZeroProductReferencesAndBlocksRecreation(t *testing.T) {
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	_, err := tx.ExecContext(t.Context(), `
CREATE TEMP TABLE product_specification_templates(id BIGINT PRIMARY KEY, slug TEXT);
CREATE TEMP TABLE product_spec_definitions(id BIGINT PRIMARY KEY, product_specification_template_id BIGINT);
CREATE TEMP TABLE product_spec_option_items(id BIGINT PRIMARY KEY, spec_definition_id BIGINT);
CREATE TEMP TABLE product_spec_values(spec_definition_id BIGINT, value TEXT);
CREATE TEMP TABLE product_variant_option_values(id BIGINT PRIMARY KEY, spec_definition_id BIGINT, template_option_item_id BIGINT);
CREATE TEMP TABLE product_option_group_variant_rules(spec_definition_id BIGINT);
CREATE TEMP TABLE product_option_value_variant_rules(product_variant_option_value_id BIGINT);
CREATE TEMP TABLE product_option_value_relations(source_option_value_id BIGINT, target_option_value_id BIGINT);
CREATE TEMP TABLE product_custom_option_policies(product_variant_option_value_id BIGINT);
CREATE TEMP TABLE quick_buy_step_filters(spec_definition_id BIGINT);
CREATE TEMP TABLE quick_buy_step_product_specification_templates(product_specification_template_id BIGINT);
INSERT INTO product_specification_templates VALUES (77, 'wheelset'), (78, 'handlebar');
INSERT INTO product_spec_definitions VALUES (101, 77), (102, 78);
INSERT INTO product_spec_option_items VALUES (201, 101), (202, 102);
SET LOCAL search_path TO pg_temp;`)
	if err != nil {
		t.Fatal(err)
	}

	up := readMigrationFile(t, "443_retire_legacy_wheelset_product_specification_template.up.sql")
	expectWheelsetTemplateMigrationStatementRejection(t, tx, up)
	if _, err = tx.ExecContext(t.Context(), `UPDATE products SET deleted_at = NOW() WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	expectWheelsetTemplateMigrationStatementRejection(t, tx, up)
	if _, err = tx.ExecContext(t.Context(), `DELETE FROM products WHERE id = 1`); err != nil {
		t.Fatal(err)
	}

	referenceCases := []struct {
		name   string
		insert string
		delete string
	}{
		{"product specification value", `INSERT INTO product_spec_values VALUES (101, '29')`, `DELETE FROM product_spec_values`},
		{"materialized variant option", `INSERT INTO product_variant_option_values VALUES (1, 101, NULL)`, `DELETE FROM product_variant_option_values`},
		{"variant-specific option rule", `INSERT INTO product_variant_option_values VALUES (301, 101, NULL); INSERT INTO product_option_value_variant_rules VALUES (301)`, `DELETE FROM product_option_value_variant_rules; DELETE FROM product_variant_option_values`},
		{"option value relation", `INSERT INTO product_variant_option_values VALUES (301, 101, NULL); INSERT INTO product_option_value_relations VALUES (301, 302)`, `DELETE FROM product_option_value_relations; DELETE FROM product_variant_option_values`},
		{"custom option policy", `INSERT INTO product_variant_option_values VALUES (301, 101, NULL); INSERT INTO product_custom_option_policies VALUES (301)`, `DELETE FROM product_custom_option_policies; DELETE FROM product_variant_option_values`},
		{"variant group rule", `INSERT INTO product_option_group_variant_rules VALUES (101)`, `DELETE FROM product_option_group_variant_rules`},
		{"quick buy filter", `INSERT INTO quick_buy_step_filters VALUES (101)`, `DELETE FROM quick_buy_step_filters`},
		{"quick buy template binding", `INSERT INTO quick_buy_step_product_specification_templates VALUES (77)`, `DELETE FROM quick_buy_step_product_specification_templates`},
	}
	for _, referenceCase := range referenceCases {
		t.Run(referenceCase.name, func(t *testing.T) {
			if _, err := tx.ExecContext(t.Context(), referenceCase.insert); err != nil {
				t.Fatal(err)
			}
			expectWheelsetTemplateMigrationStatementRejection(t, tx, up)
			if _, err := tx.ExecContext(t.Context(), referenceCase.delete); err != nil {
				t.Fatal(err)
			}
		})
	}

	if _, err = tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
	for query, expected := range map[string]int{
		`SELECT count(*) FROM products`:                                              0,
		`SELECT count(*) FROM product_specification_templates`:                       1,
		`SELECT count(*) FROM product_spec_definitions`:                              1,
		`SELECT count(*) FROM product_spec_option_items`:                             1,
		`SELECT count(*) FROM product_specification_templates WHERE slug='wheelset'`: 0,
	} {
		var actual int
		if err = tx.QueryRowContext(t.Context(), query).Scan(&actual); err != nil || actual != expected {
			t.Fatalf("unexpected row count for %q: got %d want %d err=%v", query, actual, expected, err)
		}
	}

	for _, slug := range []string{"wheelset", "WHEELSET", " wheelset "} {
		expectWheelsetTemplateMigrationStatementRejection(t, tx, `INSERT INTO product_specification_templates VALUES (79, '`+slug+`')`)
	}
	expectWheelsetTemplateMigrationStatementRejection(t, tx, `UPDATE product_specification_templates SET slug = 'wheelset' WHERE id = 78`)

	if _, err = tx.ExecContext(t.Context(), readMigrationFile(t, "443_retire_legacy_wheelset_product_specification_template.down.sql")); err != nil {
		t.Fatal(err)
	}
	var recreatedTemplateCount int
	if err = tx.QueryRowContext(t.Context(), `SELECT count(*) FROM product_specification_templates WHERE lower(btrim(slug)) = 'wheelset'`).Scan(&recreatedTemplateCount); err != nil || recreatedTemplateCount != 0 {
		t.Fatalf("rollback unexpectedly restored the removed template: count=%d err=%v", recreatedTemplateCount, err)
	}
}

func expectWheelsetTemplateMigrationStatementRejection(t *testing.T, tx *sql.Tx, statement string) {
	t.Helper()
	const savepointName = "expected_wheelset_template_retirement_rejection"
	if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), statement); err == nil {
		t.Fatal("expected statement to reject a legacy wheelset product or template reference")
	}
	if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
}
