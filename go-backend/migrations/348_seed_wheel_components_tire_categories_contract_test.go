package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeedWheelComponentsTireCategoriesContract(t *testing.T) {
	upSQL, err := os.ReadFile(filepath.Join("348_seed_wheel_components_tire_categories.up.sql"))
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile(filepath.Join("348_seed_wheel_components_tire_categories.down.sql"))
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"wheel-components",
		"tire",
		"parent_id",
		"depth = 2",
		"'zh_cn'",
		"'轮组配件'",
		"'外胎'",
	} {
		if !strings.Contains(up, strings.ToLower(required)) {
			t.Fatalf("up migration missing %q", required)
		}
	}
	for _, required := range []string{"wheel-components", "tire", "product_category_translations"} {
		if !strings.Contains(down, strings.ToLower(required)) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
