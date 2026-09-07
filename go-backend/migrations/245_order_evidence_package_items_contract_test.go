package migrations_test

import (
	"strings"
	"testing"
)

func TestOrderEvidencePackageItemsMigrationDefinesScopedEvidenceContracts(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "245_order_evidence_package_items.up.sql"))
	downSQL := strings.ToLower(readMigrationFile(t, "245_order_evidence_package_items.down.sql"))

	for _, fragment := range []string{
		"create table if not exists order_evidence_packages",
		"snapshot_id bigint not null references order_evidence_snapshots(id)",
		"unique (order_id, package_version)",
		"status in ('incomplete', 'ready', 'locked', 'superseded')",
		"create table if not exists order_evidence_items",
		"package_id bigint not null references order_evidence_packages(id)",
		"order_item_id bigint references order_items(id)",
		"snapshot_id bigint references order_evidence_snapshots(id)",
		"configuration_confirmation",
		"outbound_weight_packaging",
		"spoke_qc_tension",
		"create table if not exists order_evidence_attachments",
		"storage_key text not null",
		"sha256 char(64) not null",
		"uq_order_evidence_attachment_item_storage_key",
		"uq_order_evidence_item_order_scope",
		"where order_item_id is null",
		"uq_order_evidence_item_line_scope",
		"where order_item_id is not null",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("order evidence package/items migration is missing contract fragment %q", fragment)
		}
	}
	for _, fragment := range []string{
		"drop table if exists order_evidence_attachments",
		"drop table if exists order_evidence_items",
		"drop table if exists order_evidence_packages",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("order evidence package/items down migration is missing %q", fragment)
		}
	}
}
