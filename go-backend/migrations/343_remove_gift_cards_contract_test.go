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

func TestRemoveGiftCardsRollbackRestoresOnlyEmptySchemaNeededByHistoricalMigrations(t *testing.T) {

	sql := strings.ToLower(readMigrationFile(t, "343_remove_gift_cards.down.sql"))

	for _, statement := range []string{
		"create table if not exists gift_cards",
		"initial_value_minor bigint not null",
		"balance_minor bigint not null",
		"create table if not exists gift_card_transactions",
		"amount_minor bigint not null",
		"refund_id bigint references refunds(id)",
		"create table if not exists gift_card_redemptions",
		"gift_card_value_minor bigint not null",
		"create table if not exists loyalty_program_redeem_options",
		"value_minor bigint not null",
		"add column if not exists min_redeem_points",
		"max_value_per_day_minor bigint not null",
		"card_expiry_days integer not null",
		"add column if not exists gift_card_refund_amount_minor bigint not null default 0",
		"intentionally not recreated",
	} {
		if !strings.Contains(sql, statement) {
			t.Fatalf("rollback migration must contain %q", statement)
		}
	}
}
