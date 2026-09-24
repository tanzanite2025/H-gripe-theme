package migrations_test

import (
	"strings"
	"testing"
)

func TestReferralAntiFraudSignalsMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "294_referral_anti_fraud_signals.up.sql"))
	for _, fragment := range []string{
		"ALTER TABLE REFERRAL_RECORDS",
		"ADD COLUMN IF NOT EXISTS CLIENT_IP_SUBNET_HASH",
		"IDX_REFERRAL_RECORDS_IP_SUBNET_CREATED",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration must contain %q", fragment)
		}
	}
	if strings.TrimSpace(readMigrationFile(t, "294_referral_anti_fraud_signals.down.sql")) == "" {
		t.Fatal("down migration must not be empty")
	}
}
