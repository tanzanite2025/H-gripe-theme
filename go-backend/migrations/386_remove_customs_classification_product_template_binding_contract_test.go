package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCustomsClassificationDecouplingMigrationRemovesProductTemplateBinding(t *testing.T) {
	contents, err := os.ReadFile("386_remove_customs_classification_product_template_binding.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"drop column if exists product_specification_template_id",
		"drop index if exists idx_customs_classification_profiles_product_specification_template_id",
		"customs classification profiles are independent reusable master data",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs classification decoupling migration is missing %q", requiredFragment)
		}
	}
}

func TestCustomsClassificationDecouplingMigrationRollbackRestoresNullableBinding(t *testing.T) {
	contents, err := os.ReadFile("386_remove_customs_classification_product_template_binding.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"add column if not exists product_specification_template_id bigint",
		"references product_specification_templates(id)",
		"create index if not exists idx_customs_classification_profiles_product_specification_template_id",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs classification rollback is missing %q", requiredFragment)
		}
	}
}
