package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestProductCustomOptionProductionReturnPoliciesMigrationContract(t *testing.T) {
	data, err := os.ReadFile("306_product_custom_option_production_return_policies.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"alter table product_custom_option_policies",
		"add column if not exists production_lead_time_days",
		"add column if not exists requires_production",
		"add column if not exists cancellation_policy",
		"add column if not exists return_policy",
		"production_lead_time_days >= 0",
		"cancellation_policy in ('standard', 'before_production', 'never')",
		"return_policy in ('standard', 'not_allowed')",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
