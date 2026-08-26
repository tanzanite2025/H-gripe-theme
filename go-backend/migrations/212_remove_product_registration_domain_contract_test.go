package migrations_test

import (
	"strings"
	"testing"
)

func TestProductRegistrationDomainRemovalIsFinal(t *testing.T) {
	upSQL := readMigrationFile(t, "212_remove_product_registration_domain.up.sql")
	downSQL := readMigrationFile(t, "212_remove_product_registration_domain.down.sql")

	for _, fragment := range []string{
		"ALTER TABLE warranty_service_records",
		"DROP COLUMN IF EXISTS registration_id",
		"ALTER TABLE warranty_claims",
		"DROP TABLE IF EXISTS product_registrations CASCADE",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("registration removal migration is missing contract fragment %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"CREATE TABLE",
		"CREATE INDEX",
		"INSERT INTO",
		"product_registrations",
		"registration_id",
	} {
		if strings.Contains(strings.ToUpper(downSQL), strings.ToUpper(forbidden)) {
			t.Fatalf("registration removal down migration must not restore %q", forbidden)
		}
	}
}
