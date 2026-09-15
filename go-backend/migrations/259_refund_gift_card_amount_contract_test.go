package migrations_test

import (
	"strings"
	"testing"
)

func TestRefundGiftCardAmountUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "259_refund_gift_card_amount.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS gift_card_refund_amount NUMERIC(12,2) NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS refund_id BIGINT",
		"gift_card_transactions_refund_fk",
		"CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_refund_id",
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_gift_card_transactions_refund_card",
		"WHERE type = 'refund' AND refund_id IS NOT NULL",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestRefundGiftCardAmountDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "259_refund_gift_card_amount.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS uq_gift_card_transactions_refund_card",
		"DROP INDEX IF EXISTS idx_gift_card_transactions_refund_id",
		"DROP CONSTRAINT gift_card_transactions_refund_fk",
		"DROP COLUMN IF EXISTS refund_id",
		"DROP COLUMN IF EXISTS gift_card_refund_amount",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
