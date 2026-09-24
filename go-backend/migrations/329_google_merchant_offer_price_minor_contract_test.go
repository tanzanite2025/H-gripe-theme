package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration328UsesMinorMerchantOfferOverrides(t *testing.T) {
	data, err := os.ReadFile("329_google_merchant_offer_price_minor.up.sql")
	if err != nil { t.Fatal(err) }
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{"price_override_minor bigint", "sale_price_override_minor bigint", "drop column if exists price_override"} {
		if !strings.Contains(sql, fragment) { t.Fatalf("migration missing %q", fragment) }
	}
}
