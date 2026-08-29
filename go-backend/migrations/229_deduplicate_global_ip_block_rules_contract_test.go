package migrations_test

import (
	"strings"
	"testing"
)

func TestGlobalIPBlockRuleDeduplicationMigrationAddsActiveIdentityConstraint(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "229_deduplicate_global_ip_block_rules.up.sql"))
	for _, fragment := range []string{
		"row_number() over",
		"partition by source, source_reference, cidr",
		"duplicate_rank > 1",
		"uq_global_ip_block_rules_active_identity",
		"where deleted_at is null and enabled = true",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("global IP block deduplication migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "229_deduplicate_global_ip_block_rules.down.sql"))
	if !strings.Contains(downSQL, "drop index if exists uq_global_ip_block_rules_active_identity") {
		t.Fatal("global IP block deduplication down migration is missing unique index cleanup")
	}
}
