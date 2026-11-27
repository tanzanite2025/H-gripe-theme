package migrations_test

import (
	"database/sql"
	"testing"
)

func TestOptionalGenericProductSchemaDatabaseConstraintRejectsMissingOrInvalidVersions(t *testing.T) {
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	if _, err := tx.ExecContext(t.Context(), `ALTER TABLE products ADD COLUMN product_type TEXT NOT NULL DEFAULT ''`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), readMigrationFile(t, "448_add_optional_generic_product_schema.up.sql")); err != nil {
		t.Fatalf("apply generic product schema migration: %v", err)
	}

	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO products(product_type, generic_product_schema)
VALUES ('', '{"schema_version":1,"material":"Aluminium"}'), ('spoke', NULL);`); err != nil {
		t.Fatalf("valid generic or dedicated product schema was rejected: %v", err)
	}
	for _, invalidSchema := range []string{
		`'{}'`,
		`'{"material":"Aluminium"}'`,
		`'{"schema_version":null}'`,
		`'{"schema_version":2}'`,
		`'[]'`,
	} {
		statement := "INSERT INTO products(product_type, generic_product_schema) VALUES ('', " + invalidSchema + ")"
		expectOptionalGenericProductSchemaDatabaseConstraintRejection(t, tx, statement)
	}
	expectOptionalGenericProductSchemaDatabaseConstraintRejection(
		t,
		tx,
		`INSERT INTO products(product_type, generic_product_schema) VALUES ('spoke', '{"schema_version":1}')`,
	)

	down := readMigrationFile(t, "448_add_optional_generic_product_schema.down.sql")
	expectOptionalGenericProductSchemaDatabaseConstraintRejection(t, tx, down)
	if _, err := tx.ExecContext(t.Context(), "DELETE FROM products WHERE generic_product_schema IS NOT NULL"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), down); err != nil {
		t.Fatalf("rollback generic product schema after its facts were removed: %v", err)
	}
	var columnExists bool
	if err := tx.QueryRowContext(t.Context(), `
SELECT EXISTS (
    SELECT 1 FROM pg_attribute
    WHERE attrelid = 'products'::regclass
      AND attname = 'generic_product_schema'
      AND NOT attisdropped
)`).Scan(&columnExists); err != nil || columnExists {
		t.Fatalf("generic schema column still exists after rollback: exists=%t err=%v", columnExists, err)
	}
}

func expectOptionalGenericProductSchemaDatabaseConstraintRejection(t *testing.T, tx *sql.Tx, statement string) {
	t.Helper()
	const savepointName = "expected_generic_product_schema_rejection"
	if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), statement); err == nil {
		t.Fatalf("expected generic product schema constraint to reject statement %q", statement)
	}
	if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepointName); err != nil {
		t.Fatal(err)
	}
}
