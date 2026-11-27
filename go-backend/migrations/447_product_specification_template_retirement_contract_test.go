package migrations_test

import (
	"strings"
	"testing"
)

func TestProductSpecificationTemplateRetirementMigrationDisablesEveryTemplateWritePath(t *testing.T) {
	up := readMigrationFile(t, "447_disable_product_specification_templates_and_preserve_manual_sku_variants.up.sql")
	for _, required := range []string{
		"product_option_value_variant_rules",
		"product_custom_option_policies",
		"product_media WHERE variant_option_value_id IS NOT NULL",
		"product_variants WHERE option_values <> '{}'::jsonb",
		"quick_buy_step_product_specification_templates",
		"quick_buy_step_filters WHERE spec_definition_id IS NOT NULL",
		"product_specification_templates_disabled_check CHECK (FALSE)",
		"product_spec_definitions_disabled_check CHECK (FALSE)",
		"product_spec_values_disabled_check CHECK (FALSE)",
		"product_variant_option_values_disabled_check CHECK (FALSE)",
		"product_option_value_relations_disabled_check CHECK (FALSE)",
		"product_media_no_template_option_check CHECK (variant_option_value_id IS NULL)",
		"product_variants_empty_template_options_check CHECK (option_values = '{}'::jsonb)",
		"WHERE deleted_at IS NULL AND option_values <> '{}'::jsonb",
	} {
		if !strings.Contains(up, required) {
			t.Errorf("migration 447 is missing retirement contract %q", required)
		}
	}
	if strings.Contains(up, "product_specification_template_translations") {
		t.Fatal("migration 447 references the removed product specification translation table")
	}
}

func TestProductSpecificationTemplateRetirementMigrationCannotBeRolledBack(t *testing.T) {
	down := readMigrationFile(t, "447_disable_product_specification_templates_and_preserve_manual_sku_variants.down.sql")
	if !strings.Contains(down, "RAISE EXCEPTION") || !strings.Contains(down, "irreversible") {
		t.Fatal("migration 447 down script must refuse to reopen the retired template system")
	}
}
