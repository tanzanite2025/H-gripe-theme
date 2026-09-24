package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration327UsesMinorWarrantyServiceCost(t *testing.T) {
	data, err := os.ReadFile("327_warranty_service_cost_amount_minor.up.sql")
	if err != nil { t.Fatal(err) }
	sql := strings.ToLower(string(data))
	if !strings.Contains(sql, "cost_amount_minor bigint") || !strings.Contains(sql, "drop column if exists cost_amount") {
		t.Fatal("migration must replace warranty service cost with minor units")
	}
}
