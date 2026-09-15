package migrations_test

import (
	"strings"
	"testing"
)

func TestRefundLoyaltySettlementUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "254_refund_loyalty_settlement.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS loyalty_settlement_prepared BOOLEAN NOT NULL DEFAULT FALSE",
		"ADD COLUMN IF NOT EXISTS loyalty_points_clawback INTEGER NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS loyalty_points_returned INTEGER NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS loyalty_cash_deduction_amount NUMERIC(12,2) NOT NULL DEFAULT 0",
		"CREATE INDEX IF NOT EXISTS idx_refunds_loyalty_settlement_prepared",
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_loyalty_transactions_refund_settlement",
		"'refund_loyalty_clawback'",
		"'refund_loyalty_clawback_reversal'",
		"'refund_loyalty_points_return'",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestRefundLoyaltySettlementDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "254_refund_loyalty_settlement.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS uq_loyalty_transactions_refund_settlement",
		"DROP INDEX IF EXISTS idx_refunds_loyalty_settlement_prepared",
		"DROP COLUMN IF EXISTS loyalty_cash_deduction_amount",
		"DROP COLUMN IF EXISTS loyalty_points_returned",
		"DROP COLUMN IF EXISTS loyalty_points_clawback",
		"DROP COLUMN IF EXISTS loyalty_settlement_prepared",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
