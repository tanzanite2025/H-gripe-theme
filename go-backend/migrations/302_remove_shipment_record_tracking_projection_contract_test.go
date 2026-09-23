package migrations_test

import (
	"strings"
	"testing"
)

func TestRemoveShipmentRecordTrackingProjectionMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "302_remove_shipment_record_tracking_projection.up.sql"))
	for _, column := range []string{"TRACKING_SHIPMENT_ID", "TRACKING_NUMBER"} {
		if !strings.Contains(up, "DROP COLUMN IF EXISTS "+column) {
			t.Fatalf("up migration must drop shipment_records.%s", strings.ToLower(column))
		}
	}
	if strings.TrimSpace(readMigrationFile(t, "302_remove_shipment_record_tracking_projection.down.sql")) == "" {
		t.Fatal("down migration must be present")
	}
}
