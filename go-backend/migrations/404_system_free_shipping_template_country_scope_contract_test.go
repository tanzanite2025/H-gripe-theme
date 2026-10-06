package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestSystemFreeShippingTemplateCountryScopeMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("404_add_system_free_shipping_template_country_scope.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("404_add_system_free_shipping_template_country_scope.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"add column if not exists template_kind varchar(40)",
		"add column if not exists is_system_managed boolean",
		"add column if not exists free_shipping_countries text not null default '[]'",
		"create unique index if not exists idx_shipping_templates_single_system_free_shipping",
		"insert into shipping_templates",
		"'system_free_shipping'",
		"'free_shipping'",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}

	for _, forbidden := range []string{
		"shipping_carrier_services",
		"shipping_fpx_channels",
		"shipping_yanwen_published_channels",
		"fpx_channel_id",
		"yanwen_published_channel_id",
	} {
		if strings.Contains(up, forbidden) {
			t.Fatalf("system free-shipping migration must not depend on carrier collection field/table %q", forbidden)
		}
	}

	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"delete from shipping_templates",
		"where template_kind = 'system_free_shipping'",
		"drop column if exists free_shipping_countries",
		"drop column if exists is_system_managed",
		"drop column if exists template_kind",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
