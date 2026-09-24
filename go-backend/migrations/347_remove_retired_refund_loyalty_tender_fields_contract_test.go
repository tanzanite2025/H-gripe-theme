package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveRetiredRefundLoyaltyTenderFieldsMigrationContract(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "347_remove_retired_refund_loyalty_tender_fields.up.sql"))
	for _, fragment := range []string{
		"drop column if exists loyalty_points_returned",
		"drop column if exists loyalty_points_cash_recovered",
		"drop column if exists loyalty_cash_deduction_amount_minor",
		"refund_loyalty_cash_recovery_debt",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing contract fragment %q", fragment)
		}
	}
}
