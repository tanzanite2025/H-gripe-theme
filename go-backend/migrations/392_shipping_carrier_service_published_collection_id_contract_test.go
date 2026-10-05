package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestShippingCarrierServicePublishedCollectionIDMigrationContract(t *testing.T) {
	upSQL, err := os.ReadFile("392_add_stable_published_collection_ids_to_shipping_carrier_services.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downSQL, err := os.ReadFile("392_add_stable_published_collection_ids_to_shipping_carrier_services.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	up := strings.ToLower(string(upSQL))
	for _, required := range []string{
		"add column if not exists fpx_channel_id bigint",
		"add column if not exists yanwen_published_channel_id bigint",
		"update shipping_carrier_services as service",
		"environment = 'production'",
		"fk_shipping_carrier_services_fpx_channel",
		"fk_shipping_carrier_services_yanwen_channel",
		"chk_shipping_carrier_services_single_published_collection",
		"check (not (fpx_channel_id is not null and yanwen_published_channel_id is not null))",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration missing %q", required)
		}
	}

	down := strings.ToLower(string(downSQL))
	for _, required := range []string{
		"drop constraint if exists chk_shipping_carrier_services_single_published_collection",
		"drop constraint if exists fk_shipping_carrier_services_fpx_channel",
		"drop constraint if exists fk_shipping_carrier_services_yanwen_channel",
		"drop column if exists fpx_channel_id",
		"drop column if exists yanwen_published_channel_id",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("down migration missing %q", required)
		}
	}
}
