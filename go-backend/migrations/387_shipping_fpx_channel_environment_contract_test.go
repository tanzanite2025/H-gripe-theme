package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingFpxChannelEnvironmentMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("387_separate_shipping_fpx_channel_environments.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("387_separate_shipping_fpx_channel_environments.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"add column if not exists environment varchar(16) not null default 'production'",
		"drop index if exists idx_shipping_fpx_channel_service_code",
		"create unique index if not exists idx_shipping_fpx_channel_environment_service_code",
		"on shipping_fpx_channels (environment, service_code)",
		"where deleted_at is null",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}

	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"cannot remove 4px channel environments while active test channels exist",
		"drop index if exists idx_shipping_fpx_channel_environment_service_code",
		"drop column if exists environment",
		"create unique index if not exists idx_shipping_fpx_channel_service_code",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
