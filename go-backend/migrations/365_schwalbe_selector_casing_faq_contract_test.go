package migrations_test

import (
	"strings"
	"testing"
)

func TestSchwalbeSelectorCasingFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "365_add_schwalbe_selector_casing_faq.up.sql"))
	for _, required := range []string{
		"GUIDES-SCHWALBE-TIRE-SELECTOR",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"VERSION_LABEL",
		"SUPER RACE",
		"SUPER GROUND",
		"SUPER TRAIL",
		"SUPER DOWNHILL",
		"TRAIL PRO",
		"GRAVITY PRO",
		"GRAVITY PRO, RADIAL",
		"EPI",
		"LOAD CAPACITY",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("Schwalbe casing FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 1 {
		t.Fatalf("expected one English casing FAQ seed, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 1 {
		t.Fatalf("expected one Chinese casing FAQ seed, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "365_add_schwalbe_selector_casing_faq.down.sql")))
	for _, required := range []string{
		"USING SEED_FAQS",
		"EXISTING.QUESTION = SEED_FAQS.QUESTION",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("Schwalbe casing FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("casing FAQ down migration must not remove the FAQ page shell")
	}
}
