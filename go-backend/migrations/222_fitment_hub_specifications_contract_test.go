package migrations_test

import (
	"strings"
	"testing"
)

func TestFitmentHubSpecificationsMigrationKeepsTheDomainIndependent(t *testing.T) {
	upSQL := readMigrationFile(t, "222_create_fitment_hub_specifications.up.sql")
	downSQL := readMigrationFile(t, "222_create_fitment_hub_specifications.down.sql")
	lowerSQL := strings.ToLower(upSQL)

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS fitment_hub_specifications",
		"spec_code VARCHAR(80) NOT NULL",
		"position VARCHAR(16) NOT NULL",
		"axle_type VARCHAR(32) NOT NULL",
		"axle_spacing_mm INTEGER NOT NULL",
		"CREATE TABLE IF NOT EXISTS fitment_frame_hub_specifications",
		"PRIMARY KEY (frame_entry_id, hub_specification_id)",
		"uk_fitment_hub_specifications_code",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("fitment hub migration is missing contract fragment %q", fragment)
		}
	}

	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS fitment_frame_hub_specifications") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS fitment_hub_specifications") {
		t.Fatal("fitment hub down migration must remove only its own domain tables")
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"product_variant_id",
		"sku",
		"shipping",
	} {
		if strings.Contains(lowerSQL, strings.ToLower(forbidden)) {
			t.Fatalf("fitment hub migration must not reference unrelated domain fragment %q", forbidden)
		}
	}
}
