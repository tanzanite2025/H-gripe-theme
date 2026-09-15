package migrations_test

import (
	"strings"
	"testing"
)

func TestAfterSalesCaseScopeMigrationReplacesOrderWideLock(t *testing.T) {
	upSQL := strings.ToUpper(readMigrationFile(t, "269_after_sales_case_scope.up.sql"))
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS UQ_ACTIVE_AFTER_SALES_CASE_PER_ORDER",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_ACTIVE_CUSTOMER_AFTER_SALES_REQUEST_PER_ORDER",
		"ON AFTER_SALES_CASES(ORDER_ID)",
		"TYPE = 'CUSTOMER_REQUEST'",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToUpper(readMigrationFile(t, "269_after_sales_case_scope.down.sql"))
	for _, fragment := range []string{
		"DROP INDEX IF EXISTS UQ_ACTIVE_CUSTOMER_AFTER_SALES_REQUEST_PER_ORDER",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_ACTIVE_AFTER_SALES_CASE_PER_ORDER",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
