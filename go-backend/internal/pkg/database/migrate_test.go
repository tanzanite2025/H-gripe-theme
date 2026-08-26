package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var migrationNamePattern = regexp.MustCompile(`^\d+_[a-z0-9_]+\.(up|down)\.sql$`)
var upMigrationNamePattern = regexp.MustCompile(`^\d+_[a-z0-9_]+\.up\.sql$`)
var unsupportedMigrationSyntaxPattern = regexp.MustCompile(
	`(?i)\bAUTO_INCREMENT\b|\bUNSIGNED\b|\bUNIX_TIMESTAMP\b|\bUNIQUE\s+KEY\b|\bENGINE=|\+goose`,
)

func TestSQLMigrationFilesFollowGolangMigrateConvention(t *testing.T) {
	migrationDir := filepath.Join("..", "..", "..", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}

	versions := make([]int, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		if !migrationNamePattern.MatchString(entry.Name()) {
			t.Errorf("migration %q does not follow <version>_<name>.(up|down).sql", entry.Name())
			continue
		}
		contents, err := os.ReadFile(filepath.Join(migrationDir, entry.Name()))
		if err != nil {
			t.Errorf("read migration %q: %v", entry.Name(), err)
			continue
		}
		if unsupportedMigrationSyntaxPattern.Match(contents) {
			t.Errorf("migration %q contains unsupported MySQL or Goose syntax", entry.Name())
		}

		if upMigrationNamePattern.MatchString(entry.Name()) {
			versionText, _, _ := strings.Cut(entry.Name(), "_")
			version, err := strconv.Atoi(versionText)
			if err != nil {
				t.Errorf("parse migration version from %q: %v", entry.Name(), err)
				continue
			}
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		t.Fatal("no SQL migrations found")
	}

	sort.Ints(versions)
	for index, version := range versions {
		expected := index + 1
		if version != expected {
			t.Fatalf("migration sequence is not contiguous: expected %03d, got %03d", expected, version)
		}
	}
}

func TestPrepareSchemaAgainstFreshPostgres(t *testing.T) {
	host := os.Getenv("DB_HOST")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	if host == "" || username == "" || password == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	port := 5432
	if portText := os.Getenv("DB_PORT"); portText != "" {
		parsedPort, err := strconv.Atoi(portText)
		if err != nil {
			t.Fatalf("parse DB_PORT: %v", err)
		}
		port = parsedPort
	}

	adminDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host,
		port,
		username,
		password,
	)
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatalf("open PostgreSQL admin connection: %v", err)
	}

	databaseName := fmt.Sprintf("commerce_platform_migration_test_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(databaseName)); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create migration test database: %v", err)
	}

	testDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		username,
		password,
		databaseName,
	)
	gormDB, err := gorm.Open(postgresdriver.Open(testDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		_ = adminDB.Close()
		t.Fatalf("open GORM migration test database: %v", err)
	}
	testDB, err := gormDB.DB()
	if err != nil {
		_ = adminDB.Close()
		t.Fatalf("get migration test database: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_ = testDB.Close()
		_, _ = adminDB.ExecContext(
			cleanupCtx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1",
			databaseName,
		)
		_, _ = adminDB.ExecContext(
			cleanupCtx,
			"DROP DATABASE IF EXISTS "+pq.QuoteIdentifier(databaseName),
		)
		_ = adminDB.Close()
	})

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	backendRoot := filepath.Clean(filepath.Join(originalDir, "..", "..", ".."))
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("change to backend root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	cfg := config.DatabaseConfig{Driver: "postgres"}
	if err := PrepareSchema(ctx, gormDB, &cfg, "release"); err != nil {
		t.Fatalf("prepare empty PostgreSQL schema from SQL migrations: %v", err)
	}
	migrationDriver, err := postgres.WithInstance(testDB, &postgres.Config{})
	if err != nil {
		t.Fatalf("create PostgreSQL migration driver: %v", err)
	}
	migrator, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", migrationDriver)
	if err != nil {
		t.Fatalf("create migration runner: %v", err)
	}
	if err := migrator.Migrate(110); err != nil {
		t.Fatalf("roll back migrations to version 110: %v", err)
	}
	assertMigrationState(ctx, t, testDB, 110, false)
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("reapply migrations after version 110: %v", err)
	}
	if err := PrepareSchema(ctx, gormDB, &cfg, "release"); err != nil {
		t.Fatalf("prepare existing PostgreSQL schema: %v", err)
	}

	expectedVersion := latestUpMigrationVersion(t, filepath.Join(backendRoot, "migrations"))
	assertMigrationState(ctx, t, testDB, expectedVersion, false)

	// The catalog seed is deliberately idempotent so a manual recovery rerun cannot
	// duplicate products or variants.
	catalogMigration, err := os.ReadFile(filepath.Join(backendRoot, "migrations", "015_seed_g35_catalog.up.sql"))
	if err != nil {
		t.Fatalf("read G35 catalog migration: %v", err)
	}
	if _, err := testDB.ExecContext(ctx, string(catalogMigration)); err != nil {
		t.Fatalf("rerun G35 catalog migration: %v", err)
	}

	requiredTables := []string{
		"orders",
		"order_items",
		"transactions",
		"refunds",
		"refund_line_items",
		"order_policy_disclosures",
		"product_attributes",
		"product_variants",
		"spoke_rim_brands",
		"spoke_rim_models",
		"spoke_hub_brands",
		"spoke_hub_models",
		"spoke_build_presets",
	}
	for _, table := range requiredTables {
		var exists bool
		if err := testDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists); err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("required table %s does not exist", table)
		}
	}

	removedConstraints := []struct {
		tableName      string
		constraintName string
	}{
		{
			tableName:      "product_specification_templates",
			constraintName: "fk_product_specification_templates_image_media_asset",
		},
		{
			tableName:      "product_specification_templates",
			constraintName: "ck_product_specification_templates_image_reference_pair",
		},
	}
	for _, constraint := range removedConstraints {
		var exists bool
		if err := testDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints
				WHERE constraint_schema = 'public'
				  AND table_name = $1
				  AND constraint_name = $2
			)
		`, constraint.tableName, constraint.constraintName).Scan(&exists); err != nil {
			t.Fatalf("check constraint %s: %v", constraint.constraintName, err)
		}
		if exists {
			t.Fatalf("removed constraint %s still exists on %s", constraint.constraintName, constraint.tableName)
		}
	}

	for _, columnName := range []string{"image_media_asset_id", "image_url"} {
		var exists bool
		if err := testDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = 'product_specification_templates'
				  AND column_name = $1
			)
		`, columnName).Scan(&exists); err != nil {
			t.Fatalf("check removed column %s: %v", columnName, err)
		}
		if exists {
			t.Fatalf("removed column %s still exists on product_specification_templates", columnName)
		}
	}

	retiredTables := []string{
		"chat_messages",
		"chat_sessions",
		"shipping_template_bindings",
		"product_registrations",
	}
	for _, table := range retiredTables {
		var exists bool
		if err := testDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists); err != nil {
			t.Fatalf("check retired table %s: %v", table, err)
		}
		if exists {
			t.Fatalf("retired table %s should not exist", table)
		}
	}

	emptyBusinessTables := []string{
		"users",
		"galleries",
		"gallery_images",
		"warranty_claims",
		"tickets",
		"ticket_messages",
		"browsing_history",
	}
	for _, table := range emptyBusinessTables {
		var count int
		query := "SELECT COUNT(*) FROM " + pq.QuoteIdentifier(table)
		if err := testDB.QueryRowContext(ctx, query).Scan(&count); err != nil {
			t.Fatalf("count rows in %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("business table %s contains %d seeded rows", table, count)
		}
	}

	for _, column := range []struct {
		table string
		name  string
	}{
		{table: "warranty_claims", name: "registration_id"},
		{table: "warranty_service_records", name: "registration_id"},
	} {
		var exists bool
		if err := testDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = $1
				  AND column_name = $2
			)
		`, column.table, column.name).Scan(&exists); err != nil {
			t.Fatalf("check removed column %s.%s: %v", column.table, column.name, err)
		}
		if exists {
			t.Fatalf("removed column %s.%s still exists", column.table, column.name)
		}
	}

	var faqCount int
	if err := testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM faqs").Scan(&faqCount); err != nil {
		t.Fatalf("count seeded FAQs: %v", err)
	}
	if faqCount == 0 {
		t.Fatalf("expected seeded FAQs")
	}

	assertProductTemplateSourceReset(ctx, t, testDB)
	assertRefundAndPolicyMigrationState(ctx, t, testDB)
}

func assertMigrationState(ctx context.Context, t *testing.T, db *sql.DB, expectedVersion int, expectedDirty bool) {
	t.Helper()

	var version int
	var dirty bool
	if err := db.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if version != expectedVersion || dirty != expectedDirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t, want version=%d dirty=%t",
			version, dirty, expectedVersion, expectedDirty)
	}
}

func latestUpMigrationVersion(t *testing.T, migrationDir string) int {
	t.Helper()

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}

	latest := 0
	for _, entry := range entries {
		if entry.IsDir() || !upMigrationNamePattern.MatchString(entry.Name()) {
			continue
		}
		versionText, _, _ := strings.Cut(entry.Name(), "_")
		version, err := strconv.Atoi(versionText)
		if err != nil {
			t.Fatalf("parse migration version from %q: %v", entry.Name(), err)
		}
		if version > latest {
			latest = version
		}
	}
	if latest == 0 {
		t.Fatal("no SQL migrations found")
	}

	return latest
}

func assertProductTemplateSourceReset(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()

	for _, table := range []string{"products", "product_variants", "product_media", "product_spec_values", "shipping_packaging_rule_applies"} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+pq.QuoteIdentifier(table)).Scan(&count); err != nil {
			t.Fatalf("count rows in %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected empty %s table, got %d rows", table, count)
		}
	}

	var productSpecificationTemplateCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_specification_templates").Scan(&productSpecificationTemplateCount); err != nil {
		t.Fatalf("count product specification templates: %v", err)
	}
	if productSpecificationTemplateCount != 6 {
		t.Fatalf("expected six product templates, got %d", productSpecificationTemplateCount)
	}

	var specCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_spec_definitions").Scan(&specCount); err != nil {
		t.Fatalf("count product spec definitions: %v", err)
	}
	if specCount != 46 {
		t.Fatalf("expected forty-six product spec definitions, got %d", specCount)
	}
}

func assertRefundAndPolicyMigrationState(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()

	assertPostgresColumns(ctx, t, db, "refunds", map[string]string{
		"requested_amount":         "numeric",
		"discount_clawback_amount": "numeric",
		"calculation_snapshot":     "text",
		"fx_snapshot":              "jsonb",
	})
	assertPostgresColumns(ctx, t, db, "refund_line_items", map[string]string{
		"refund_id":     "int8",
		"order_id":      "int8",
		"order_item_id": "int8",
		"restock":       "bool",
		"restocked_at":  "timestamptz",
	})
	assertPostgresColumns(ctx, t, db, "order_policy_disclosures", map[string]string{
		"order_id":         "int8",
		"policy_key":       "varchar",
		"locale":           "varchar",
		"requested_locale": "varchar",
		"fallback":         "bool",
		"policy_version":   "varchar",
		"policy_hash":      "varchar",
		"policy_json":      "text",
		"policy_url":       "text",
		"disclosed_at":     "timestamptz",
		"consented_at":     "timestamptz",
		"source":           "varchar",
	})

	var policyValue, policyType, policyGroup string
	var isPublic bool
	if err := db.QueryRowContext(ctx, `
		SELECT value, type, "group", is_public
		FROM settings
		WHERE key = 'refund_return_policy' AND locale = 'en'
	`).Scan(&policyValue, &policyType, &policyGroup, &isPublic); err != nil {
		t.Fatalf("load default refund and return policy setting: %v", err)
	}
	if policyType != "json" || policyGroup != "refund_return" || !isPublic {
		t.Fatalf(
			"unexpected refund and return policy metadata: type=%q group=%q is_public=%t",
			policyType,
			policyGroup,
			isPublic,
		)
	}
	var policy struct {
		Title    string            `json:"title"`
		Sections []json.RawMessage `json:"sections"`
	}
	if err := json.Unmarshal([]byte(policyValue), &policy); err != nil {
		t.Fatalf("decode default refund and return policy setting: %v", err)
	}
	if policy.Title != "Refund & Return Policy" || len(policy.Sections) == 0 {
		t.Fatalf("default refund and return policy payload is incomplete")
	}

	var hasUniqueConstraint bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints
			WHERE constraint_schema = 'public'
			  AND table_name = 'order_policy_disclosures'
			  AND constraint_name = 'uq_order_policy_disclosures_order_policy'
			  AND constraint_type = 'UNIQUE'
		)
	`).Scan(&hasUniqueConstraint); err != nil {
		t.Fatalf("check order policy disclosure uniqueness constraint: %v", err)
	}
	if !hasUniqueConstraint {
		t.Fatal("order policy disclosure uniqueness constraint is missing")
	}
}

func assertPostgresColumns(ctx context.Context, t *testing.T, db *sql.DB, tableName string, expected map[string]string) {
	t.Helper()

	for columnName, expectedType := range expected {
		var actualType string
		if err := db.QueryRowContext(ctx, `
			SELECT udt_name
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND column_name = $2
		`, tableName, columnName).Scan(&actualType); err != nil {
			t.Fatalf("load column %s.%s: %v", tableName, columnName, err)
		}
		if actualType != expectedType {
			t.Fatalf("unexpected type for %s.%s: got %q, want %q", tableName, columnName, actualType, expectedType)
		}
	}
}
