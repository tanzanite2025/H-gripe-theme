package migrations_test

import (
	"strings"
	"testing"
)

func TestSiteQualityLighthouseRunnerMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "172_site_quality_lighthouse_runner.up.sql")
	for _, fragment := range []string{
		"natively internal Lighthouse Runner based",
		"no compatibility rename",
		"SELECT 1",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("site quality runner migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "172_site_quality_lighthouse_runner.down.sql")
	if !strings.Contains(downSQL, "SELECT 1") {
		t.Fatal("site quality runner down migration must remain a no-op")
	}
}
