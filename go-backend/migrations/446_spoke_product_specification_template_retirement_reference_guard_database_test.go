package migrations_test

import (
	"database/sql"
	"testing"
)

func TestSpokeProductSpecificationTemplateRetirementChecksReferencesAndBlocksRecreation(t *testing.T) {
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
INSERT INTO product_specification_templates VALUES (77, 'spoke'), (78, 'hub');
INSERT INTO product_spec_definitions VALUES (101, 77), (102, 78);
INSERT INTO product_spec_option_items VALUES (201, 101), (202, 102);
SET LOCAL search_path TO pg_temp;`)
	if err != nil {
		t.Fatal(err)
	}

	up := readMigrationFile(t, "446_retire_spoke_product_specification_template_and_block_recreation.up.sql")
	for _, reference := range []string{
		`INSERT INTO products(product_specification_template_id) VALUES (77)`,
		`INSERT INTO product_spec_values VALUES (101, 'CX-Ray')`,
		`INSERT INTO product_variant_option_values VALUES (101, NULL)`,
		`INSERT INTO product_variant_option_values VALUES (NULL, 201)`,
		`INSERT INTO product_option_group_variant_rules VALUES (101)`,
		`INSERT INTO quick_buy_step_filters VALUES (101)`,
		`INSERT INTO quick_buy_step_product_specification_templates VALUES (77)`,
	} {
		if _, err := tx.ExecContext(t.Context(), reference); err != nil {
			t.Fatalf("insert reference fixture %q: %v", reference, err)
		}
		expectSpokeProductSpecificationTemplateRetirementRejection(t, tx, up)
		if _, err := tx.ExecContext(t.Context(), `DELETE FROM products; DELETE FROM product_spec_values; DELETE FROM product_variant_option_values; DELETE FROM product_option_group_variant_rules; DELETE FROM quick_buy_step_filters; DELETE FROM quick_buy_step_product_specification_templates;`); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := tx.ExecContext(t.Context(), up); err != nil {
		t.Fatalf("retire unreferenced spoke template: %v", err)
	}
	for table, expectedCount := range map[string]int{
		"product_specification_templates": 1,
		"product_spec_definitions":        1,
		"product_spec_option_items":       1,
	} {
		var count int
		if err := tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != expectedCount {
			t.Fatalf("wrong retained rows in %s: count=%d err=%v", table, count, err)
		}
	}
	for _, slug := range []string{"spoke", "SPOKE", " spoke "} {
		expectSpokeProductSpecificationTemplateRetirementRejection(t, tx, "INSERT INTO product_specification_templates VALUES (79, '"+slug+"')")
	}
	expectSpokeProductSpecificationTemplateRetirementRejection(t, tx, `UPDATE product_specification_templates SET slug = 'spoke' WHERE id = 78`)

	if _, err := tx.ExecContext(t.Context(), readMigrationFile(t, "446_retire_spoke_product_specification_template_and_block_recreation.down.sql")); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), `INSERT INTO product_specification_templates VALUES (79, 'spoke')`); err != nil {
		t.Fatalf("rollback did not remove the retired slug guard: %v", err)
	}
}

func expectSpokeProductSpecificationTemplateRetirementRejection(t *testing.T, tx *sql.Tx, statement string) {
	t.Helper()
	const savepointName = "expected_spoke_template_retirement_rejection"
	if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), statement); err == nil {
		t.Fatal("expected spoke template migration or statement to reject the referenced or retired state")
	}
	if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
}
