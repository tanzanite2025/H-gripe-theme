package migrations_test

import (
	"strings"
	"testing"
)

func TestSocialPlatformsMigrationKeepsTheSupportedPublicShape(t *testing.T) {
	upSQL := readMigrationFile(t, "215_update_social_platforms.up.sql")
	downSQL := readMigrationFile(t, "215_update_social_platforms.down.sql")

	for _, fragment := range []string{
		"DELETE FROM settings",
		"key = 'twitter'",
		"key IN ('linkedin', 'wechat')",
		"('x', '', 'string', 'en', 'social'",
		"('reddit', '', 'string', 'en', 'social'",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("social platform migration is missing contract fragment %q", fragment)
		}
	}

	for _, forbidden := range []string{"linkedin", "wechat"} {
		if strings.Contains(strings.ToLower(upSQL), "\n    ('"+forbidden+"'") {
			t.Fatalf("social platform migration must not seed removed platform %q", forbidden)
		}
	}

	if !strings.Contains(downSQL, "key IN ('x', 'reddit')") {
		t.Fatal("social platform down migration must remove the settings introduced by the migration")
	}

	for _, forbidden := range []string{"twitter", "linkedin", "wechat"} {
		if strings.Contains(strings.ToLower(downSQL), forbidden) {
			t.Fatalf("social platform down migration must not recreate removed platform %q", forbidden)
		}
	}
}
