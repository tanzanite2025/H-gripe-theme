package migrations_test

import (
	"strings"
	"testing"
)

func TestCouponUsageReversalAuditMigrationContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "238_coupon_usage_reversal_audit.up.sql"))
	for _, fragment := range []string{
		"add column if not exists status varchar(20) not null default 'applied'",
		"add column if not exists reversed_at timestamptz",
		"add column if not exists reversal_reason text",
		"coupon_usage_status_valid",
		"check (status in ('applied', 'reversed'))",
		"idx_coupon_usage_coupon_user_status",
		"and status = 'applied'",
		"create or replace function enforce_coupon_per_user_usage_limit()",
		"before insert or update of coupon_id, user_id, status on coupon_usage",
		"execute function enforce_coupon_per_user_usage_limit()",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("coupon usage reversal audit migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "238_coupon_usage_reversal_audit.down.sql"))
	for _, fragment := range []string{
		"drop index if exists idx_coupon_usage_coupon_user_status",
		"drop constraint if exists coupon_usage_status_valid",
		"drop column if exists reversal_reason",
		"drop column if exists reversed_at",
		"drop column if exists status",
		"before insert or update of coupon_id, user_id on coupon_usage",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("coupon usage reversal audit rollback is missing contract fragment %q", fragment)
		}
	}
}
