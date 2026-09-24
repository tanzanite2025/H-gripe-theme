package migrations_test

import (
	"strings"
	"testing"
)

func TestReferralPointsOnlyMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "304_referral_points_only.up.sql"))
	for _, fragment := range []string{
		"DROP COLUMN IF EXISTS REFEREE_BENEFIT_MAX_AMOUNT_MINOR",
		"DROP COLUMN IF EXISTS COUPON_STACKABLE",
		"DROP CONSTRAINT IF EXISTS REFERRAL_REWARD_LOGS_TARGET_VALID",
		"DELETE FROM REFERRAL_REWARD_LOGS",
		"DROP COLUMN IF EXISTS COUPON_ID",
		"CHECK (REWARD_TYPE = 'POINTS' AND POINTS_AMOUNT > 0)",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	down := strings.ToUpper(readMigrationFile(t, "304_referral_points_only.down.sql"))
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS REFEREE_BENEFIT_MAX_AMOUNT_MINOR",
		"ADD COLUMN IF NOT EXISTS COUPON_STACKABLE",
		"ADD COLUMN IF NOT EXISTS COUPON_ID",
		"REWARD_TYPE IN ('POINTS', 'COUPON')",
	} {
		if !strings.Contains(down, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}
