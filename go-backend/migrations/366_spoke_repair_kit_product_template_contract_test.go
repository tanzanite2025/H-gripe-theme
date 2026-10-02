package migrations_test

import (
	"strings"
	"testing"
)

func TestSpokeRepairKitProductTemplateMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "366_add_spoke_repair_kit_product_template.up.sql"))
	for _, required := range []string{
		"SPOKE REPAIR KIT",
		"SPOKE_REPAIR_KIT",
		"INSERT INTO PRODUCT_SPECIFICATION_TEMPLATES",
		"INSERT INTO PRODUCT_SPEC_DEFINITIONS",
		"SPOKE_MODEL",
		"HEAD_TYPE",
		"SPOKE_LENGTH",
		"SPOKE_DIAMETER",
		"NIPPLE_MODEL",
		"NIPPLE_LENGTH",
		"PACK_QUANTITY",
		"ROLE",
		"ATTRIBUTE",
		"INSERT INTO PRODUCT_SPEC_OPTION_ITEMS",
		"STRAIGHT_PULL",
		"J_BEND",
		"ON CONFLICT (SLUG) DO NOTHING",
		"NOT EXISTS",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("spoke repair-kit template migration is missing contract fragment %q", required)
		}
	}

	if strings.Contains(up, "INSERT INTO PRODUCTS") || strings.Contains(up, "INSERT INTO PRODUCT_CATEGORIES") {
		t.Fatal("spoke repair-kit template migration must not create products or categories")
	}
}

func TestSpokeRepairKitProductTemplateDownMigrationContract(t *testing.T) {
	down := strings.ToUpper(readMigrationFile(t, "366_add_spoke_repair_kit_product_template.down.sql"))
	for _, required := range []string{
		"DELETE FROM PRODUCT_SPECIFICATION_TEMPLATES",
		"SPOKE_REPAIR_KIT",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("spoke repair-kit template down migration is missing contract fragment %q", required)
		}
	}
}
