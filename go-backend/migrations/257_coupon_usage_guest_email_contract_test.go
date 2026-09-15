package migrations_test

import (
	"strings"
	"testing"
)

func TestCouponUsageGuestEmailUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "257_coupon_usage_guest_email.up.sql")
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT ''",
		"SET email = lower(btrim(COALESCE(orders.shipping_email, '')))",
		"idx_coupon_usage_coupon_email_status",
		"lower(btrim(COALESCE(NEW.email, '')))",
		"ELSIF normalized_email <> ''",
		"lower(btrim(email)) = normalized_email",
		"coupon usage identity required",
		"UPDATE OF coupon_id, user_id, email, status",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestCouponUsageGuestEmailDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "257_coupon_usage_guest_email.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_coupon_usage_coupon_email_status",
		"DROP COLUMN IF EXISTS email",
		"UPDATE OF coupon_id, user_id, status",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
