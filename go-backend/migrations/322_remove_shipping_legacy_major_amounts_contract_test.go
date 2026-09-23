package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestRemoveShippingLegacyMajorAmountsMigration(t *testing.T) {
	data, err := os.ReadFile("322_remove_shipping_legacy_major_amounts.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"alter table shipping_templates",
		"drop column if exists free_threshold",
		"drop column if exists default_fee",
		"alter table shipping_rules",
		"drop column if exists fee",
		"drop column if exists additional",
		"update shipping_rules",
		"set min_value = 0",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	if strings.Contains(sql, "drop column if exists min_value") || strings.Contains(sql, "drop column if exists max_value") {
		t.Fatal("dimensional shipping thresholds must remain available for weight/quantity rules")
	}
}
