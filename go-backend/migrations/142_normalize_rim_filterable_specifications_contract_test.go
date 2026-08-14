package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readRimFilterableSpecificationsMigration(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test location")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "142_normalize_rim_filterable_specifications.up.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(data)
}

func TestNormalizeRimFilterableSpecificationsContract(t *testing.T) {
	sql := readRimFilterableSpecificationsMigration(t)

	for _, fragment := range []string{
		"product_type.slug = 'rim'",
		"definition.slug IN",
		"WHEN 'material' THEN TRUE",
		"WHEN 'brake_type' THEN TRUE",
		"WHEN 'tire_type' THEN TRUE",
		"WHEN 'wheel_size' THEN TRUE",
		"WHEN 'rim_depth' THEN FALSE",
		"WHEN 'inner_width' THEN FALSE",
		"WHEN 'outer_width' THEN FALSE",
		"WHEN 'spoke_holes' THEN FALSE",
		"WHEN 'erd' THEN FALSE",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing contract fragment %q", fragment)
		}
	}
}
