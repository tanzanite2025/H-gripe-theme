package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestProductCustomOptionFulfillmentDeltasMigrationContract(t *testing.T) {
	data, err := os.ReadFile("300_product_custom_option_fulfillment_deltas.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"alter table product_custom_option_policies",
		"add column if not exists weight_delta_grams",
		"add column if not exists packaging_weight_delta_grams",
		"chk_product_custom_option_policy_weight_deltas",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
