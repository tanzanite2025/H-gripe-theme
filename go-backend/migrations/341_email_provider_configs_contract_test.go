package migrations_test

import (
	"strings"
	"testing"
)

func TestEmailProviderConfigsMigrationStoresEncryptedPasswordAndSingleDefault(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "341_email_provider_configs.up.sql"))
	for _, fragment := range []string{
		"create table if not exists email_provider_configs",
		"password_encrypted text",
		"encryption_type varchar(16)",
		"chk_email_provider_port",
		"chk_email_provider_encryption",
		"create unique index if not exists uq_email_provider_default",
		"where is_default = true",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}

func TestEmailProviderConfigsMigrationRollbackDropsTable(t *testing.T) {
	sql := strings.ToLower(readMigrationFile(t, "341_email_provider_configs.down.sql"))
	if !strings.Contains(sql, "drop table if exists email_provider_configs") {
		t.Fatal("rollback must drop email_provider_configs")
	}
}
