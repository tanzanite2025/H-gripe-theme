package migrations_test

import (
	"strings"
	"testing"
)

func TestSchwalbeTireSystemTemplateMigrationContract(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "355_add_schwalbe_tire_system_template.up.sql"))
	fieldSlugs := []string{
		"ARTICLE_NO",
		"EAN",
		"MODEL_NAME",
		"ETRTO",
		"INCH_DESIGNATION",
		"WEIGHT_G",
		"VERSION_LABEL",
		"COMPOUND",
		"COLOR",
		"BEAD",
		"E_BIKE_RATING",
		"EPI",
		"LOAD_KG",
		"SEAL",
		"TREAD",
		"MIN_PRESSURE_BAR",
		"MAX_PRESSURE_BAR",
		"MIN_PRESSURE_PSI",
		"MAX_PRESSURE_PSI",
	}

	for _, required := range []string{
		"INSERT INTO PRODUCT_SPECIFICATION_TEMPLATES",
		"'SCHWALBE TIRE'",
		"'SCHWALBE_TIRE'",
		"IS_SYSTEM_MANAGED",
		"TRUE",
		"ON CONFLICT (SLUG) DO NOTHING",
		"INSERT INTO PRODUCT_SPEC_DEFINITIONS",
		"WHERE NOT EXISTS",
		"'PRODUCT NAME'",
		"'LOAD (KG)'",
		"'MIN. BAR'",
		"'MAX. BAR'",
		"'MIN. PSI'",
		"'MAX. PSI'",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("system template migration is missing contract fragment %q", required)
		}
	}
	for _, slug := range fieldSlugs {
		if count := strings.Count(up, "'"+slug+"'"); count == 0 {
			t.Fatalf("system template migration must define field slug %s", slug)
		}
	}
	if len(fieldSlugs) != 19 {
		t.Fatalf("Schwalbe system template contract must contain 19 field slugs, found %d", len(fieldSlugs))
	}
	for _, forbidden := range []string{"METADATA", "OPTIONS", "IS_VARIANT_OPTION", "PRODUCT_CATEGORIES"} {
		if strings.Contains(up, forbidden) {
			t.Fatalf("system template migration must not contain %q", forbidden)
		}
	}

	// A fresh database has no definitions yet. The seed must execute before
	// the exact-count post-condition, otherwise migration 355 fails at zero
	// definitions and never creates the 19-field template.
	seedInsert := strings.Index(up, "INSERT INTO PRODUCT_SPEC_DEFINITIONS")
	countCheck := strings.Index(up, "SELECT COUNT(*)")
	if seedInsert < 0 || countCheck < 0 || seedInsert > countCheck {
		t.Fatal("system template migration must seed definitions before validating the 19-definition post-condition")
	}
}

func TestSchwalbeHistoricalMigrationsRemainForwardOnly(t *testing.T) {
	for _, name := range []string{
		"354_create_schwalbe_tire_specifications.down.sql",
		"355_add_schwalbe_tire_system_template.down.sql",
		"356_add_schwalbe_review_audit.down.sql",
	} {
		down := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, name)))
		if down == "" {
			t.Fatalf("%s must document the intentional forward-only policy", name)
		}
		if strings.Contains(down, "DROP TABLE") || strings.Contains(down, "DELETE FROM") {
			t.Fatalf("%s must not delete Schwalbe production data", name)
		}
	}
}

func TestSchwalbeCatalogIsRestoredAsNonSalesCandidateData(t *testing.T) {
	up := strings.ToUpper(readMigrationFile(t, "357_retire_schwalbe_standalone_catalog.up.sql"))
	for _, required := range []string{
		"TO_REGCLASS('PUBLIC.SCHWALBE_TIRE_SPECIFICATIONS')",
		"EXECUTE 'SELECT EXISTS (SELECT 1 FROM PUBLIC.SCHWALBE_TIRE_SPECIFICATIONS)' INTO HAS_ROWS",
		"RAISE EXCEPTION",
		"DROP TABLE IF EXISTS SCHWALBE_TIRE_SPECIFICATIONS",
		"WHERE SLUG = 'SCHWALBE_TIRE'",
		"ENTERED DIRECTLY ON EACH PRODUCT",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("catalog retirement migration is missing safety or template contract fragment %q", required)
		}
	}
	down := strings.ToUpper(readMigrationFile(t, "357_retire_schwalbe_standalone_catalog.down.sql"))
	if strings.Contains(down, "CREATE TABLE") || strings.Contains(down, "INSERT INTO SCHWALBE_TIRE_SPECIFICATIONS") {
		t.Fatal("catalog retirement down migration must not recreate the redundant catalog")
	}

	restore := strings.ToUpper(readMigrationFile(t, "358_restore_schwalbe_candidate_catalog.up.sql"))
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS SCHWALBE_TIRE_SPECIFICATIONS",
		"ARTICLE_NO VARCHAR(32) PRIMARY KEY",
		"SOURCE_URL TEXT NOT NULL",
		"SOURCE_CHECKED_AT DATE NOT NULL",
		"CHK_SCHWALBE_CATALOG_PRESSURE_BAR_ORDER",
		"CHK_SCHWALBE_CATALOG_PRESSURE_PSI_ORDER",
		"SEPARATE FROM SALES PRODUCTS",
	} {
		if !strings.Contains(restore, required) {
			t.Fatalf("catalog restore migration is missing fragment %q", required)
		}
	}
	for _, forbidden := range []string{"VERIFICATION_STATUS", "REVIEWED_BY_ID", "SOURCE_SUBMITTED_BY_ID", "REVIEW_NOTE"} {
		if strings.Contains(restore, forbidden) {
			t.Fatalf("candidate catalog migration must not add approval field %q", forbidden)
		}
	}
	restoreDown := strings.ToUpper(strings.TrimSpace(readMigrationFile(t, "358_restore_schwalbe_candidate_catalog.down.sql")))
	if strings.Contains(restoreDown, "DROP TABLE") || strings.Contains(restoreDown, "DELETE FROM") {
		t.Fatal("candidate catalog down migration must not delete catalog data")
	}
}
