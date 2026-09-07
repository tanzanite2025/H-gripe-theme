package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestRefreshRefundCancellationPolicyMigrationUsesStageBasedLanguage(t *testing.T) {
	payload, err := os.ReadFile("246_refresh_refund_cancellation_policy.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	sql := strings.ToLower(string(payload))
	for _, required := range []string{
		"for stocked, non-custom items",
		"before production or material cutting starts",
		"15%-20% custom handling fee",
		"restocking & refurbishment",
		"2026-09-04t00:00:00z",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing %q", required)
		}
	}
	if strings.Contains(sql, "within 24 hours") || strings.Contains(sql, "24 hours after payment") {
		t.Fatal("migration still contains the legacy payment-time cancellation rule")
	}
}
