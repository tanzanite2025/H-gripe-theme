package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestYanwenChannelCustomsDeclarationRequirementsMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("397_add_yanwen_channel_customs_declaration_requirements.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("397_add_yanwen_channel_customs_declaration_requirements.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"alter table shipping_yanwen_published_channels",
		"add column if not exists require_receiver_tax_number boolean not null default false",
		"add column if not exists require_ioss boolean not null default false",
		"add column if not exists require_eori boolean not null default false",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}

	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"alter table shipping_yanwen_published_channels",
		"drop column if exists require_receiver_tax_number",
		"drop column if exists require_ioss",
		"drop column if exists require_eori",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
