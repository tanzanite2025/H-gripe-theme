package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingYanwenWaybillEnvironmentIdempotencyMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("389_scope_yanwen_waybill_idempotency_to_environment.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("389_scope_yanwen_waybill_idempotency_to_environment.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"drop constraint if exists uq_shipping_yanwen_waybills_order_product",
		"drop index if exists idx_shipping_yanwen_waybill_order_channel",
		"create unique index if not exists idx_shipping_yanwen_waybill_environment_order_product",
		"on shipping_yanwen_waybills (environment, order_id, product_code)",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}

	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"cannot remove yanwen waybill environment idempotency while cross-environment duplicates exist",
		"drop index if exists idx_shipping_yanwen_waybill_environment_order_product",
		"unique (order_id, product_code)",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
