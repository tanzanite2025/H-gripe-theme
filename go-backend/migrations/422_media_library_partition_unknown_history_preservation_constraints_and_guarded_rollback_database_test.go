package migrations_test

import "testing"

func TestMediaLibraryBusinessPartitionMigrationPreservesUnknownHistoryConstrainsPurposesAndGuardsRollback(t *testing.T) {
	transaction := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	if _, err := transaction.ExecContext(t.Context(), "ALTER TABLE media_assets ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW()"); err != nil {
		t.Fatal(err)
	}
	up := readMigrationFile(t, "422_add_media_library_business_partitions_trusted_upload_origins_and_customer_service_purposes.up.sql")
	down := readMigrationFile(t, "422_add_media_library_business_partitions_trusted_upload_origins_and_customer_service_purposes.down.sql")
	for range 2 {
		if _, err := transaction.ExecContext(t.Context(), up); err != nil {
			t.Fatal(err)
		}
	}
	var partition, origin, method, purpose, url, key string
	if err := transaction.QueryRowContext(t.Context(), "SELECT library_partition, upload_origin, classification_method, customer_service_media_purpose, url, storage_key FROM media_assets WHERE id=1").
		Scan(&partition, &origin, &method, &purpose, &url, &key); err != nil {
		t.Fatal(err)
	}
	if partition != "unclassified" || origin != "legacy_unknown" || method != "historical_unknown" || purpose != "" || url != "https://old.example/uploads/original.webp" || key != "2026/08/20/original.webp" {
		t.Fatal("partition migration modified or guessed historical ownership")
	}
	for _, statement := range []string{
		"UPDATE media_assets SET library_partition='invalid' WHERE id=1",
		"UPDATE media_assets SET upload_origin='client_forged_origin' WHERE id=1",
		"UPDATE media_assets SET classification_method='guessed_from_current_references' WHERE id=1",
		"UPDATE media_assets SET customer_service_media_purpose='conversation_attachment' WHERE id=1",
		"UPDATE media_assets SET library_partition='customer_service' WHERE id=1",
	} {
		expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, statement, "media_assets_business_library_classification_valid")
	}
	if _, err := transaction.ExecContext(t.Context(), "UPDATE media_assets SET library_partition='customer_service', upload_origin='customer_service_conversation_upload', classification_method='recorded_upload_context', customer_service_media_purpose='conversation_attachment' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, down, "Cannot discard recorded media library ownership")
	if _, err := transaction.ExecContext(t.Context(), "UPDATE media_assets SET library_partition='unclassified', upload_origin='legacy_unknown', classification_method='historical_unknown', customer_service_media_purpose='' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
}
