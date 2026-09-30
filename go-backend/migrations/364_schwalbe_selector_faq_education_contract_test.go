package migrations_test

import (
	"strings"
	"testing"
)

func TestSchwalbeSelectorEducationFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "364_add_schwalbe_selector_faq_education.up.sql"))
	for _, required := range []string{
		"GUIDES-SCHWALBE-TIRE-SELECTOR",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"LOAD_KG",
		"RIDER-WEIGHT",
		"ENDS PER INCH",
		"REFLEX",
		"BLACKREFLEX",
		"WHITEWALL",
		"GUMWALL",
		"EPI",
		"GRAVITY PRO MAGIC MARY",
		"TRAIL PRO NOBBY NIC",
		"PRO ONE AERO FRONT",
		"G-ONE RX PRO",
		"BRAKE_TYPE",
		"BSD",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("Schwalbe selector education FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 5 {
		t.Fatalf("expected five English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 5 {
		t.Fatalf("expected five Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "364_add_schwalbe_selector_faq_education.down.sql")))
	for _, required := range []string{
		"USING SEED_FAQS",
		"EXISTING.QUESTION = SEED_FAQS.QUESTION",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("Schwalbe selector education FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("education FAQ down migration must not remove the FAQ page shell")
	}
}
