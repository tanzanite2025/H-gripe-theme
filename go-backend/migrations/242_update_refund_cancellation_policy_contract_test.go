package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestUpdateRefundCancellationPolicyMigrationContainsProductionAndSignatureDisclosure(t *testing.T) {
	payload, err := os.ReadFile("242_update_refund_cancellation_policy.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	sql := strings.ToLower(string(payload))
	for _, required := range []string{
		"production or material cutting starts",
		"after production begins",
		"$750 usd",
		"signature",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing %q disclosure", required)
		}
	}
}
