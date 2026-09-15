package migrations_test

import (
	"strings"
	"testing"
)

func TestShippingQuotePlansMigrationContract(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "267_shipping_quote_plans.up.sql"))
	downSQL := strings.ToUpper(readMigrationFile(t, "267_shipping_quote_plans.down.sql"))

	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS SHIPPING_QUOTE_SNAPSHOTS",
		"REQUEST_HASH CHAR(64) NOT NULL",
		"RATE_VERSION CHAR(64) NOT NULL",
		"QUOTE_DATA JSONB NOT NULL",
		"EXPIRES_AT TIMESTAMPTZ NOT NULL",
		"ADD COLUMN IF NOT EXISTS SHIPPING_QUOTE_ID VARCHAR(36)",
		"ADD COLUMN IF NOT EXISTS SHIPPING_QUOTE_PLAN_ID VARCHAR(36)",
		"ADD COLUMN IF NOT EXISTS SHIPPING_PLAN_SNAPSHOT JSONB NOT NULL",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}

	dropIndex := strings.Index(downSQL, "DROP INDEX IF EXISTS IDX_ORDERS_SHIPPING_QUOTE_ID")
	dropOrderColumns := strings.Index(downSQL, "ALTER TABLE ORDERS")
	dropSnapshots := strings.Index(downSQL, "DROP TABLE IF EXISTS SHIPPING_QUOTE_SNAPSHOTS")
	if dropIndex < 0 || dropOrderColumns < 0 || dropSnapshots < 0 {
		t.Fatal("down migration must remove the quote index, order columns, and snapshot table")
	}
	if !(dropIndex < dropOrderColumns && dropOrderColumns < dropSnapshots) {
		t.Fatal("down migration must remove dependent indexes and order columns before the snapshot table")
	}
}
