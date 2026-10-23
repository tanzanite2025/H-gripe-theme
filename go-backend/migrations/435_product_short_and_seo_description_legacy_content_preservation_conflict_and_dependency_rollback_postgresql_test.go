package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createProductDescriptionColumnMigrationTransactionInDisposablePostgreSQL(t *testing.T) *sql.Tx {
	t.Helper()
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	if connectionURL == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server")
	}
	db, err := sql.Open("postgres", connectionURL)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(); cancel(); require.NoError(t, db.Close()) })
	_, err = tx.ExecContext(ctx, `CREATE TEMP TABLE products(id BIGSERIAL PRIMARY KEY,deleted_at TIMESTAMPTZ,preserved_fact TEXT NOT NULL DEFAULT 'unchanged') ON COMMIT DROP; SET LOCAL search_path TO pg_temp`)
	require.NoError(t, err)
	return tx
}

func captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t *testing.T, tx *sql.Tx) string {
	t.Helper()
	var fingerprint string
	require.NoError(t, tx.QueryRowContext(t.Context(), `SELECT jsonb_build_object(
'columns',(SELECT jsonb_agg(jsonb_build_array(attname,atttypid::text,attnotnull) ORDER BY attnum) FROM pg_attribute WHERE attrelid='products'::regclass AND attnum>0 AND NOT attisdropped),
'rows',(SELECT jsonb_agg(to_jsonb(p) ORDER BY p.id) FROM products p),
'sequence',(SELECT jsonb_build_array(last_value,log_cnt,is_called) FROM pg_temp.products_id_seq))::text`).Scan(&fingerprint))
	return fingerprint
}

