package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPaymentRiskAmountMinorMigrationContract(t *testing.T) {
	up, err := os.ReadFile(filepath.Join("312_payment_risk_amount_minor.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile(filepath.Join("312_payment_risk_amount_minor.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	upSQL := strings.ToLower(string(up))
	for _, fragment := range []string{
		"add column if not exists amount_minor bigint",
		"successful_payment_amount_minor_by_currency jsonb",
		"dispute_amount_minor_by_currency jsonb",
		"refund_amount_minor_by_currency jsonb",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	downSQL := strings.ToLower(string(down))
	for _, fragment := range []string{
		"drop column if exists amount_minor",
		"drop column if exists successful_payment_amount_minor_by_currency",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
