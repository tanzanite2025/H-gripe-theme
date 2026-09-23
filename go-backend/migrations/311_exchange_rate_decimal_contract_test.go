package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExchangeRateDecimalMigrationPersistsCanonicalRate(t *testing.T) {
	root := filepath.Join("311_exchange_rate_decimal.up.sql")
	data, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"add column if not exists rate_decimal numeric(30,15)",
		"set rate_decimal = rate::numeric(30,15)",
		"rate_decimal_positive",
	} {
		if !strings.Contains(sql, strings.ToLower(fragment)) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
