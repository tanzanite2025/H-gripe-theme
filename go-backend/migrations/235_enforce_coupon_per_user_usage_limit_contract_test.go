package migrations_test

import (
	"strings"
	"testing"
)

func TestEnforceCouponPerUserUsageLimitMigrationContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "235_enforce_coupon_per_user_usage_limit.up.sql"))
	for _, fragment := range []string{
		"create index if not exists idx_coupon_usage_coupon_user",
		"on coupon_usage(coupon_id, user_id)",
		"create or replace function enforce_coupon_per_user_usage_limit()",
		"select usage_limit_per_user",
		"for update",
		"if per_user_limit > 0 and new.user_id > 0 then",
		"select count(*)",
		"existing_usage_count >= per_user_limit",
		"coupon per-user usage limit reached",
		"before insert or update of coupon_id, user_id on coupon_usage",
		"execute function enforce_coupon_per_user_usage_limit()",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("coupon per-user usage limit migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "235_enforce_coupon_per_user_usage_limit.down.sql"))
	for _, fragment := range []string{
		"drop trigger if exists trigger_enforce_coupon_per_user_usage_limit on coupon_usage",
		"drop function if exists enforce_coupon_per_user_usage_limit()",
		"drop index if exists idx_coupon_usage_coupon_user",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("coupon per-user usage limit down migration is missing contract fragment %q", fragment)
		}
	}
}
