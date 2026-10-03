package migrations_test

import (
	"strings"
	"testing"
)

func TestSpokeRepairKitProductCategoryMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "367_seed_spoke_repair_kit_product_category.up.sql"))
	for _, required := range []string{
		"WHEEL-COMPONENTS",
		"SPOKE-REPAIR-KITS",
		"SPOKE REPAIR KITS",
		"辐条修补件",
		"PARENT_ID",
		"DEPTH",
		"PRODUCT_CATEGORY_TRANSLATIONS",
		"'ZH_CN'",
		"ON CONFLICT (PRODUCT_CATEGORY_ID, LOCALE) DO UPDATE",
	} {
		if !strings.Contains(up, strings.ToUpper(required)) {
			t.Fatalf("spoke repair-kit category migration is missing contract fragment %q", required)
		}
	}

	if strings.Contains(up, "INSERT INTO PRODUCTS") || strings.Contains(up, "PRODUCT_SPECIFICATION_TEMPLATES") {
		t.Fatal("spoke repair-kit category migration must not create products or modify the specification template")
	}
}

func TestSpokeRepairKitProductCategoryDownMigrationContract(t *testing.T) {
	down := strings.ToUpper(readMigrationFile(t, "367_seed_spoke_repair_kit_product_category.down.sql"))
	if !strings.Contains(down, "DELETE FROM PRODUCT_CATEGORIES") || !strings.Contains(down, "SPOKE-REPAIR-KITS") {
		t.Fatal("spoke repair-kit category down migration must remove only its category")
	}
}
