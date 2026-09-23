package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestRefundLoyaltyCashDeductionMinorMigrationContract(t *testing.T) {
	up, err := os.ReadFile("309_refund_loyalty_cash_deduction_minor.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile("309_refund_loyalty_cash_deduction_minor.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	upSQL := strings.ToLower(string(up))
	downSQL := strings.ToLower(string(down))
	for _, fragment := range []string{
		"alter table refunds",
		"loyalty_cash_deduction_amount_minor",
		"round(loyalty_cash_deduction_amount * 100)",
		"chk_refunds_loyalty_cash_deduction_minor_non_negative",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"drop constraint if exists chk_refunds_loyalty_cash_deduction_minor_non_negative",
		"drop column if exists loyalty_cash_deduction_amount_minor",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
