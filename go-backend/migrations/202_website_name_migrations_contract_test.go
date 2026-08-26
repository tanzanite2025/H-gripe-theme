package migrations_test

import (
	"strings"
	"testing"
)

func TestWebsiteNameSeedMigrationDoesNotBlockLegacyNamespaceMigration(t *testing.T) {
	upSQL := readMigrationFile(t, "202_seed_website_name_defaults.up.sql")
	for _, fragment := range []string{
		"WHERE NOT EXISTS",
		"legacy.key = seed.legacy_key",
		"ON CONFLICT (key, locale) DO NOTHING",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("website name seed migration is missing contract fragment %q", fragment)
		}
	}

	for _, defaultText := range []string{
		"Website context",
		"This page will explain the working relationship",
		"网站说明",
		"这里会先说明这个网站背后的工作关系",
	} {
		if strings.Contains(upSQL, defaultText) {
			t.Fatalf("website name seed migration must not contain editable default text %q", defaultText)
		}
	}
}

func TestWebsiteNameSeedRollbackPreservesSavedContent(t *testing.T) {
	downSQL := readMigrationFile(t, "202_seed_website_name_defaults.down.sql")
	if !strings.Contains(downSQL, "AND value = ''") {
		t.Fatal("website name seed rollback must not delete non-empty saved content")
	}
}

func TestWebsiteNameNamespaceMigrationPreservesLegacyContentWhenSeedRowsExist(t *testing.T) {
	upSQL := readMigrationFile(t, "207_namespace_website_name_setting_keys.up.sql")
	for _, fragment := range []string{
		"UPDATE settings AS namespaced",
		"NULLIF(BTRIM(namespaced.value), '') IS NULL",
		"NULLIF(BTRIM(legacy.value), '') IS NOT NULL",
		"DELETE FROM settings",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("website name namespace migration is missing contract fragment %q", fragment)
		}
	}
}
