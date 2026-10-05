package migrations_test

import (
	"strings"
	"testing"
)

func TestMoveTirePressureCalculatorFAQToGuideMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "400_move_tire_pressure_calculator_faq_to_tire_pressure_guide.up.sql"))
	for _, required := range []string{
		"UPDATE FAQS",
		"UPDATE FAQ_PAGES",
		"GUIDES-TIREGUIDES-TIRE-PRESSURE-CALCULATOR",
		"GUIDES-TIREGUIDES-TIRE-PRESSURE",
		"/GUIDES/TIREGUIDES/TIRE-PRESSURE",
		"MANIFEST_VERSION",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("tire-pressure FAQ move migration is missing contract fragment %q", required)
		}
	}

	down := strings.ToUpper(readMigrationFile(t, "400_move_tire_pressure_calculator_faq_to_tire_pressure_guide.down.sql"))
	for _, required := range []string{
		"UPDATE FAQS",
		"UPDATE FAQ_PAGES",
		"/GUIDES/TIREGUIDES/TIRE-PRESSURE-CALCULATOR",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("tire-pressure FAQ rollback is missing contract fragment %q", required)
		}
	}
}
