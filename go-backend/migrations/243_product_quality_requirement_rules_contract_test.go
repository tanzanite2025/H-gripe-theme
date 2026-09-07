package migrations_test

import (
	"strings"
	"testing"
)

func TestProductQualityRequirementRulesMigrationDefinesExplicitProductAndVariantScopes(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "243_product_quality_requirement_rules.up.sql"))
	downSQL := strings.ToLower(readMigrationFile(t, "243_product_quality_requirement_rules.down.sql"))

	for _, fragment := range []string{
		"create table if not exists product_quality_requirement_rules",
		"product_id bigint not null references products(id)",
		"variant_id bigint references product_variants(id)",
		"requirement_type varchar(64) not null",
		"spoke_tension_qc_required boolean not null default false",
		"status varchar(16) not null default 'active'",
		"rule_version varchar(64) not null",
		"check (requirement_type = 'spoke_tension_qc')",
		"check (status in ('active', 'inactive'))",
		"uq_product_quality_requirement_product_default",
		"where variant_id is null",
		"uq_product_quality_requirement_variant",
		"where variant_id is not null",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("product quality requirement migration is missing contract fragment %q", fragment)
		}
	}
	for _, fragment := range []string{
		"drop index if exists uq_product_quality_requirement_variant",
		"drop index if exists uq_product_quality_requirement_product_default",
		"drop table if exists product_quality_requirement_rules",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("product quality requirement down migration is missing %q", fragment)
		}
	}
}
