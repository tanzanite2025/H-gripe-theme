package migrations_test

import (
	"strings"
	"testing"
)

func TestBrandWheelsetSpokeSpecsFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "371_add_brand_wheelset_spoke_specs_faq_content.up.sql"))
	for _, required := range []string{
		"RESOURCES-BRAND-WHEELSET-SPOKE-SPECS",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"ROTOR DISH",
		"SAPIM POLYAX",
		"AEROLITE",
		"CX-RAY",
		"WH-404-FTLD-A1",
		"WH-404-FTLD-B1",
		"WH-404-FTLD-V2",
		"±1 MM",
		"9–10 MM",
		"8 MM",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("wheelset spoke specs FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 4 {
		t.Fatalf("expected four English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 4 {
		t.Fatalf("expected four Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "371_add_brand_wheelset_spoke_specs_faq_content.down.sql")))
	for _, required := range []string{
		"USING SEED_FAQS",
		"EXISTING.QUESTION = SEED_FAQS.QUESTION",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("wheelset spoke specs FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("wheelset spoke specs FAQ down migration must not remove the FAQ page shell")
	}
}
