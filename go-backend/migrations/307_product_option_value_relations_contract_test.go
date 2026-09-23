package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestProductOptionValueRelationsMigrationContract(t *testing.T) {
	data, err := os.ReadFile("307_product_option_value_relations.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	for _, fragment := range []string{
		"create table if not exists product_option_value_relations",
		"product_id bigint not null references products(id)",
		"source_option_value_id bigint not null references product_variant_option_values(id)",
		"target_option_value_id bigint not null references product_variant_option_values(id)",
		"relation_type in ('requires', 'conflicts')",
		"source_option_value_id <> target_option_value_id",
		"unique (product_id, source_option_value_id, target_option_value_id, relation_type)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
