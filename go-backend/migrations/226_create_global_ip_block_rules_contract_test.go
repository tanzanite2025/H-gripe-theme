package migrations_test

import (
	"strings"
	"testing"
)

func TestCreateGlobalIPBlockRulesMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "226_create_global_ip_block_rules.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS global_ip_block_rules",
		"cidr VARCHAR(120) NOT NULL",
		"source VARCHAR(64) NOT NULL",
		"source_reference VARCHAR(160) NOT NULL",
		"expires_at TIMESTAMPTZ NULL",
		"enabled BOOLEAN NOT NULL DEFAULT TRUE",
		"idx_global_ip_block_rules_active_live",
		"ALTER TABLE visitor_profiles",
		"ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45)",
		"idx_visitor_profiles_ip_address_live",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("global IP block migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readMigrationFile(t, "226_create_global_ip_block_rules.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_global_ip_block_rules_active_live",
		"DROP TABLE IF EXISTS global_ip_block_rules",
		"DROP INDEX IF EXISTS idx_visitor_profiles_ip_address_live",
		"DROP COLUMN IF EXISTS ip_address",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("global IP block down migration is missing contract fragment %q", fragment)
		}
	}
}
