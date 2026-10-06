package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingYanwenPublishedChannelEnvironmentMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("388_separate_shipping_yanwen_published_channel_environments.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("388_separate_shipping_yanwen_published_channel_environments.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"add column if not exists environment varchar(16) not null default 'production'",
		"drop index if exists idx_shipping_yanwen_channel_product_code",
		"create unique index if not exists idx_shipping_yanwen_channel_environment_product_code",
		"on shipping_yanwen_published_channels (environment, product_code)",
		"where deleted_at is null",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}

	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"cannot remove yanwen channel environments while active fat channels exist",
		"drop index if exists idx_shipping_yanwen_channel_environment_product_code",
		"drop column if exists environment",
		"create unique index if not exists idx_shipping_yanwen_channel_product_code",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
