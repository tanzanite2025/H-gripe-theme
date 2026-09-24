package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveGiftCardsMigrationRemovesForeignKeyDependentsBeforeLedgerRows(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "343_remove_gift_cards.up.sql"))

	for _, statement := range []string{
		"drop table if exists gift_card_transactions;",
		"drop table if exists gift_card_redemptions;",
		"delete from loyalty_transactions",
		"drop table if exists gift_cards;",
		"drop table if exists loyalty_program_redeem_options;",
	} {
		if !strings.Contains(sql, statement) {
			t.Fatalf("migration must contain %q", statement)
		}
	}

	transactionsIndex := strings.Index(sql, "drop table if exists gift_card_transactions;")
	redemptionsIndex := strings.Index(sql, "drop table if exists gift_card_redemptions;")
	ledgerDeleteIndex := strings.Index(sql, "delete from loyalty_transactions\nwhere lower(coalesce(source, '')) = 'giftcard';")
	cardsIndex := strings.Index(sql, "drop table if exists gift_cards;")
	if !(transactionsIndex < redemptionsIndex && redemptionsIndex < ledgerDeleteIndex && ledgerDeleteIndex < cardsIndex) {
		t.Fatal("gift-card dependents must be dropped before the loyalty ledger and cards")
	}
}

func TestRemoveGiftCardsRollbackDoesNotRecreateRetiredGiftCardSchema(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "343_remove_gift_cards.down.sql"))
	for _, statement := range []string{"create table", "gift_cards", "redeem_options", "add column"} {
		if strings.Contains(sql, statement) {
			t.Fatalf("rollback migration must not recreate retired schema: %q", statement)
		}
	}
}
