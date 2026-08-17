package migrations_test

import (
	"strings"
	"testing"
)

func TestSiteQualityScheduledJobsRestorationMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "175_site_quality_restore_scheduled_jobs.up.sql")
	for _, fragment := range []string{
		"WHERE kind = 'scheduled'",
		"status = 'queued'",
		"lease_expires_at = NULL",
		"CHECK (status IN ('queued', 'processing', 'succeeded', 'failed', 'dead_letter'))",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("scheduled jobs restoration migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "175_site_quality_restore_scheduled_jobs.down.sql")
	for _, fragment := range []string{
		"status = 'cancelled'",
		"automatic Site Quality scheduling retired",
		"cancelled",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("scheduled jobs restoration down migration is missing contract fragment %q", fragment)
		}
	}
}
