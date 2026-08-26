package migrations_test

import (
	"strings"
	"testing"
)

func TestFitmentForkEntriesMigrationKeepsTheDomainIndependent(t *testing.T) {
	upSQL := readMigrationFile(t, "223_create_fitment_fork_entries.up.sql")
	downSQL := readMigrationFile(t, "223_create_fitment_fork_entries.down.sql")
	lowerSQL := strings.ToLower(upSQL)

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS fitment_fork_entries",
		"brand_name VARCHAR(160) NOT NULL",
		"model_name VARCHAR(160) NOT NULL",
		"year_mode VARCHAR(16) NOT NULL",
		"is_enabled BOOLEAN NOT NULL DEFAULT FALSE",
		"uk_fitment_fork_entries_identity",
		"CREATE TABLE IF NOT EXISTS fitment_fork_hub_specifications",
		"PRIMARY KEY (fork_entry_id, hub_specification_id)",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("fitment fork migration is missing contract fragment %q", fragment)
		}
	}

	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS fitment_fork_hub_specifications") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS fitment_fork_entries") {
		t.Fatal("fitment fork down migration must remove only its own domain tables")
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"product_variant_id",
		"sku",
		"shipping",
		"packaging",
	} {
		if strings.Contains(lowerSQL, strings.ToLower(forbidden)) {
			t.Fatalf("fitment fork migration must not reference unrelated domain fragment %q", forbidden)
		}
	}
}
