package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCustomsTradeRemedyRiskFieldsMigrationAddsStructuredRiskDataAndWheelsetSeed(t *testing.T) {
	contents, err := os.ReadFile("396_add_customs_trade_remedy_risk_fields.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"add column if not exists trade_remedy_risk_level varchar(16) not null default 'none'",
		"add column if not exists trade_remedy_risk_tags_json jsonb not null default '[]'::jsonb",
		"add column if not exists trade_remedy_declaration_advice text not null default ''",
		"slug = 'bicycle-wheelset'",
		"trade_remedy_risk_level = 'high'",
		"eu_anti_dumping_attention",
		"eu_complete_wheelset_anti_circumvention",
		"us_section_301_list_3_review",
		"ec 88/97",
		"hts 9903.88.03",
		"不得为规避贸易救济措施虚假拆单",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs trade-remedy risk migration is missing %q", requiredFragment)
		}
	}
}

func TestCustomsTradeRemedyRiskFieldsMigrationRollbackRemovesOnlyAddedColumns(t *testing.T) {
	contents, err := os.ReadFile("396_add_customs_trade_remedy_risk_fields.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"drop column if exists trade_remedy_risk_level",
		"drop column if exists trade_remedy_risk_tags_json",
		"drop column if exists trade_remedy_declaration_advice",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs trade-remedy risk rollback is missing %q", requiredFragment)
		}
	}
	if strings.Contains(sql, "drop table customs_classification_profiles") {
		t.Fatal("customs trade-remedy risk rollback must not drop the customs profile table")
	}
}
