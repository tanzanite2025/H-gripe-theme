package migrations_test

import "testing"

func TestRetiredCarbonRimTemplateMigrationPhysicallyDeletesUnusedRowsAndBlocksRecreation(t *testing.T) {
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	_, err := tx.Exec(`CREATE TEMP TABLE product_specification_templates(id BIGINT PRIMARY KEY, slug TEXT);
 CREATE TEMP TABLE product_spec_definitions(id BIGINT PRIMARY KEY, product_specification_template_id BIGINT);
 CREATE TEMP TABLE product_spec_option_items(id BIGINT PRIMARY KEY,spec_definition_id BIGINT);
 CREATE TEMP TABLE product_spec_values(spec_definition_id BIGINT);
 CREATE TEMP TABLE product_variant_option_values(spec_definition_id BIGINT);
 CREATE TEMP TABLE product_option_group_variant_rules(spec_definition_id BIGINT);
 CREATE TEMP TABLE quick_buy_step_filters(spec_definition_id BIGINT);
 CREATE TEMP TABLE quick_buy_step_product_specification_templates(product_specification_template_id BIGINT);
 INSERT INTO product_specification_templates VALUES(77,'rim'),(78,'wheelset');
 INSERT INTO product_spec_definitions VALUES(101,77),(102,78);
 INSERT INTO product_spec_option_items VALUES(201,101),(202,102);`)
	if err != nil {
		t.Fatal(err)
	}
	up := readMigrationFile(t, "430_physically_delete_unused_legacy_carbon_rim_templates_and_block_recreation.up.sql")
	expectCarbonRimMigrationStatementRejection(t, tx, up)
	if _, err = tx.Exec(`UPDATE products SET product_specification_template_id=78 WHERE id=1; INSERT INTO product_variant_option_values VALUES(101)`); err != nil {
		t.Fatal(err)
	}
	expectCarbonRimMigrationStatementRejection(t, tx, up)
	if _, err = tx.Exec(`DELETE FROM product_variant_option_values`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(up); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"product_specification_templates", "product_spec_definitions", "product_spec_option_items"} {
		var count int
		if err = tx.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("wrong retained rows in %s: %d %v", table, count, err)
		}
	}
	for _, slug := range []string{"rim", "carbon_rim", "CARBON-RIM"} {
		expectCarbonRimMigrationStatementRejection(t, tx, "INSERT INTO product_specification_templates VALUES(79,'"+slug+"')")
	}
	expectCarbonRimMigrationStatementRejection(t, tx, `UPDATE product_specification_templates SET slug='rim' WHERE id=78`)
	if _, err = tx.Exec(readMigrationFile(t, "430_physically_delete_unused_legacy_carbon_rim_templates_and_block_recreation.down.sql")); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = tx.QueryRow(`SELECT count(*) FROM product_specification_templates WHERE slug='rim'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback resurrected removed template: %v", err)
	}
}
