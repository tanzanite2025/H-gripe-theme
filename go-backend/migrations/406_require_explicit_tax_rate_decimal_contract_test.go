package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestTaxRateDecimalMigrationRequiresAnExplicitConfiguredValue(t *testing.T) {
	upSQL, err := os.ReadFile("406_require_explicit_tax_rate_decimal.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("406_require_explicit_tax_rate_decimal.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	if !strings.Contains(up, "alter column rate_decimal drop default") {
		t.Fatal("up migration must remove the implicit zero tax-rate default")
	}
	if strings.Contains(up, "drop not null") || strings.Contains(up, "drop constraint") {
		t.Fatal("up migration must keep the tax-rate field and its validation constraints")
	}

	down := strings.ToLower(string(downSQL))
	if !strings.Contains(down, "alter column rate_decimal set default 0") {
		t.Fatal("down migration must restore the prior database default")
	}
}
