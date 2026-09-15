package migrations_test

import (
	"strings"
	"testing"
)

func TestAfterSalesActiveCaseOrderUniqueIndexUpMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "255_after_sales_active_case_order_unique_index.up.sql")
	for _, fragment := range []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_active_after_sales_case_per_order",
		"ON after_sales_cases(order_id)",
		"WHERE status NOT IN ('completed', 'rejected', 'cancelled')",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestAfterSalesActiveCaseOrderUniqueIndexDownMigrationContract(t *testing.T) {
	sql := readMigrationFile(t, "255_after_sales_active_case_order_unique_index.down.sql")
	if !strings.Contains(sql, "DROP INDEX IF EXISTS uq_active_after_sales_case_per_order") {
		t.Fatal("down migration is missing active after-sales index removal")
	}
}
