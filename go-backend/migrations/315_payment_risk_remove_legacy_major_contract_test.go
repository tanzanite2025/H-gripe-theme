package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestPaymentRiskLegacyMajorMigrationRemovesFloatTotals(t *testing.T) {
	data, err := os.ReadFile("315_payment_risk_remove_legacy_major.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"drop column if exists successful_payment_amount",
		"drop column if exists dispute_amount",
		"drop column if exists refund_amount",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

