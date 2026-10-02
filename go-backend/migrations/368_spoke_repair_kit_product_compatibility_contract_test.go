package migrations_test

import (
	"strings"
	"testing"
)

func TestSpokeRepairKitProductCompatibilityMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "368_spoke_repair_kit_product_compatibility.up.sql"))
	for _, required := range []string{
		"PRODUCT_SPOKE_REPAIR_KIT_MODELS",
		"PRODUCT_ID",
		"BRAND_SLUG",
		"WHEELSET_MODEL_SLUG",
		"WHEELSET_MODEL_NAME",
		"LIFECYCLE_STATUS",
		"UNIQUE (PRODUCT_ID, BRAND_SLUG, WHEELSET_MODEL_SLUG)",
		"ON DELETE CASCADE",
		"DELETE FROM PRODUCT_SPECIFICATION_TEMPLATES",
		"SPOKE_REPAIR_KIT",
		"NOT EXISTS (\n          SELECT 1\n          FROM PRODUCTS",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("spoke repair-kit compatibility migration is missing contract fragment %q", required)
		}
	}
	if strings.Contains(up, "INSERT INTO PRODUCT_SPEC_DEFINITIONS") {
		t.Fatal("repair-kit compatibility migration must not create generic specification fields")
	}
}
