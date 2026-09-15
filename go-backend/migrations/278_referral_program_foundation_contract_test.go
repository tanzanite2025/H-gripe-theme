package migrations_test

import (
	"strings"
	"testing"
)

func TestReferralProgramFoundationMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "278_referral_program_foundation.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "278_referral_program_foundation.down.sql"))

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS REFERRAL_PROGRAM_CONFIGS",
		"ENABLED BOOLEAN NOT NULL DEFAULT FALSE",
		"CREATE TABLE IF NOT EXISTS USER_REFERRAL_IDENTITIES",
		"CREATE TABLE IF NOT EXISTS REFERRAL_RECORDS",
		"CREATE TABLE IF NOT EXISTS REFERRAL_REWARD_LOGS",
		"CREATE TABLE IF NOT EXISTS REFERRAL_TRANSITIONS",
		"UNIQUE INDEX IF NOT EXISTS UQ_REFERRAL_RECORDS_REFEREE",
		"UNIQUE INDEX IF NOT EXISTS UQ_REFERRAL_RECORDS_ORDER",
		"REFERRAL_RECORDS_NOT_SELF",
		"TRG_REFERRAL_TRANSITIONS_APPEND_ONLY",
		"WHERE STATUS = 'VESTING'",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("up migration must contain %q", required)
		}
	}

	for _, required := range []string{
		"DROP TABLE IF EXISTS REFERRAL_TRANSITIONS",
		"DROP TABLE IF EXISTS REFERRAL_REWARD_LOGS",
		"DROP TABLE IF EXISTS REFERRAL_RECORDS",
		"DROP TABLE IF EXISTS USER_REFERRAL_IDENTITIES",
		"DROP TABLE IF EXISTS REFERRAL_PROGRAM_CONFIGS",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("down migration must contain %q", required)
		}
	}
}
