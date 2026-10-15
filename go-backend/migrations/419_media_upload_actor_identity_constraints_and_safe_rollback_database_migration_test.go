package migrations_test

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func createIsolatedMediaUploadActorIdentityMigrationTransaction(t *testing.T) *sql.Tx {
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
	_, err = transaction.ExecContext(ctx, `
		CREATE TEMP TABLE media_assets (
			id BIGSERIAL PRIMARY KEY, uploader_id BIGINT DEFAULT 0,
			url TEXT NOT NULL, storage_key TEXT, size BIGINT, deleted_at TIMESTAMPTZ
		) ON COMMIT DROP;
		SET LOCAL search_path TO pg_temp;
		INSERT INTO media_assets (uploader_id, url, storage_key, size, deleted_at)
		VALUES (42, 'https://old.example/uploads/original.webp', '2026/08/20/original.webp', 123, NOW());
	`)
	if err != nil {
		t.Fatal(err)
	}
	return transaction
}

func TestMediaUploadActorIdentityMigrationPreservesLegacyDataConstrainsTypedUploadsAndGuardsRollback(t *testing.T) {
	transaction := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	upMigration := readMigrationFile(t, "419_add_explicit_media_upload_actor_identity_and_visitor_session_quota_indexes.up.sql")
	downMigration := readMigrationFile(t, "419_add_explicit_media_upload_actor_identity_and_visitor_session_quota_indexes.down.sql")
	for run := 0; run < 2; run++ {
		if _, err := transaction.ExecContext(t.Context(), upMigration); err != nil {
			t.Fatal(err)
		}
	}
	var actorType, visitorHash, originalURL, storageKey string
	var uploaderID, originalSize int64
	var retainedDeletion bool
	if err := transaction.QueryRowContext(t.Context(), `SELECT uploader_actor_type, uploader_visitor_session_hash, uploader_id, url, storage_key, size, deleted_at IS NOT NULL FROM media_assets WHERE id = 1`).Scan(
		&actorType, &visitorHash, &uploaderID, &originalURL, &storageKey, &originalSize, &retainedDeletion); err != nil {
		t.Fatal(err)
	}
	if actorType != "legacy_account" || visitorHash != "" || uploaderID != 42 || originalURL != "https://old.example/uploads/original.webp" || storageKey != "2026/08/20/original.webp" || originalSize != 123 || !retainedDeletion {
		t.Fatal("migration changed legacy identity, object references, size, or retained deletion")
	}
	for _, invalidIdentitySQL := range []string{
		`INSERT INTO media_assets (url, uploader_id, uploader_actor_type) VALUES ('invalid', 0, 'administrator')`,
		`INSERT INTO media_assets (url, uploader_id, uploader_actor_type) VALUES ('invalid', NULL, 'customer')`,
		`INSERT INTO media_assets (url, uploader_id, uploader_actor_type) VALUES ('invalid', 1, 'unknown')`,
		`INSERT INTO media_assets (url, uploader_id, uploader_actor_type, uploader_visitor_session_hash) VALUES ('invalid', 1, 'visitor', repeat('a', 64))`,
		`INSERT INTO media_assets (url, uploader_id, uploader_actor_type, uploader_visitor_session_hash) VALUES ('invalid', 0, 'visitor', repeat('A', 64))`,
		`INSERT INTO media_assets (url, uploader_id, uploader_actor_type, uploader_visitor_session_hash) VALUES ('invalid', 42, 'customer', repeat('a', 64))`,
	} {
		expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, invalidIdentitySQL, "media_assets_upload_actor_identity_valid")
	}
	if _, err := transaction.ExecContext(t.Context(), `
		INSERT INTO media_assets (url, uploader_id, uploader_actor_type, uploader_visitor_session_hash)
		VALUES ('visitor', 0, 'visitor', repeat('a', 64)), ('administrator', 42, 'administrator', ''), ('customer', 42, 'customer', '');
	`); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, downMigration, "Cannot remove media upload actor identity")
	var count int
	if err := transaction.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM media_assets`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("rollback guard lost asset rows: %d", count)
	}
	// Only fixture records are removed; the whole test uses a temporary table.
	if _, err := transaction.ExecContext(t.Context(), `DELETE FROM media_assets WHERE uploader_actor_type <> 'legacy_account'`); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(t.Context(), downMigration); err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRowContext(t.Context(), `SELECT uploader_id, url, storage_key, size FROM media_assets WHERE id = 1`).Scan(&uploaderID, &originalURL, &storageKey, &originalSize); err != nil {
		t.Fatal(err)
	}
	if uploaderID != 42 || originalSize != 123 || storageKey != "2026/08/20/original.webp" || originalURL != "https://old.example/uploads/original.webp" {
		t.Fatal("safe rollback changed original data")
	}
}

func expectMediaUploadActorIdentityMigrationStatementFailure(t *testing.T, transaction *sql.Tx, statement, expectedErrorFragment string) {
	t.Helper()
	if _, err := transaction.ExecContext(t.Context(), "SAVEPOINT media_upload_actor_constraint_check"); err != nil {
		t.Fatal(err)
	}
	_, statementErr := transaction.ExecContext(t.Context(), statement)
	if _, err := transaction.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT media_upload_actor_constraint_check"); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(t.Context(), "RELEASE SAVEPOINT media_upload_actor_constraint_check"); err != nil {
		t.Fatal(err)
	}
	if statementErr == nil || !strings.Contains(statementErr.Error(), expectedErrorFragment) {
		t.Fatalf("expected failure containing %q, got %v", expectedErrorFragment, statementErr)
	}
}
