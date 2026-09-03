package migrations_test

import (
	"strings"
	"testing"
)

func TestSpokeRimBrandsProductBrandIDMigrationContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "236_spoke_rim_brands_product_brand_id.up.sql"))
	for _, fragment := range []string{
		"alter table spoke_rim_brands",
		"add column if not exists product_brand_id bigint null",
		"create unique index if not exists idx_spoke_rim_brands_product_brand_id",
		"on spoke_rim_brands(product_brand_id)",
		"where conname = 'fk_spoke_rim_brands_product_brand'",
		"add constraint fk_spoke_rim_brands_product_brand",
		"foreign key (product_brand_id) references product_brands (id)",
		"on delete restrict",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("spoke rim brands product brand id migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "236_spoke_rim_brands_product_brand_id.down.sql"))
	for _, fragment := range []string{
		"drop constraint if exists fk_spoke_rim_brands_product_brand",
		"drop index if exists idx_spoke_rim_brands_product_brand_id",
		"drop column if exists product_brand_id",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("spoke rim brands product brand id down migration is missing contract fragment %q", fragment)
		}
	}
}
