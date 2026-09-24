package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration324RemovesTransactionLegacyMajorAmount(t *testing.T) {
	data, err := os.ReadFile("324_remove_transaction_legacy_major_amount.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(data))
	if !strings.Contains(sql, "alter table transactions") || !strings.Contains(sql, "drop column if exists amount") {
		t.Fatal("migration must remove transactions.amount")
	}
}
