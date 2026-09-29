package migrations_test

import (
	"strings"
	"testing"
)

func TestSchwalbeTireRimWidthCombinationRulesMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "362_add_schwalbe_tire_rim_width_combination_rules.up.sql"))
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS SCHWALBE_TIRE_RIM_WIDTH_COMBINATION_RULES",
		"TIRE_WIDTH_MIN_MM INTEGER NOT NULL",
		"TIRE_WIDTH_MAX_MM INTEGER NOT NULL",
		"INNER_RIM_WIDTH_MIN_MM INTEGER NOT NULL",
		"INNER_RIM_WIDTH_MAX_MM INTEGER NOT NULL",
		"SOURCE_BASIS TEXT NOT NULL",
		"SOURCE_VERSION VARCHAR(64) NOT NULL",
		"SOURCE_URL TEXT NOT NULL",
		"SOURCE_CHECKED_AT DATE NOT NULL",
		"CHK_SCHWALBE_TIRE_RIM_WIDTH_RULE_POSITIVE_RANGES",
		"CHK_SCHWALBE_TIRE_RIM_WIDTH_RULE_RANGE_ORDER",
		"UQ_SCHWALBE_TIRE_RIM_WIDTH_COMBINATION_RULE",
		"ON CONFLICT (",
		"DO UPDATE SET",
		"COMMENT ON TABLE SCHWALBE_TIRE_RIM_WIDTH_COMBINATION_RULES",
		"POSSIBLE COMBINATIONS",
		"NOT A MODEL-SPECIFIC COMPATIBILITY CERTIFICATION",
		"FRAME-CLEARANCE CHECKS",
		"HOOKLESS",
		"STRAIGHT-SIDE",
		"TLE/TLR",
		"RIM MANUFACTURER",
		"ETRTO STANDARD 2024",
		"05/2024",
		"REIFEN_FELGENKOMBINATION_ETRTO_EN_(2).PDF",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("rim-width combination migration is missing contract fragment %q", required)
		}
	}

	for _, rangeRow := range []string{
		"(20, 21, 15, 17,",
		"(22, 24, 15, 20,",
		"(25, 27, 15, 22,",
		"(28, 28, 16, 23,",
		"(29, 34, 16, 25,",
		"(35, 46, 17, 27,",
		"(47, 57, 17, 30,",
		"(58, 65, 21, 35,",
		"(66, 71, 25, 43,",
		"(72, 83, 31, 53,",
		"(84, 95, 41, 64,",
		"(96, 113, 59, 89,",
		"(114, 132, 72, 100,",
	} {
		if !strings.Contains(up, rangeRow) {
			t.Fatalf("rim-width combination migration is missing official range %q", rangeRow)
		}
	}

	if count := strings.Count(up, "DATE '2026-09-28'"); count != 13 {
		t.Fatalf("expected 13 checked-at dates for the official matrix, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "362_add_schwalbe_tire_rim_width_combination_rules.down.sql")))
	if !strings.Contains(down, "FORWARD-ONLY") {
		t.Fatal("rim-width combination down migration must document the forward-only data policy")
	}
	if strings.Contains(down, "DROP TABLE") || strings.Contains(down, "DELETE FROM") {
		t.Fatal("rim-width combination down migration must not delete official selector rules")
	}
}
