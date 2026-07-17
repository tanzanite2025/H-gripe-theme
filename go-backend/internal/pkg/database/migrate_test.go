package database

import (
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
)

var upMigrationNamePattern = regexp.MustCompile(`^\d+_[a-z0-9_]+\.up\.sql$`)

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

func TestRunSQLMigrationsAgainstFreshPostgres(t *testing.T) {
	host := os.Getenv("DB_HOST")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	if host == "" || username == "" || password == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}

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
	if _, err := adminDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(databaseName)); err != nil {
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
	testDB, err := sql.Open("postgres", testDSN)
	if err != nil {
		_ = adminDB.Close()
		t.Fatalf("open migration test database: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Close()
		_, _ = adminDB.Exec(
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1",
			databaseName,
		)
		_, _ = adminDB.Exec("DROP DATABASE IF EXISTS " + pq.QuoteIdentifier(databaseName))
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
	if err := RunSQLMigrations(testDB, &cfg); err != nil {
		t.Fatalf("run SQL migrations: %v", err)
	}

	var version int
	var dirty bool
	if err := testDB.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if version != 13 || dirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t", version, dirty)
	}
}
