package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readSystemTemplateOptionsMigration(t *testing.T, name string) string {
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

func TestNormalizeSystemProductTemplateOptionsUpMigrationContract(t *testing.T) {
	sql := readSystemTemplateOptionsMigration(t, "141_normalize_system_product_template_options.up.sql")

	for _, fragment := range []string{
		"SET options = '[]'",
		"product_type.slug IN ('rim', 'carbon_frame', 'wheelset', 'handlebar', 'hub', 'spoke')",
		"definition.field_type = 'select'",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("up migration is missing contract fragment %q", fragment)
		}
	}
}

func TestNormalizeSystemProductTemplateOptionsDoesNotReintroduceFixedValues(t *testing.T) {
	for _, name := range []string{
		"027_reset_product_template_source.up.sql",
		"136_add_hub_and_spoke_product_templates.up.sql",
		"141_normalize_system_product_template_options.up.sql",
	} {
		sql := readSystemTemplateOptionsMigration(t, name)
		for _, fixedOption := range []string{
			`"Aluminum"`,
			`"Carbon Fiber"`,
			`"Disc Brake"`,
			`"Clincher"`,
			`"Center Lock"`,
			`"Standard External"`,
		} {
			if strings.Contains(sql, fixedOption) {
				t.Fatalf("%s reintroduces fixed system template option %s", name, fixedOption)
			}
		}
	}
}
