package migrations_test

import (
	"strings"
	"testing"
)

func TestFitmentHubSpokeGeometryMigrationAddsCompleteGeometryContract(t *testing.T) {
	upSQL := readMigrationFile(t, "227_add_fitment_hub_spoke_geometry.up.sql")
	downSQL := readMigrationFile(t, "227_add_fitment_hub_spoke_geometry.down.sql")
	lowerSQL := strings.ToLower(upSQL)

	for _, fragment := range []string{
		"add column if not exists wr_mm double precision",
		"add column if not exists wl_mm double precision",
		"add column if not exists pcdr_mm double precision",
		"add column if not exists pcdl_mm double precision",
		"fitment_hub_specifications_spoke_geometry_complete_check",
		"wr_mm > 0",
		"wr_mm is not null",
		"wl_mm is not null",
		"pcdr_mm is not null",
		"pcdl_mm is not null",
		"pcdr_mm >= 10",
	} {
		if !strings.Contains(lowerSQL, fragment) {
			t.Fatalf("fitment hub spoke geometry migration is missing contract fragment %q", fragment)
		}
	}

	for _, column := range []string{"wr_mm", "wl_mm", "pcdr_mm", "pcdl_mm"} {
		if !strings.Contains(strings.ToLower(downSQL), "drop column if exists "+column) {
			t.Fatalf("fitment hub spoke geometry down migration must remove %s", column)
		}
	}
}
