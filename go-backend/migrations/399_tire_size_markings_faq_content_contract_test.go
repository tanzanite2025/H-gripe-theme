package migrations_test

import (
	"strings"
	"testing"
)

func TestTireSizeMarkingsFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "399_add_tire_size_markings_faq_content.up.sql"))
	for _, required := range []string{
		"GUIDES-TIREGUIDES-TIRE-SIZE-MARKINGS",
		"/GUIDES/TIREGUIDES/TIRE-SIZE-MARKINGS",
		"INSERT INTO FAQ_PAGES",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"BEAD SEAT DIAMETER",
		"700 X 35C",
		"28 INCH, 700C, AND 29 INCH",
		"FRAME OR FORK CLEARANCE",
		"SIDEWALL MARKINGS",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("tire-size-markings FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 7 {
		t.Fatalf("expected seven English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 7 {
		t.Fatalf("expected seven Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "399_add_tire_size_markings_faq_content.down.sql")))
	for _, required := range []string{
		"USING SEED_FAQS",
		"EXISTING.QUESTION = SEED_FAQS.QUESTION",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("tire-size-markings FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("tire-size-markings FAQ down migration must preserve the route-owned FAQ page shell")
	}
}
