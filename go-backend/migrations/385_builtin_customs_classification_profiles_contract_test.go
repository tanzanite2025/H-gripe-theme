package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestBuiltInCustomsClassificationProfilesMigrationSeedsReusableBicycleTemplates(t *testing.T) {
	contents, err := os.ReadFile("385_seed_builtin_customs_classification_profiles.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"insert into customs_classification_profiles",
		"on conflict (slug) do nothing",
		"'carbon-bicycle-rim'",
		"'carbon-bicycle-spoke'",
		"'bicycle-spoke'",
		"'bicycle-hub'",
		"'bicycle-wheelset'",
		"'碳纤维车圈'",
		"'碳纤维辐条'",
		"'自行车辐条'",
		"'自行车花鼓'",
		"'自行车轮组'",
		"'built_in'",
		"'871499'",
		"'87149990'",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("built-in customs classification migration is missing %q", requiredFragment)
		}
	}
}

func TestBuiltInCustomsClassificationProfilesMigrationDoesNotDeleteCatalogRowsOnRollback(t *testing.T) {
	contents, err := os.ReadFile("385_seed_builtin_customs_classification_profiles.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(contents)), "delete from customs_classification_profiles") {
		t.Fatal("built-in customs classification rollback must not delete shared catalog rows")
	}
}
