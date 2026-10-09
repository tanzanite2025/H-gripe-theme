package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestTaxRateSourceSnapshotMigrationKeepsReferenceRatesSeparateFromCheckout(t *testing.T) {
	up, err := os.ReadFile("405_create_tax_rate_source_snapshots.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	down, err := os.ReadFile("405_create_tax_rate_source_snapshots.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	upSQL := string(up)
	downSQL := string(down)

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS tax_rate_source_configs",
		"provider_code = 'vatcomply'",
		"refresh_interval_hours BETWEEN 24 AND 8760",
		"CREATE TABLE IF NOT EXISTS tax_rate_source_snapshots",
		"content_sha256 CHAR(64) NOT NULL",
		"CREATE TABLE IF NOT EXISTS tax_rate_source_snapshot_entries",
		"REFERENCES tax_rate_source_snapshots(id) ON DELETE CASCADE",
		"rate_decimal NUMERIC(12, 8)",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("tax-rate source snapshot migration is missing %q", required)
		}
	}

	for _, forbidden := range []string{
		"INSERT INTO tax_rates",
		"UPDATE tax_rates",
		"ALTER TABLE tax_rates",
		"shipping_templates",
	} {
		if strings.Contains(upSQL, forbidden) {
			t.Fatalf("tax-rate source snapshot migration must not modify checkout or shipping data via %q", forbidden)
		}
	}

	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS tax_rate_source_snapshot_entries") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS tax_rate_source_snapshots") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS tax_rate_source_configs") {
		t.Fatal("tax-rate source snapshot rollback must drop only its own tables")
	}
}