func TestPostgreSQLProductDescriptionColumnMigrationPreservesContentAndRejectsAmbiguousCleanupAtomically(t *testing.T) {
	up := readProductDescriptionColumnMigrationSQLForDisposablePostgreSQLTest(t, "up")
	down := readProductDescriptionColumnMigrationSQLForDisposablePostgreSQLTest(t, "down")
	for _, scenario := range []struct{ name, setup, expectedError string }{
		{"canonical_columns_are_unchanged", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN meta_description TEXT; INSERT INTO products(id,short_description,meta_description) VALUES(1,'<p>Canonical &amp; exact</p>','SEO exact'),(2,NULL,''),(3,'   ',NULL)`, ""},
		{"legacy_only_columns_are_renamed_with_indexes_and_deleted_rows_preserved", `ALTER TABLE products ADD COLUMN short_desc TEXT,ADD COLUMN meta_desc VARCHAR(500); CREATE INDEX product_legacy_description_index ON products(short_desc); INSERT INTO products(id,short_desc,meta_desc,deleted_at) VALUES(1,'<img src="/uploads/exact.png">','Legacy SEO',now())`, ""},
		{"dual_columns_copy_only_missing_values_and_keep_exact_whitespace", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN short_desc TEXT,ADD COLUMN meta_description TEXT,ADD COLUMN meta_desc TEXT; INSERT INTO products(id,short_description,short_desc,meta_description,meta_desc) VALUES(1,NULL,'<p>Legacy &amp; exact</p>','','Legacy SEO'),(2,'Keep canonical','','Equal','Equal'),(3,'Same','Same',NULL,NULL),(4,NULL,'   ','  ',NULL),(5,NULL,'',NULL,''),(6,'',NULL,'',NULL)`, ""},
		{"conflicting_short_description_blocks_both_pairs", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN short_desc TEXT,ADD COLUMN meta_description TEXT,ADD COLUMN meta_desc TEXT; INSERT INTO products(id,short_description,short_desc,meta_desc) VALUES(1,'Canonical','Legacy','SEO would copy')`, "short_description columns contain conflicting content"},
		{"conflicting_seo_rolls_back_prior_short_description_copy_and_drop", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN short_desc TEXT,ADD COLUMN meta_description TEXT,ADD COLUMN meta_desc TEXT; INSERT INTO products(id,short_desc,meta_description,meta_desc) VALUES(1,'Short would copy','Canonical SEO','Legacy SEO')`, "meta_description columns contain conflicting content"},
		{"deleted_product_conflict_is_not_ignored", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN short_desc TEXT,ADD COLUMN meta_description TEXT; INSERT INTO products(id,short_description,short_desc,deleted_at) VALUES(1,'Canonical','Legacy',now())`, "conflicting content"},
		{"whitespace_is_real_content_and_cannot_be_silently_trimmed", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN short_desc TEXT,ADD COLUMN meta_description TEXT; INSERT INTO products(id,short_description,short_desc) VALUES(1,' ','Legacy')`, "conflicting content"},
		{"dependent_view_blocks_drop_and_rolls_back_copied_content", `ALTER TABLE products ADD COLUMN short_description TEXT,ADD COLUMN short_desc TEXT,ADD COLUMN meta_description TEXT; INSERT INTO products(id,short_desc) VALUES(1,'Would copy'); CREATE TEMP VIEW product_legacy_description_dependency AS SELECT short_desc FROM products`, "depend on it"},
		{"missing_description_columns_are_rejected", `ALTER TABLE products ADD COLUMN meta_description TEXT`, "no recognized short_description column"},
		{"unregistered_column_type_is_rejected", `ALTER TABLE products ADD COLUMN short_description INTEGER,ADD COLUMN meta_description TEXT`, "recognized text type"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			tx := createProductDescriptionColumnMigrationTransactionInDisposablePostgreSQL(t)
			_, err := tx.ExecContext(t.Context(), scenario.setup)
			require.NoError(t, err)
			before := captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t, tx)
			_, err = tx.ExecContext(t.Context(), `SAVEPOINT product_description_migration_attempt`)
			require.NoError(t, err)
			_, err = tx.ExecContext(t.Context(), up)
			if scenario.expectedError != "" {
				require.ErrorContains(t, err, scenario.expectedError)
				_, err = tx.ExecContext(t.Context(), `ROLLBACK TO SAVEPOINT product_description_migration_attempt`)
				require.NoError(t, err)
				require.Equal(t, before, captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t, tx))
				return
			}
			require.NoError(t, err)
			var columns []byte
			require.NoError(t, tx.QueryRowContext(t.Context(), `SELECT json_agg(attname ORDER BY attname) FROM pg_attribute WHERE attrelid='products'::regclass AND attname IN ('short_desc','short_description','meta_desc','meta_description') AND attnum>0 AND NOT attisdropped`).Scan(&columns))
			require.JSONEq(t, `["meta_description","short_description"]`, string(columns))
			var actual string
			switch scenario.name {
			case "canonical_columns_are_unchanged":
				require.Equal(t, before, captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t, tx))
			case "legacy_only_columns_are_renamed_with_indexes_and_deleted_rows_preserved":
				require.NoError(t, tx.QueryRowContext(t.Context(), `SELECT jsonb_build_array(short_description,meta_description,deleted_at IS NOT NULL,preserved_fact)::text FROM products WHERE id=1`).Scan(&actual))
				require.JSONEq(t, `["<img src=\"/uploads/exact.png\">","Legacy SEO",true,"unchanged"]`, actual)
				require.NoError(t, tx.QueryRowContext(t.Context(), `SELECT pg_get_indexdef('product_legacy_description_index'::regclass)`).Scan(&actual))
				require.Contains(t, actual, "(short_description)")
			default:
				require.NoError(t, tx.QueryRowContext(t.Context(), `SELECT jsonb_agg(jsonb_build_array(id,short_description,meta_description,preserved_fact) ORDER BY id)::text FROM products`).Scan(&actual))
				require.JSONEq(t, `[[1,"<p>Legacy &amp; exact</p>","Legacy SEO","unchanged"],[2,"Keep canonical","Equal","unchanged"],[3,"Same",null,"unchanged"],[4,"   ","  ","unchanged"],[5,null,null,"unchanged"],[6,"","","unchanged"]]`, actual)
			}
			after := captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t, tx)
			_, err = tx.ExecContext(t.Context(), up)
			require.NoError(t, err)
			require.Equal(t, after, captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t, tx), "a repeated migration must not change reconciled content")
			_, err = tx.ExecContext(t.Context(), down)
			require.NoError(t, err)
			require.Equal(t, after, captureProductDescriptionMigrationColumnsRowsAndSequenceFingerprint(t, tx), "down preserves canonical content instead of recreating incompatible duplicates")
		})
	}
}

func readProductDescriptionColumnMigrationSQLForDisposablePostgreSQLTest(t *testing.T, direction string) string {
	t.Helper()
	name := "435_consolidate_product_short_and_seo_descriptions_without_discarding_conflicting_legacy_content." + direction + ".sql"
	if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY"); directory != "" {
		contents, err := os.ReadFile(filepath.Join(directory, name))
		require.NoError(t, err)
		return string(contents)
	}
	return readMigrationFile(t, name)
}
