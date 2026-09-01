package migrations_test

import (
	"strings"
	"testing"
)

func TestTransactionsLiabilityShiftedMigrationContract(t *testing.T) {
	upSQL := strings.ToLower(readMigrationFile(t, "233_transactions_liability_shifted.up.sql"))
	for _, fragment := range []string{
		"add column if not exists liability_shifted boolean null",
		"idx_transactions_liability_shifted",
		"where liability_shifted is not null",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("transactions liability shifted migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := strings.ToLower(readMigrationFile(t, "233_transactions_liability_shifted.down.sql"))
	for _, fragment := range []string{
		"drop index if exists idx_transactions_liability_shifted",
		"drop column if exists liability_shifted",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("transactions liability shifted down migration is missing contract fragment %q", fragment)
		}
	}
}
