package migrations_test

import (
	"strings"
	"testing"
)

func TestFitmentFrameEntriesMigrationIsIndependent(t *testing.T) {
	upSQL := readMigrationFile(t, "220_create_fitment_frame_entries.up.sql")
	downSQL := readMigrationFile(t, "220_create_fitment_frame_entries.down.sql")
	lowerSQL := strings.ToLower(upSQL)

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS fitment_frame_entries",
		"brand_name VARCHAR(160) NOT NULL",
		"model_name VARCHAR(160) NOT NULL",
		"year_mode VARCHAR(16) NOT NULL",
		"is_enabled BOOLEAN NOT NULL DEFAULT FALSE",
		"uk_fitment_frame_entries_identity",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("fitment frame migration is missing contract fragment %q", fragment)
		}
	}

	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS fitment_frame_entries") {
		t.Fatal("fitment frame down migration must only drop its own table")
	}

	for _, forbidden := range []string{
		"products",
		"product_variants",
		"product_id",
		"product_variant_id",
		"REFERENCES",
		"FOREIGN KEY",
	} {
		if strings.Contains(lowerSQL, strings.ToLower(forbidden)) {
			t.Fatalf("fitment frame migration must not reference catalog fragment %q", forbidden)
		}
	}
}
