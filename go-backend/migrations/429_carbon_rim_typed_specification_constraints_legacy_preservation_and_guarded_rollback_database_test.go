package migrations_test

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func createIsolatedCarbonRimSpecificationMigrationTransaction(t *testing.T) *sql.Tx {
	t.Helper()
	host, username, password := os.Getenv("DB_HOST"), os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")
	if host == "" || username == "" || password == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}
	port, databaseName := os.Getenv("DB_PORT"), os.Getenv("DB_NAME")
	if port == "" {
		port = "5432"
	}
	if databaseName == "" {
		databaseName = "postgres"
	}
	connectionURL := &url.URL{Scheme: "postgres", User: url.UserPassword(username, password), Host: net.JoinHostPort(host, port), Path: "/" + databaseName, RawQuery: "sslmode=disable&connect_timeout=5"}
	db, err := sql.Open("postgres", connectionURL.String())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		_ = db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transaction.Rollback(); cancel(); _ = db.Close() })
	_, err = transaction.ExecContext(ctx, `CREATE TEMP TABLE products (id BIGSERIAL PRIMARY KEY, product_specification_template_id BIGINT, deleted_at TIMESTAMPTZ) ON COMMIT DROP; CREATE TEMP TABLE product_variants (id BIGSERIAL PRIMARY KEY) ON COMMIT DROP; SET LOCAL search_path TO pg_temp; INSERT INTO products (product_specification_template_id) VALUES (77);`)
	if err != nil {
		t.Fatal(err)
	}
	return transaction
}

func TestDedicatedCarbonRimMigrationPreservesTemplateProductsRejectsMissingVersionsAndGuardsDataRollback(t *testing.T) {
	tx := createIsolatedCarbonRimSpecificationMigrationTransaction(t)
	up := readMigrationFile(t, "429_add_dedicated_carbon_rim_product_and_variant_specifications.up.sql")
	down := readMigrationFile(t, "429_add_dedicated_carbon_rim_product_and_variant_specifications.down.sql")
	if _, err := tx.Exec(up); err != nil {
		t.Fatal(err)
	}
	var preserved bool
	if err := tx.QueryRow("SELECT product_type = '' AND carbon_rim_specification IS NULL AND product_specification_template_id = 77 FROM products WHERE id = 1").Scan(&preserved); err != nil || !preserved {
		t.Fatalf("legacy template changed: %v", err)
	}
	for _, invalid := range []string{
		"INSERT INTO products(product_type, carbon_rim_specification) VALUES ('carbon_rim', '{}')",
		"INSERT INTO products(product_type, carbon_rim_specification) VALUES ('carbon_rim', '{\"schema_version\":2}')",
		"INSERT INTO products(product_type, carbon_rim_specification, product_specification_template_id) VALUES ('carbon_rim', '{\"schema_version\":1}', 77)",
	} {
		expectCarbonRimMigrationStatementRejection(t, tx, invalid)
	}
	if _, err := tx.Exec("INSERT INTO products(product_type, carbon_rim_specification, deleted_at) VALUES ('carbon_rim', '{\"schema_version\":1}', NOW())"); err != nil {
		t.Fatal(err)
	}
	expectCarbonRimMigrationStatementRejection(t, tx, down)
	if _, err := tx.Exec("DELETE FROM products WHERE product_type = 'carbon_rim'; INSERT INTO product_variants(carbon_rim_specification) VALUES ('{}')"); err != nil {
		t.Fatal(err)
	}
	expectCarbonRimMigrationStatementRejection(t, tx, down)
	if _, err := tx.Exec("DELETE FROM product_variants"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(down); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(up); err != nil {
		t.Fatal(err)
	}
}

func expectCarbonRimMigrationStatementRejection(t *testing.T, tx *sql.Tx, statement string) {
	t.Helper()
	if _, err := tx.Exec("SAVEPOINT expected_carbon_rim_rejection"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(statement); err == nil {
		t.Fatal("expected statement to reject invalid state or unsafe rollback")
	}
	if _, err := tx.Exec("ROLLBACK TO SAVEPOINT expected_carbon_rim_rejection"); err != nil {
		t.Fatal(err)
	}
}
