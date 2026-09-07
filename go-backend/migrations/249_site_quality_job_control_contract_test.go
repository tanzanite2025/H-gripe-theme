package migrations_test

import (
	"strings"
	"testing"
)

func TestSiteQualityJobControlMigrationDefinesCancellationAndProgress(t *testing.T) {
	upSQL := strings.ToLower(readSiteQualityMigration(t, "249_site_quality_job_control.up.sql"))
	for _, fragment := range []string{
		"add column if not exists progress_total integer",
		"add column if not exists completed_samples integer",
		"add column if not exists progress_stage varchar(32)",
		"status in ('queued', 'processing', 'succeeded', 'failed', 'dead_letter', 'cancelled')",
		"progress_total >= 0",
		"idx_site_quality_jobs_progress",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("site quality job control migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readSiteQualityMigration(t, "249_site_quality_job_control.down.sql"))
	for _, fragment := range []string{
		"status = 'failed'",
		"drop column if exists progress_stage",
		"drop column if exists completed_samples",
		"drop column if exists progress_total",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("site quality job control down migration is missing contract fragment %q", fragment)
		}
	}
}
