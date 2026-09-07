package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderEvidenceSnapshotsMigrationDefinesImmutableOrderTimeContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "244_order_evidence_snapshots.up.sql"))
	downSQL := strings.ToLower(readMigrationFile(t, "244_order_evidence_snapshots.down.sql"))

	for _, fragment := range []string{
		"alter table order_items",
		"add column if not exists weight_grams integer not null default 0",
		"create table if not exists order_evidence_snapshots",
		"order_id bigint not null unique references orders(id) on delete restrict",
		"schema_version integer not null",
		"confirmed_at timestamptz not null",
		"order_total_amount numeric(14,2) not null",
		"order_total_usd numeric(14,2) not null",
		"has_spoke_tension_qc boolean not null default false",
		"snapshot_data jsonb not null",
		"snapshot_sha256 char(64) not null",
		"check (schema_version = 1)",
		"idx_order_evidence_snapshots_high_value",
		"idx_order_evidence_snapshots_spoke_tension_qc",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("order evidence snapshot migration is missing contract fragment %q", fragment)
		}
	}
	for _, fragment := range []string{
		"drop table if exists order_evidence_snapshots",
		"drop column if exists weight_grams",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("order evidence snapshot down migration is missing %q", fragment)
		}
	}
}
