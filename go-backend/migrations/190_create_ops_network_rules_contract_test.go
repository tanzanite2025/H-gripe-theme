package migrations_test

import (
	"strings"
	"testing"
)

func TestCreateOpsNetworkRulesMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "190_create_ops_network_rules.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS ops_network_rules",
		"vps_binding_id BIGINT NULL",
		"project_binding_id BIGINT NULL",
		"domain_binding_id BIGINT NULL",
		"connector_id BIGINT NULL",
		"fk_ops_network_rule_vps",
		"fk_ops_network_rule_project",
		"fk_ops_network_rule_domain",
		"fk_ops_network_rule_connector",
		"ON DELETE SET NULL",
		"idx_ops_network_rule_owner",
		"idx_ops_network_rule_manager",
		"idx_ops_network_rule_scope",
		"'Hostinger production ingress'",
		"'shared-edge to theme-web'",
		"'unknown'",
		"'pending'",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("network rules up migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readMigrationFile(t, "190_create_ops_network_rules.down.sql")
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS idx_ops_network_rule_connector",
		"DROP TABLE IF EXISTS ops_network_rules",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("network rules down migration is missing contract fragment %q", fragment)
		}
	}
}
