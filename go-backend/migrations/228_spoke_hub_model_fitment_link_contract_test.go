package migrations_test

import (
	"strings"
	"testing"
)

func TestSpokeHubModelFitmentLinkMigrationKeepsProjectionTraceable(t *testing.T) {
	upSQL := readMigrationFile(t, "228_link_spoke_hub_models_to_fitment_specifications.up.sql")
	downSQL := readMigrationFile(t, "228_link_spoke_hub_models_to_fitment_specifications.down.sql")
	lowerSQL := strings.ToLower(upSQL)

	for _, fragment := range []string{
		"alter table spoke_hub_models",
		"add column if not exists fitment_hub_specification_id bigint",
		"update spoke_hub_models as hub_model",
		"set fitment_hub_specification_id = specification.id",
		"lower(btrim(hub_model.code)) = lower(btrim(specification.spec_code))",
		"fk_spoke_hub_models_fitment_hub_specification",
		"references fitment_hub_specifications(id)",
		"on delete set null",
		"idx_spoke_hub_models_fitment_specification_id",
	} {
		if !strings.Contains(lowerSQL, fragment) {
			t.Fatalf("spoke hub projection link migration is missing contract fragment %q", fragment)
		}
	}

	for _, fragment := range []string{
		"drop constraint if exists fk_spoke_hub_models_fitment_hub_specification",
		"drop index if exists idx_spoke_hub_models_fitment_specification_id",
		"drop column if exists fitment_hub_specification_id",
	} {
		if !strings.Contains(strings.ToLower(downSQL), fragment) {
			t.Fatalf("spoke hub projection link down migration is missing contract fragment %q", fragment)
		}
	}
}
