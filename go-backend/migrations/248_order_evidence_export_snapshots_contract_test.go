package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderEvidenceExportSnapshotsMigrationDefinesImmutableManifestBoundary(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "248_order_evidence_export_snapshots.up.sql"))
	downSQL := strings.ToLower(readMigrationFile(t, "248_order_evidence_export_snapshots.down.sql"))

	for _, fragment := range []string{
		"create table if not exists order_evidence_export_snapshots",
		"order_id bigint not null references orders(id) on delete restrict",
		"evidence_package_id bigint not null references order_evidence_packages(id) on delete restrict",
		"evidence_package_version integer not null",
		"snapshot_data jsonb not null",
		"snapshot_sha256 char(64) not null",
		"check (status = 'locked')",
		"unique (evidence_package_id, version)",
		"idx_order_evidence_export_snapshots_order",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("order evidence export snapshot migration is missing contract fragment %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "drop table if exists order_evidence_export_snapshots") {
		t.Fatal("order evidence export snapshot down migration is missing table drop")
	}
}
