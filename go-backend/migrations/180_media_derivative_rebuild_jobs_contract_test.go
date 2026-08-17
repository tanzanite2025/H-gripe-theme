package migrations_test

import (
	"strings"
	"testing"
)

func TestMediaDerivativeRebuildJobMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "180_media_derivative_rebuild_jobs.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS media_derivative_rebuild_jobs",
		"status IN ('pending', 'running', 'succeeded')",
		"cursor_asset_id BIGINT NOT NULL DEFAULT 0",
		"idx_media_derivative_rebuild_jobs_single_pending",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("media derivative rebuild migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "180_media_derivative_rebuild_jobs.down.sql")
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS media_derivative_rebuild_jobs") {
		t.Fatal("media derivative rebuild down migration does not drop the table")
	}
}
