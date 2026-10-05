package migrations_test

import (
	"strings"
	"testing"
)

func TestWheelsetLacingTopologyFAQMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "374_add_wheelset_lacing_topology_faq_content.up.sql"))
	for _, required := range []string{
		"RESOURCES-WHEELSET-LACING-TOPOLOGY",
		"/RESOURCES/WHEELSET-SPOKE-LACING-TOPOLOGY-AND-GEOMETRY-REFERENCE",
		"INSERT INTO FAQ_PAGES",
		"INSERT INTO FAQS",
		"WHERE NOT EXISTS",
		"ERD AND PCD ARE DIAMETERS IN MM",
		"21H G3",
		"UNIFORM 24H 2:1",
		"NOT MEASURED STIFFNESS",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("wheelset lacing topology FAQ migration is missing contract fragment %q", required)
		}
	}

	if count := strings.Count(up, "'EN',"); count != 4 {
		t.Fatalf("expected four English FAQ seeds, found %d", count)
	}
	if count := strings.Count(up, "'ZH_CN',"); count != 4 {
		t.Fatalf("expected four Chinese FAQ seeds, found %d", count)
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "374_add_wheelset_lacing_topology_faq_content.down.sql")))
	for _, required := range []string{
		"WITH SEED_FAQS",
		"EXISTING.ANSWER = SEED_FAQS.ANSWER",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("wheelset lacing topology FAQ down migration is missing safe rollback fragment %q", required)
		}
	}
	if strings.Contains(down, "DELETE FROM FAQ_PAGES") {
		t.Fatal("wheelset lacing topology FAQ down migration must preserve the route-owned FAQ page shell")
	}
}
