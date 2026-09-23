package migrations_test

import (
	"strings"
	"testing"
)

func TestPaymentOperationIdempotencyLeasesMigrationDefinesRecoveryStateMachine(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "308_payment_operation_idempotency_leases.up.sql"))
	for _, fragment := range []string{
		"add column if not exists claim_token varchar(64)",
		"add column if not exists lease_expires_at timestamptz",
		"add column if not exists reconciliation_started_at timestamptz",
		"when scope = 'paypal_capture' then 'reconciling'",
		"status in ('pending', 'reconciling', 'completed')",
		"idx_payment_operation_idempotencies_claimable",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("payment operation lease migration is missing contract fragment %q", fragment)
		}
	}
}

func TestPaymentOperationIdempotencyLeasesDownMigrationRemovesLeaseState(t *testing.T) {
	downSQL := strings.ToLower(readMigrationFile(t, "308_payment_operation_idempotency_leases.down.sql"))
	for _, fragment := range []string{
		"status = 'pending'",
		"where status = 'reconciling'",
		"drop column if exists reconciliation_started_at",
		"drop column if exists lease_expires_at",
		"drop column if exists claim_token",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("payment operation lease down migration is missing contract fragment %q", fragment)
		}
	}
}
