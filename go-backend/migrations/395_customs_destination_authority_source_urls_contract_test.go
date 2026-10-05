package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestCustomsDestinationAuthoritySourceURLsMigrationAddsIndependentOfficialLinks(t *testing.T) {
	contents, err := os.ReadFile("395_add_customs_destination_authority_source_urls.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"add column if not exists source_url_us text not null default ''",
		"add column if not exists source_url_eu text not null default ''",
		"add column if not exists source_url_uk text not null default ''",
		"set source_url_us = source_url",
		"https://ec.europa.eu/taxation_customs/dds2/taric",
		"https://www.gov.uk/trade-tariff/8714921000",
		"https://www.gov.uk/trade-tariff/8714929000",
		"https://www.gov.uk/trade-tariff/8714930090",
		"https://www.gov.uk/trade-tariff/8714999011",
		"https://www.gov.uk/trade-tariff/8714999040",
		"https://www.gov.uk/trade-tariff/8714991020",
		"date '2026-10-05'",
		"date '2027-10-05'",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs destination authority source migration is missing %q", requiredFragment)
		}
	}
}

func TestCustomsDestinationAuthoritySourceURLsMigrationRollbackRemovesOnlyAddedColumns(t *testing.T) {
	contents, err := os.ReadFile("395_add_customs_destination_authority_source_urls.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(contents))
	for _, requiredFragment := range []string{
		"drop column if exists source_url_us",
		"drop column if exists source_url_eu",
		"drop column if exists source_url_uk",
	} {
		if !strings.Contains(sql, requiredFragment) {
			t.Fatalf("customs destination authority rollback is missing %q", requiredFragment)
		}
	}
	if strings.Contains(sql, "drop table customs_classification_profiles") {
		t.Fatal("customs destination authority rollback must not drop the customs profile table")
	}
}
