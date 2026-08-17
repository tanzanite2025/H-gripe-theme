package migrations_test

import (
	"strings"
	"testing"
)

func TestMediaDerivativePresetMigrationContract(t *testing.T) {
	upSQL := readSiteQualityMigration(t, "179_media_derivative_presets.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS media_derivative_presets",
		"generation_version INTEGER NOT NULL DEFAULT 1",
		"ADD COLUMN IF NOT EXISTS preset_version INTEGER NOT NULL DEFAULT 1",
		"ck_media_asset_derivatives_preset_version",
		"('thumbnail', '缩略图', 320, 10, 1, TRUE)",
		"('card', '卡片图', 640, 20, 1, TRUE)",
		"('large', '大图', 1600, 30, 1, TRUE)",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("media derivative preset migration is missing contract fragment %q", fragment)
		}
	}

	downSQL := readSiteQualityMigration(t, "179_media_derivative_presets.down.sql")
	for _, fragment := range []string{
		"DROP COLUMN IF EXISTS preset_version",
		"DROP TABLE IF EXISTS media_derivative_presets",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("media derivative preset down migration is missing contract fragment %q", fragment)
		}
	}
}
