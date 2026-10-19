package migrations_test

import "testing"

func TestOrderEvidenceDetachedOwnershipMigrationPreservesHistoryAndGuardsRollback(t *testing.T) {
	tx := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	if _, err := tx.ExecContext(t.Context(), "CREATE TABLE order_evidence_attachments (id BIGINT PRIMARY KEY, storage_key TEXT NOT NULL); INSERT INTO order_evidence_attachments VALUES (1, 'order-evidence/1/old-proof.png')"); err != nil {
		t.Fatal(err)
	}
	up := readMigrationFile(t, "425_retain_detached_order_evidence_attachment_ownership_history.up.sql")
	down := readMigrationFile(t, "425_retain_detached_order_evidence_attachment_ownership_history.down.sql")
	for range 2 {
		if _, err := tx.ExecContext(t.Context(), up); err != nil {
			t.Fatal(err)
		}
	}
	var preserved bool
	if err := tx.QueryRowContext(t.Context(), "SELECT storage_key='order-evidence/1/old-proof.png' AND detached_at IS NULL FROM order_evidence_attachments WHERE id=1").Scan(&preserved); err != nil {
		t.Fatal(err)
	}
	if !preserved {
		t.Fatal("migration changed existing evidence ownership")
	}
	if _, err := tx.ExecContext(t.Context(), "UPDATE order_evidence_attachments SET detached_at=NOW() WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, tx, down, "Cannot discard detached immutable evidence ownership")
	if _, err := tx.ExecContext(t.Context(), "UPDATE order_evidence_attachments SET detached_at=NULL WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(t.Context(), up); err != nil {
		t.Fatal(err)
	}
}
