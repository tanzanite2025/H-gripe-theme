package migrations_test

import (
	"strings"
	"testing"
)

func TestSchwalbeTireHooklessCompatibilityMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "369_add_schwalbe_tire_hookless_compatibility.up.sql"))
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS SCHWALBE_TIRE_HOOKLESS_COMPATIBILITY",
		"SCOPE_KEY VARCHAR(240) PRIMARY KEY",
		"ARTICLE_NO VARCHAR(32)",
		"MODEL_NAME VARCHAR(180)",
		"STATUS VARCHAR(32) NOT NULL",
		"SOURCE_BASIS TEXT NOT NULL",
		"SOURCE_VERSION VARCHAR(64) NOT NULL",
		"SOURCE_URL TEXT NOT NULL",
		"SOURCE_CHECKED_AT DATE NOT NULL",
		"ARTICLE_NO IS NOT NULL OR MODEL_NAME IS NOT NULL",
		"STATUS IN ('SUPPORTED', 'NOT_SUPPORTED', 'UNKNOWN')",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_SCHWALBE_HOOKLESS_COMPATIBILITY_ARTICLE",
		"CREATE UNIQUE INDEX IF NOT EXISTS UQ_SCHWALBE_HOOKLESS_COMPATIBILITY_MODEL",
		"HOOKLESS COMPATIBLE = YES",
		"HOOKLESS COMPATIBLE = NO",
		"ON CONFLICT (SCOPE_KEY) DO UPDATE SET",
		"2026-09-28",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("Hookless compatibility migration is missing contract fragment %q", required)
		}
	}

	for _, unknownModel := range []string{"JUMBO JIM", "ROCKET RON", "WICKED WILL"} {
		if strings.Contains(up, "('MODEL:"+unknownModel+"'") {
			t.Fatalf("model %s is present despite conflicting official Yes/No sources", unknownModel)
		}
	}

	down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "369_add_schwalbe_tire_hookless_compatibility.down.sql")))
	if !strings.Contains(down, "FORWARD-ONLY") {
		t.Fatal("Hookless compatibility down migration must document the forward-only data policy")
	}
	if strings.Contains(down, "DROP TABLE") || strings.Contains(down, "DELETE FROM") {
		t.Fatal("Hookless compatibility down migration must not delete reviewed facts")
	}
}
