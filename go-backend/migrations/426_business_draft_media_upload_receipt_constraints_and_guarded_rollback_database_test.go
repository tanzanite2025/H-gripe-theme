package migrations_test

import "testing"

func TestBusinessDraftMediaUploadReceiptMigrationPreservesExistingAssetsAndGuardsRollback(t *testing.T) {
	tx := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	up := readMigrationFile(t, "426_add_business_draft_media_preallocated_upload_receipts.up.sql")
	down := readMigrationFile(t, "426_add_business_draft_media_preallocated_upload_receipts.down.sql")
	for range 2 {
		if _, err := tx.ExecContext(t.Context(), up); err != nil {
			t.Fatal(err)
		}
	}
	const insert = "INSERT INTO business_draft_media_upload_receipts (id,uploader_id,upload_origin,original_object_key,object_keys_json,content_sha256,size,mime_type,status,expires_at) VALUES ('d92bdd48-19df-472f-b5de-657a62f46ab0',42,'administrator_catalog_upload','business-draft-media/d92bdd48-19df-472f-b5de-657a62f46ab0/original.png','[\"business-draft-media/d92bdd48-19df-472f-b5de-657a62f46ab0/original.png\"]',repeat('a',64),12,'image/png','uploading',NOW()+INTERVAL '24 hours')"
	if _, err := tx.ExecContext(t.Context(), insert); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, "UPDATE business_draft_media_upload_receipts SET status='bound'", "business_draft_media_receipt_status_asset_identity_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, "UPDATE business_draft_media_upload_receipts SET original_object_key='order-evidence/1/2/file.png'", "check constraint")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, "UPDATE business_draft_media_upload_receipts SET content_sha256='invalid'", "business_draft_media_receipt_content_digest_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, down, "Cannot discard business draft media upload receipts")
	if _, err := tx.ExecContext(t.Context(), "UPDATE business_draft_media_upload_receipts SET asset_id=1,status='ready'"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), "UPDATE business_draft_media_upload_receipts SET status='bound'"); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, down, "Cannot discard business draft media upload receipts")
	if _, err := tx.ExecContext(t.Context(), "DELETE FROM business_draft_media_upload_receipts"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := tx.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM media_assets WHERE id=1").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("migration changed preexisting assets")
	}
}
