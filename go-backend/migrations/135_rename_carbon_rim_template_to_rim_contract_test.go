package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readRimTemplateRenameMigration(t *testing.T, name string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test location")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(data)
}

func TestRimTemplateRenameUpMigrationContract(t *testing.T) {
	sql := readRimTemplateRenameMigration(t, "135_rename_carbon_rim_template_to_rim.up.sql")

	for _, fragment := range []string{
		"slug = 'carbon_rim'",
		"slug = 'rim'",
		"name = 'Rim'",
		"product_type_translations",
		"BTRIM(name) = 'Carbon Rim'",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestRimTemplateRenameDownMigrationContract(t *testing.T) {
	sql := readRimTemplateRenameMigration(t, "135_rename_carbon_rim_template_to_rim.down.sql")

	for _, fragment := range []string{
		"slug = 'rim'",
		"slug = 'carbon_rim'",
		"name = 'Carbon Rim'",
		"product_type_translations",
		"BTRIM(name) = 'Rim'",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
