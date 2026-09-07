package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderEvidenceSubmissionSnapshotsMigrationDefinesImmutableSubmissionBoundary(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "247_order_evidence_submission_snapshots.up.sql"))
	downSQL := strings.ToLower(readMigrationFile(t, "247_order_evidence_submission_snapshots.down.sql"))

	for _, fragment := range []string{
		"create table if not exists order_evidence_submission_snapshots",
		"provider varchar(32) not null",
		"dispute_id bigint not null",
		"order_id bigint not null references orders(id) on delete restrict",
		"evidence_package_id bigint references order_evidence_packages(id) on delete restrict",
		"evidence_package_version integer not null default 0",
		"version integer not null",
		"status varchar(16) not null default 'locked'",
		"locked_at timestamptz not null",
		"snapshot_data jsonb not null",
		"snapshot_sha256 char(64) not null",
		"check (status = 'locked')",
		"unique (provider, dispute_id, version)",
		"idx_order_evidence_submission_snapshots_dispute",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("order evidence submission snapshot migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "drop table if exists order_evidence_submission_snapshots") {
		t.Fatal("order evidence submission snapshot down migration is missing table drop")
	}
}
