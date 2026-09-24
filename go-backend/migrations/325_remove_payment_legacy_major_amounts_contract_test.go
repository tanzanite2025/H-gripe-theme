package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration325RemovesPaymentLegacyMajorAmounts(t *testing.T) {
	data, err := os.ReadFile("325_remove_payment_legacy_major_amounts.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"drop column if exists amount",
		"drop column if exists gift_card_refund_amount",
		"drop column if exists recommended_amount",
		"add column if not exists amount_minor bigint",
		"set min_amount_minor = 0",
		"max_amount_minor = 0",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	if strings.Contains(sql, "coalesce(min_amount, 0) * 100") || strings.Contains(sql, "coalesce(max_amount, 0) * 100") {
		t.Fatal("migration must not invent a currency for legacy payment thresholds")
	}
}
