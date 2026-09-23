package migrations_test

import (
	"strings"
	"testing"
)

func TestReferralLegacyImportMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "295_referral_legacy_import.up.sql"))
	for _, fragment := range []string{
		"LEGACY_REFERRAL_ID",
		"INSERT INTO USER_REFERRAL_IDENTITIES",
		"INSERT INTO REFERRAL_RECORDS",
		"INSERT INTO REFERRAL_REWARD_LOGS",
		"ON CONFLICT DO NOTHING",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration must contain %q", fragment)
		}
	}
	if strings.TrimSpace(readMigrationFile(t, "295_referral_legacy_import.down.sql")) == "" {
		t.Fatal("down migration must not be empty")
	}
}
