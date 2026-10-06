package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCustomsClassificationVerificationMigrationAddsDatesAndCorrectsBuiltInBicycleCodes(t *testing.T) {
	contents, err := os.ReadFile("390_add_customs_classification_verification_metadata_and_correct_bicycle_component_profiles.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"add column if not exists verified_at date",
		"add column if not exists review_due_at date",
		"idx_customs_classification_profiles_review_due_at",
		"'carbon-bicycle-rim'",
		"'carbon-bicycle-spoke'",
		"'bicycle-spoke'",
		"'bicycle-hub'",
		"'bicycle-wheelset'",
		"'carbon-bicycle-handlebar-stem'",
		"'carbon-bicycle-handlebar'",
		"'871492'",
		"'87149210'",
		"'87149290'",
		"'871493'",
		"'87149300'",
		"'87149990'",
		"'87149910'",
		"date '2026-10-04'",
		"date '2027-10-04'",
		"on conflict (slug) do nothing",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs classification verification migration is missing %q", requiredFragment)
		}
	}
}

func TestCustomsClassificationVerificationMigrationRollbackDoesNotDeleteSharedCatalogRows(t *testing.T) {
	contents, err := os.ReadFile("390_add_customs_classification_verification_metadata_and_correct_bicycle_component_profiles.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	if strings.Contains(sql, "delete from customs_classification_profiles") {
		t.Fatal("customs classification verification rollback must not delete shared catalog rows")
	}
	for _, requiredFragment := range []string{
		"drop column if exists review_due_at",
		"drop column if exists verified_at",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs classification verification rollback is missing %q", requiredFragment)
		}
	}
}
