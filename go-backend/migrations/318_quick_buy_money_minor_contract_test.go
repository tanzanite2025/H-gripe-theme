package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestQuickBuyMoneyMinorMigrationReplacesFloatingSnapshots(t *testing.T) {
	data, err := os.ReadFile("318_quick_buy_money_minor.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"add column if not exists subtotal_snapshot_minor bigint",
		"add column if not exists unit_price_snapshot_minor bigint",
		"drop column if exists subtotal_snapshot",
		"drop column if exists unit_price_snapshot",
		"subtotal_snapshot_minor >= 0",
		"unit_price_snapshot_minor >= 0",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
