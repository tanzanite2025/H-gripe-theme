package migrations_test

import (
	"strings"
	"testing"
)

func TestInnerTubeSelectionFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "398_add_inner_tube_selection_faq_content.up.sql"))
	for _, required := range []string{
		"GUIDES-TIREGUIDES-CHOOSE-INNER-TUBE",
		"/GUIDES/TIREGUIDES/CHOOSE-INNER-TUBE",
		"INSERT INTO FAQ_PAGES",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"6.5 MM OUTER-LIP OFFSET",
		"80 MM NATIVE VALVE WITH A 40 MM EXTENDER",
		"REMOVABLE VALVE CORE",
		"MEASUREMENT ERROR",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("inner-tube selection FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 7 {
		t.Fatalf("expected seven English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 7 {
		t.Fatalf("expected seven Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "398_add_inner_tube_selection_faq_content.down.sql")))
	for _, required := range []string{
		"WITH SEED_FAQS",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("inner-tube selection FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("inner-tube selection FAQ down migration must preserve the route-owned FAQ page shell")
	}
}
