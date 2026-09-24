package migrations_test

import (
	"strings"
	"testing"
)

func TestGiftCardAndLoyaltyMoneyMinorMigrationRenamesCentsColumns(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "334_gift_card_and_loyalty_money_minor.up.sql"))
	for _, fragment := range []string{
		"initial_value_cents to initial_value_minor",
		"balance_cents to balance_minor",
		"amount_cents to amount_minor",
		"gift_card_value_cents to gift_card_value_minor",
		"max_value_per_day_cents to max_value_per_day_minor",
		"value_cents to value_minor",
		"check (value_minor > 0)",
		"check (gift_card_value_minor > 0)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

func TestGiftCardAndLoyaltyMoneyMinorMigrationRollbackRestoresChecks(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "334_gift_card_and_loyalty_money_minor.down.sql"))
	for _, fragment := range []string{
		"initial_value_minor_non_negative to initial_value_cents_non_negative",
		"balance_minor_non_negative to balance_cents_non_negative",
		"check (value_cents > 0)",
		"check (gift_card_value_cents > 0)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("rollback migration missing %q", fragment)
		}
	}
}
