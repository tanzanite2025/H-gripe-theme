package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"tanzanite/internal/pkg/config"

	"github.com/lib/pq"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

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
		if !upMigrationNamePattern.MatchString(entry.Name()) {
			t.Errorf("migration %q does not follow <version>_<name>.up.sql", entry.Name())
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

		versionText, _, _ := strings.Cut(entry.Name(), "_")
		version, err := strconv.Atoi(versionText)
		if err != nil {
			t.Errorf("parse migration version from %q: %v", entry.Name(), err)
			continue
		}
		versions = append(versions, version)
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

	databaseName := fmt.Sprintf("tanzanite_migration_test_%d", time.Now().UnixNano())
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
		t.Fatalf("prepare fresh PostgreSQL schema: %v", err)
	}
	if err := PrepareSchema(ctx, gormDB, &cfg, "release"); err != nil {
		t.Fatalf("prepare existing PostgreSQL schema: %v", err)
	}

	var version int
	var dirty bool
	if err := testDB.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	expectedVersion := latestUpMigrationVersion(t, filepath.Join(backendRoot, "migrations"))
	if version != expectedVersion || dirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t", version, dirty)
	}

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
		"product_attributes",
		"product_variants",
		"chat_messages",
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

	emptyBusinessTables := []string{
		"users",
		"faqs",
		"galleries",
		"gallery_images",
		"product_registrations",
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

	assertProductTemplateSourceReset(ctx, t, testDB)
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

	for _, table := range []string{"products", "product_variants", "product_media", "product_spec_values", "shipping_template_bindings", "shipping_packaging_rule_applies"} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+pq.QuoteIdentifier(table)).Scan(&count); err != nil {
			t.Fatalf("count rows in %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected empty %s table, got %d rows", table, count)
		}
	}

	var productTypeCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_types").Scan(&productTypeCount); err != nil {
		t.Fatalf("count product types: %v", err)
	}
	if productTypeCount != 4 {
		t.Fatalf("expected four product templates, got %d", productTypeCount)
	}

	var specCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_spec_definitions").Scan(&specCount); err != nil {
		t.Fatalf("count product spec definitions: %v", err)
	}
	if specCount != 33 {
		t.Fatalf("expected thirty-three product spec definitions, got %d", specCount)
	}
}
