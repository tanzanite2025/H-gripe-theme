package migrations_test

import "testing"

func TestManualMediaLibraryReviewMigrationPreservesHistoryAllowsRevisionCheckedReviewsAndGuardsRollback(t *testing.T) {
	tx := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	if _, err := tx.ExecContext(t.Context(), "ALTER TABLE media_assets ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW()"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), readMigrationFile(t, "422_add_media_library_business_partitions_trusted_upload_origins_and_customer_service_purposes.up.sql")); err != nil {
		t.Fatal(err)
	}
	up := readMigrationFile(t, "423_add_audited_manual_media_library_classification_and_revision_protection.up.sql")
	down := readMigrationFile(t, "423_add_audited_manual_media_library_classification_and_revision_protection.down.sql")
	for range 2 {
		if _, err := tx.ExecContext(t.Context(), up); err != nil {
			t.Fatal(err)
		}
	}
	var partition, origin, method, url, key string
	var revision int64
	if err := tx.QueryRowContext(t.Context(), "SELECT library_partition, upload_origin, classification_method, classification_revision, url, storage_key FROM media_assets WHERE id=1").Scan(&partition, &origin, &method, &revision, &url, &key); err != nil {
		t.Fatal(err)
	}
	if partition != "unclassified" || origin != "legacy_unknown" || method != "historical_unknown" || revision != 0 || url != "https://old.example/uploads/original.webp" || key != "2026/08/20/original.webp" {
		t.Fatal("manual review migration modified or inferred history")
	}
	for _, statement := range []string{"UPDATE media_assets SET classification_revision=-1 WHERE id=1", "UPDATE media_assets SET classification_revision=1 WHERE id=1", "UPDATE media_assets SET classification_method='manual_reviewed' WHERE id=1"} {
		expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, statement, "media_assets_manual_classification_revision_valid")
	}
	if _, err := tx.ExecContext(t.Context(), "UPDATE media_assets SET library_partition='catalog',classification_method='manual_reviewed',classification_revision=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, down, "Cannot discard manually reviewed media ownership")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, "UPDATE media_assets SET upload_origin='guessed_product_origin' WHERE id=1", "media_assets_business_library_classification_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, "UPDATE media_assets SET customer_service_media_purpose='conversation_attachment' WHERE id=1", "media_assets_business_library_classification_valid")
	if _, err := tx.ExecContext(t.Context(), "UPDATE media_assets SET library_partition='unclassified',classification_method='historical_unknown',classification_revision=0 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, "UPDATE media_assets SET classification_method='manual_reviewed' WHERE id=1", "media_assets_business_library_classification_valid")
}
