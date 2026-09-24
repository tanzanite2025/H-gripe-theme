package migrations_test

import (
	"strings"
	"testing"
)

func TestRefundLoyaltyCashRecoveryDebtUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "345_refund_loyalty_cash_recovery_debt.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS loyalty_points_cash_recovered INTEGER NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS loyalty_points_debt INTEGER NOT NULL DEFAULT 0",
		"chk_refunds_loyalty_points_cash_recovered_non_negative",
		"chk_refunds_loyalty_points_debt_non_negative",
		"DROP INDEX IF EXISTS uq_loyalty_transactions_refund_settlement",
		"'refund_loyalty_cash_recovery_debt'",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestRefundLoyaltyCashRecoveryDebtDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "345_refund_loyalty_cash_recovery_debt.down.sql")
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS chk_refunds_loyalty_points_debt_non_negative",
		"DROP CONSTRAINT IF EXISTS chk_refunds_loyalty_points_cash_recovered_non_negative",
		"DROP COLUMN IF EXISTS loyalty_points_cash_recovered",
		"DROP COLUMN IF EXISTS loyalty_points_debt",
		"'refund_loyalty_points_return'",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
