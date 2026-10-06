package migrations_test

import (
	"strings"
	"testing"
)

func TestStainlessSteelSpokeDislocationMechanicsFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "401_add_stainless_steel_spoke_dislocation_mechanics_faq_content.up.sql"))
	for _, required := range []string{
		"GUIDES-SPOKEGUIDES-STAINLESS-STEEL-MICROSTRUCTURAL-DISLOCATION-MECHANICS",
		"/GUIDES/SPOKEGUIDES/STAINLESS-STEEL-MICROSTRUCTURAL-DISLOCATION-MECHANICS",
		"INSERT INTO FAQ_PAGES",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"AISI 304 MATERIAL CURVE",
		"180 KGF REFERENCE",
		"COLD-WORK HARDENING",
		"TENSILE-TEST FAILURE ELONGATION",
		"CURVE ENDPOINT",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("stainless-steel spoke dislocation mechanics FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 7 {
		t.Fatalf("expected seven English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 7 {
		t.Fatalf("expected seven Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "401_add_stainless_steel_spoke_dislocation_mechanics_faq_content.down.sql")))
	for _, required := range []string{
		"USING SEED_FAQS",
		"EXISTING.QUESTION = SEED_FAQS.QUESTION",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("stainless-steel spoke dislocation mechanics FAQ rollback is missing safe fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("stainless-steel spoke dislocation mechanics FAQ rollback must preserve the route-owned FAQ page shell")
	}
}
