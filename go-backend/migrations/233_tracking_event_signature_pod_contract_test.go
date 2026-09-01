package migrations_test

import (
	"strings"
	"testing"
)

func TestTrackingEventSignaturePODMigrationContract(t *testing.T) {
	upSQL := readMigrationFile(t, "233_tracking_event_signature_pod.up.sql")
	downSQL := readMigrationFile(t, "233_tracking_event_signature_pod.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE tracking_events",
		"ADD COLUMN IF NOT EXISTS recipient_signature_name VARCHAR(160)",
		"ADD COLUMN IF NOT EXISTS proof_of_delivery_url TEXT",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("tracking event Signature POD migration is missing contract fragment %q", fragment)
		}
	}

	for _, fragment := range []string{
		"DROP COLUMN IF EXISTS proof_of_delivery_url",
		"DROP COLUMN IF EXISTS recipient_signature_name",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("tracking event Signature POD rollback is missing contract fragment %q", fragment)
		}
	}
}
