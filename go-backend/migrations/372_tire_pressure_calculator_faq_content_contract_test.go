package migrations_test

import (
	"strings"
	"testing"
)

func TestTirePressureCalculatorFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "372_add_tire_pressure_calculator_faq_content.up.sql"))
	for _, required := range []string{
		"GUIDES-TIREGUIDES-TIRE-PRESSURE-CALCULATOR",
		"/GUIDES/TIREGUIDES/TIRE-PRESSURE-CALCULATOR",
		"INSERT INTO FAQ_PAGES",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"A = FZ / P",
		"1 MM WATER-FILM",
		"MINUS 5 OR MINUS 7 PSI",
		"NORMALIZED TO 1",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("tire-pressure calculator FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 7 {
		t.Fatalf("expected seven English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 7 {
		t.Fatalf("expected seven Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "372_add_tire_pressure_calculator_faq_content.down.sql")))
	for _, required := range []string{
		"WITH SEED_FAQS",
		"SEED_FAQS.ANSWER = FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("tire-pressure calculator FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("tire-pressure calculator FAQ down migration must preserve the route-owned FAQ page shell")
	}
}
