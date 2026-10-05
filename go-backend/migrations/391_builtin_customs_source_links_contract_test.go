package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestBuiltinCustomsSourceLinkRefinementUsesExactOfficialHTSQueries(t *testing.T) {
	contents, err := os.ReadFile("391_refine_builtin_customs_source_links_after_official_verification.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"carbon-bicycle-rim",
		"carbon-bicycle-spoke",
		"bicycle-spoke",
		"bicycle-wheelset",
		"bicycle-hub",
		"carbon-bicycle-handlebar-stem",
		"carbon-bicycle-handlebar",
		"https://hts.usitc.gov/search?query=8714.92.10.00",
		"https://hts.usitc.gov/search?query=8714.92.50.00",
		"https://hts.usitc.gov/search?query=8714.99.80.00",
		"8714930090",
		"date '2026-10-05'",
		"date '2027-10-05'",
		"and source = 'built_in'",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("official customs source-link refinement is missing %q", requiredFragment)
		}
	}
}

func TestBuiltinCustomsSourceLinkRefinementRollbackDoesNotOverwriteSharedData(t *testing.T) {
	contents, err := os.ReadFile("391_refine_builtin_customs_source_links_after_official_verification.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	if strings.Contains(sql, "update customs_classification_profiles") || strings.Contains(sql, "delete from customs_classification_profiles") {
		t.Fatal("source-link refinement rollback must not overwrite or delete shared catalog data")
	}
}
