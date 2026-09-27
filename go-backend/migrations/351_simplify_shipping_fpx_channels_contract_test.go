package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestSimplifyShippingFpxChannelsMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("351_simplify_shipping_fpx_channels.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("351_simplify_shipping_fpx_channels.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, removed := range []string{
		"drop column if exists max_length_cm",
		"drop column if exists max_width_cm",
		"drop column if exists max_height_cm",
		"drop column if exists volumetric_divisor",
		"drop column if exists max_weight_grams",
		"drop column if exists notes",
		"drop column if exists sort_order",
	} {
		if !strings.Contains(up, removed) {
			t.Fatalf("up migration missing %q", removed)
		}
	}
	down := strings.ToLower(string(downSQL))
	for _, restored := range []string{
		"add column if not exists max_length_cm",
		"add column if not exists volumetric_divisor",
		"add column if not exists max_weight_grams",
		"add column if not exists notes",
		"add column if not exists sort_order",
	} {
		if !strings.Contains(down, restored) {
			t.Fatalf("down migration missing %q", restored)
		}
	}
}
