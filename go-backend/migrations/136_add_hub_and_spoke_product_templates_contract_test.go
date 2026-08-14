package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readHubAndSpokeTemplateMigration(t *testing.T, name string) string {
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

func TestHubAndSpokeTemplateUpMigrationContract(t *testing.T) {
	sql := readHubAndSpokeTemplateMigration(t, "136_add_hub_and_spoke_product_templates.up.sql")

	for _, fragment := range []string{
		"('Hub', 'hub'",
		"('Spoke', 'spoke'",
		"'hub', '规格', 'Material'",
		"'spoke', '规格', 'Spoke Length'",
		"ON CONFLICT (slug) DO NOTHING",
		"WHERE NOT EXISTS",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestHubAndSpokeTemplateDownMigrationContract(t *testing.T) {
	sql := readHubAndSpokeTemplateMigration(t, "136_add_hub_and_spoke_product_templates.down.sql")

	for _, fragment := range []string{
		"DELETE FROM product_types",
		"slug IN ('hub', 'spoke')",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("down migration is missing contract fragment %q", fragment)
		}
	}
}
