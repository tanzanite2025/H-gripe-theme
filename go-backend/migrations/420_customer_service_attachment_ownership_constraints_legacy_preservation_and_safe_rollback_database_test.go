package migrations_test

import "testing"

func TestCustomerServiceAttachmentBindingMigrationPreservesLegacyURLsRejectsInvalidRelationsAndProtectsRollback(t *testing.T) {
	transaction := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	if _, err := transaction.ExecContext(t.Context(), `
		CREATE TEMP TABLE tickets(id BIGSERIAL PRIMARY KEY) ON COMMIT DROP;
		CREATE TEMP TABLE ticket_messages(id BIGSERIAL PRIMARY KEY,ticket_id BIGINT NOT NULL REFERENCES tickets(id),attachments TEXT) ON COMMIT DROP;
		INSERT INTO tickets DEFAULT VALUES;
		INSERT INTO ticket_messages(ticket_id,attachments) VALUES(1,'["https://old.example/uploads/photo.webp"]');
	`); err != nil {
		t.Fatal(err)
	}
	up := readMigrationFile(t, "420_add_customer_service_message_media_asset_bindings_and_pending_upload_sessions.up.sql")
	down := readMigrationFile(t, "420_add_customer_service_message_media_asset_bindings_and_pending_upload_sessions.down.sql")
	for iteration := 0; iteration < 2; iteration++ {
		if _, err := transaction.ExecContext(t.Context(), up); err != nil {
			t.Fatal(err)
		}
	}
	var format, attachments string
	if err := transaction.QueryRowContext(t.Context(), "SELECT attachment_reference_format,attachments FROM ticket_messages WHERE id=1").Scan(&format, &attachments); err != nil {
		t.Fatal(err)
	}
	if format != "legacy_urls" || attachments != `["https://old.example/uploads/photo.webp"]` {
		t.Fatal("migration changed historical attachment authority or URLs")
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE ticket_messages SET attachment_reference_format='media_asset_bindings' WHERE id=1", "ticket_messages_attachment_reference_format_valid")
	if _, err := transaction.ExecContext(t.Context(), `
		INSERT INTO ticket_messages(ticket_id,attachments,attachment_reference_format) VALUES(1,'','media_asset_bindings');
		INSERT INTO media_upload_sessions(id,business_context_type,business_context_id,upload_actor_type,upload_actor_account_id,object_key,status,expires_at)
		VALUES('11111111-1111-4111-8111-111111111111','customer_service',1,'administrator',42,'customer-service/attachments/originals/photo.png','uploading',NOW()+INTERVAL '1 day');
	`); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE media_upload_sessions SET upload_actor_account_id=0", "media_upload_sessions_owner_identity_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE media_upload_sessions SET status='ready'", "media_upload_sessions_registered_asset_state_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE media_upload_sessions SET object_key='media-derivatives/public.png'", "media_upload_sessions_customer_service_original_namespace_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE media_upload_sessions SET business_context_id=999", "violates foreign key constraint")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index) VALUES(2,999,0)", "violates foreign key constraint")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index) VALUES(2,1,4)", "violates check constraint")
	if _, err := transaction.ExecContext(t.Context(), `
		UPDATE media_upload_sessions SET media_asset_id=1,status='bound';
		INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index,upload_session_id)
		VALUES(2,1,0,'11111111-1111-4111-8111-111111111111');
	`); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"DELETE FROM media_assets WHERE id=1",
		"DELETE FROM ticket_messages WHERE id=2",
		"DELETE FROM media_upload_sessions",
		"INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index,upload_session_id) VALUES(2,1,1,'11111111-1111-4111-8111-111111111111')",
	} {
		expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, statement, "violates")
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, down, "Cannot erase customer-service attachment ownership")
	if _, err := transaction.ExecContext(t.Context(), `
		DELETE FROM ticket_message_media_attachments;
		DELETE FROM media_upload_sessions;
		DELETE FROM ticket_messages WHERE id=2;
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRowContext(t.Context(), "SELECT attachments FROM ticket_messages WHERE id=1").Scan(&attachments); err != nil {
		t.Fatal(err)
	}
	if attachments != `["https://old.example/uploads/photo.webp"]` {
		t.Fatal("safe rollback changed historical URLs")
	}
}
