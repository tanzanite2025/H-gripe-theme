package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveOrderTrackingProjectionMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "299_remove_order_tracking_projection.up.sql"))
	if !strings.Contains(up, "ALTER TABLE ORDERS") {
		t.Fatal("up migration must alter orders")
	}
	for _, column := range []string{
		"TRACKING_NUMBER",
		"TRACKING_PROVIDER_ID",
		"CARRIER_ID",
		"CARRIER_SERVICE_ID",
		"TRACKING_CARRIER_MAPPING_ID",
		"PROVIDER_CARRIER_CODE",
		"PROVIDER_CARRIER_NAME",
	} {
		if !strings.Contains(up, "DROP COLUMN IF EXISTS "+column) {
			t.Fatalf("up migration must drop orders.%s", strings.ToLower(column))
		}
	}
	if strings.TrimSpace(readMigrationFile(t, "299_remove_order_tracking_projection.down.sql")) == "" {
		t.Fatal("down migration must document intentional restoration")
	}
}
