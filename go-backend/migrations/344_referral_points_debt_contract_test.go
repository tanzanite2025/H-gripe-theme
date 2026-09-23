package migrations_test

import (
	"strings"
	"testing"
)

func TestReferralPointsDebtMigrationContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "344_referral_points_debt.up.sql"))
	for _, fragment := range []string{
		"alter table user_loyalty",
		"add column if not exists debt_points integer not null default 0",
		"add constraint debt_points_non_negative check (debt_points >= 0)",
		"alter table loyalty_transactions",
		"add column if not exists debt_balance integer not null default 0",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "344_referral_points_debt.down.sql"))
	for _, fragment := range []string{
		"drop column if exists debt_balance",
		"drop constraint if exists debt_points_non_negative",
		"drop column if exists debt_points",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
