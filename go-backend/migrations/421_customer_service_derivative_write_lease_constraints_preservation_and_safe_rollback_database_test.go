package migrations_test

import "testing"

func TestCustomerServiceDerivativeWriteLeaseMigrationPreservesExistingAssetsAndGuardsLeaseRollback(t *testing.T) {
	transaction := createIsolatedMediaUploadActorIdentityMigrationTransaction(t)
	up := readMigrationFile(t, "421_add_customer_service_private_derivative_write_lease_for_safe_asset_cleanup.up.sql")
	down := readMigrationFile(t, "421_add_customer_service_private_derivative_write_lease_for_safe_asset_cleanup.down.sql")
	for range 2 {
		if _, err := transaction.ExecContext(t.Context(), up); err != nil {
			t.Fatal(err)
		}
	}
	var token, originalURL, key string
	var expiryIsNull bool
	if err := transaction.QueryRowContext(t.Context(), `SELECT customer_service_derivative_write_token,customer_service_derivative_write_expires_at IS NULL,url,storage_key FROM media_assets WHERE id=1`).
		Scan(&token, &expiryIsNull, &originalURL, &key); err != nil {
		t.Fatal(err)
	}
	if token != "" || !expiryIsNull || originalURL != "https://old.example/uploads/original.webp" || key != "2026/08/20/original.webp" {
		t.Fatal("lease migration changed historical asset data")
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE media_assets SET customer_service_derivative_write_token='11111111-1111-4111-8111-111111111111' WHERE id=1", "media_assets_customer_service_derivative_write_lease_valid")
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction,
		"UPDATE media_assets SET customer_service_derivative_write_expires_at=NOW() WHERE id=1", "media_assets_customer_service_derivative_write_lease_valid")
	if _, err := transaction.ExecContext(t.Context(), "UPDATE media_assets SET customer_service_derivative_write_token='11111111-1111-4111-8111-111111111111',customer_service_derivative_write_expires_at=NOW()+INTERVAL '5 minutes' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	expectMediaUploadActorIdentityMigrationStatementFailure(t, transaction, down, "Cannot remove customer-service derivative write leases")
	if _, err := transaction.ExecContext(t.Context(), "UPDATE media_assets SET customer_service_derivative_write_token='',customer_service_derivative_write_expires_at=NULL WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(t.Context(), down); err != nil {
		t.Fatal(err)
	}
}
