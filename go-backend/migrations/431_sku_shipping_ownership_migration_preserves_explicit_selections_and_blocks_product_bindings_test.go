package migrations_test

import "testing"

func TestSkuShippingMigrationCopiesOnlyInheritedSelectionsAndPermanentlyRejectsProductScope(t *testing.T) {
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	_, err := tx.ExecContext(t.Context(), `
 ALTER TABLE product_variants ADD COLUMN product_id BIGINT, ADD COLUMN deleted_at TIMESTAMPTZ;
 CREATE TEMP TABLE product_shipping_template_bindings (
 id BIGSERIAL PRIMARY KEY, product_id BIGINT NOT NULL, variant_id BIGINT,
 shipping_template_id BIGINT NOT NULL, priority INTEGER NOT NULL DEFAULT 0,
 enabled BOOLEAN NOT NULL DEFAULT TRUE, deleted_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ DEFAULT NOW(), updated_at TIMESTAMPTZ DEFAULT NOW());
 CREATE UNIQUE INDEX idx_product_shipping_template_bindings_product_scope
 ON product_shipping_template_bindings(product_id,shipping_template_id) WHERE variant_id IS NULL AND deleted_at IS NULL;
 CREATE UNIQUE INDEX idx_product_shipping_template_bindings_variant_scope
 ON product_shipping_template_bindings(product_id,variant_id,shipping_template_id) WHERE variant_id IS NOT NULL AND deleted_at IS NULL;
 INSERT INTO product_variants(id,product_id) VALUES (10,1),(11,1),(12,1);
 INSERT INTO product_shipping_template_bindings(product_id,variant_id,shipping_template_id,priority,enabled)
 VALUES (1,NULL,101,0,TRUE),(1,NULL,102,1,TRUE),(1,11,103,0,TRUE),(1,12,101,5,FALSE);
 INSERT INTO products(id) VALUES(2);
 INSERT INTO product_shipping_template_bindings(product_id,shipping_template_id) VALUES(2,104);
 `)
	if err != nil {
		t.Fatal(err)
	}
	up := readMigrationFile(t, "431_move_product_shipping_bindings_to_skus_and_require_explicit_sku_ownership.up.sql")
	expectCarbonRimMigrationStatementRejection(t, tx, up)
	if _, err = tx.ExecContext(t.Context(), `DELETE FROM product_shipping_template_bindings WHERE product_id=2`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		sku int
		ids string
	}{{10, "101,102"}, {11, "103"}, {12, "101,102"}} {
		var ids string
		if err = tx.QueryRowContext(t.Context(), `SELECT string_agg(shipping_template_id::text,',' ORDER BY priority,id) FROM product_shipping_template_bindings WHERE variant_id=$1 AND enabled AND deleted_at IS NULL`, scenario.sku).Scan(&ids); err != nil || ids != scenario.ids {
			t.Fatalf("SKU %d: got %s, expected %s: %v", scenario.sku, ids, scenario.ids, err)
		}
	}
	var count int
	if err = tx.QueryRowContext(t.Context(), `SELECT count(*) FROM product_shipping_template_bindings WHERE variant_id IS NULL`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("product bindings remain: %d %v", count, err)
	}
	expectCarbonRimMigrationStatementRejection(t, tx, `INSERT INTO product_shipping_template_bindings(product_id,shipping_template_id) VALUES(1,999)`)
	expectCarbonRimMigrationStatementRejection(t, tx, `INSERT INTO product_shipping_template_bindings(product_id,variant_id,shipping_template_id) VALUES(1,0,999)`)
	expectCarbonRimMigrationStatementRejection(t, tx, readMigrationFile(t, "431_move_product_shipping_bindings_to_skus_and_require_explicit_sku_ownership.down.sql"))
}
