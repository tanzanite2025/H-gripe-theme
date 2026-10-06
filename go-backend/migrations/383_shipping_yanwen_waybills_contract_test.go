package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestYanwenWaybillMigrationStoresSuccessfulCreateSnapshotsAndIdempotencyKey(t *testing.T) {
	contents, err := os.ReadFile("383_create_shipping_yanwen_waybills.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(contents))
	for _, required := range []string{
		"create table if not exists shipping_yanwen_waybills",
		"request_data jsonb not null",
		"response_data jsonb not null",
		"unique (order_id, product_code)",
		"unique (waybill_number)",
	} {
		if !strings.Contains(lower, required) {
			t.Fatalf("Yanwen waybill migration is missing %q", required)
		}
	}
}

func TestYanwenWaybillMigrationCanBeRolledBack(t *testing.T) {
	contents, err := os.ReadFile("383_create_shipping_yanwen_waybills.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(contents)), "drop table if exists shipping_yanwen_waybills") {
		t.Fatal("Yanwen waybill migration rollback is missing table drop")
	}
}
