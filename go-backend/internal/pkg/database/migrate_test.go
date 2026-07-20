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
	if version != 15 || dirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t", version, dirty)
	}

	// The catalog seed is deliberately idempotent so a manual recovery rerun cannot
	// duplicate products, variants, or images.
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

	assertG35CatalogSeed(ctx, t, testDB)
}

func assertG35CatalogSeed(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()

	var productCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&productCount); err != nil {
		t.Fatalf("count catalog products: %v", err)
	}
	if productCount != 1 {
		t.Fatalf("expected one catalog product, got %d", productCount)
	}

	var (
		productSKU    string
		productName   string
		productSlug   string
		productPrice  string
		productStock  int
		productStatus string
		productLocale string
		productType   string
		featured      bool
	)
	if err := db.QueryRowContext(ctx, `
		SELECT p.sku, p.name, p.slug, p.price::text, p.stock, p.status, p.locale,
		       pt.slug, p.featured
		FROM products p
		JOIN product_types pt ON pt.id = p.product_type_id
	`).Scan(
		&productSKU,
		&productName,
		&productSlug,
		&productPrice,
		&productStock,
		&productStatus,
		&productLocale,
		&productType,
		&featured,
	); err != nil {
		t.Fatalf("read G35 catalog product: %v", err)
	}

	if productSKU != "G35-370G-1PC" || productName != "G35 Carbon Rim" || productSlug != "g35-carbon-rim" {
		t.Fatalf("unexpected G35 identity: sku=%q name=%q slug=%q", productSKU, productName, productSlug)
	}
	if productPrice != "111.14" || productStock != 0 || productStatus != "active" || productLocale != "en" {
		t.Fatalf(
			"unexpected G35 summary: price=%s stock=%d status=%q locale=%q",
			productPrice,
			productStock,
			productStatus,
			productLocale,
		)
	}
	if productType != "carbon_rim" || featured {
		t.Fatalf("unexpected G35 catalog classification: type=%q featured=%t", productType, featured)
	}

	var specCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM product_spec_definitions psd
		JOIN product_types pt ON pt.id = psd.product_type_id
		WHERE pt.slug = 'carbon_rim'
		  AND psd.slug IN ('listed_weight', 'pack_size')
		  AND psd.is_required = TRUE
		  AND psd.is_variant_option = TRUE
	`).Scan(&specCount); err != nil {
		t.Fatalf("count G35 variant definitions: %v", err)
	}
	if specCount != 2 {
		t.Fatalf("expected two G35 variant definitions, got %d", specCount)
	}

	type expectedVariant struct {
		sku          string
		title        string
		optionValues string
		price        string
		isDefault    bool
		sortOrder    int
	}
	expectedVariants := []expectedVariant{
		{"G35-370G-1PC", "370 g / 1 piece", `{"listed_weight":"370 g","pack_size":"1 piece"}`, "111.14", true, 10},
		{"G35-370G-2PC", "370 g / 2 pieces", `{"listed_weight":"370 g","pack_size":"2 pieces"}`, "222.28", false, 20},
		{"G35-460G-1PC", "460 g / 1 piece", `{"listed_weight":"460 g","pack_size":"1 piece"}`, "78.73", false, 30},
		{"G35-460G-2PC", "460 g / 2 pieces", `{"listed_weight":"460 g","pack_size":"2 pieces"}`, "157.45", false, 40},
	}

	rows, err := db.QueryContext(ctx, `
		SELECT pv.sku, pv.title, pv.option_values, pv.price::text, pv.stock,
		       pv.weight_grams, pv.is_default, pv.is_active, pv.sort_order
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		WHERE p.sku = 'G35-370G-1PC'
		ORDER BY pv.sort_order, pv.id
	`)
	if err != nil {
		t.Fatalf("query G35 variants: %v", err)
	}
	defer rows.Close()

	variantIndex := 0
	for rows.Next() {
		if variantIndex >= len(expectedVariants) {
			t.Fatal("G35 catalog contains more than four variants")
		}
		expected := expectedVariants[variantIndex]
		var (
			sku          string
			title        string
			optionValues string
			price        string
			stock        int
			weight       int
			isDefault    bool
			isActive     bool
			sortOrder    int
		)
		if err := rows.Scan(
			&sku,
			&title,
			&optionValues,
			&price,
			&stock,
			&weight,
			&isDefault,
			&isActive,
			&sortOrder,
		); err != nil {
			t.Fatalf("scan G35 variant %d: %v", variantIndex, err)
		}
		if sku != expected.sku || title != expected.title || optionValues != expected.optionValues ||
			price != expected.price || isDefault != expected.isDefault || sortOrder != expected.sortOrder {
			t.Fatalf(
				"unexpected G35 variant %d: sku=%q title=%q options=%q price=%s default=%t sort=%d",
				variantIndex,
				sku,
				title,
				optionValues,
				price,
				isDefault,
				sortOrder,
			)
		}
		if stock != 0 || weight != 0 || !isActive {
			t.Fatalf("unsafe G35 variant %q: stock=%d weight=%d active=%t", sku, stock, weight, isActive)
		}
		variantIndex++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate G35 variants: %v", err)
	}
	if variantIndex != len(expectedVariants) {
		t.Fatalf("expected four G35 variants, got %d", variantIndex)
	}

	var (
		imageCount int
		imageURL   string
		imageAlt   string
		imageOrder int
	)
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), MIN(pi.url), MIN(pi.alt), MIN(pi."order")
		FROM product_images pi
		JOIN products p ON p.id = pi.product_id
		WHERE p.sku = 'G35-370G-1PC'
	`).Scan(&imageCount, &imageURL, &imageAlt, &imageOrder); err != nil {
		t.Fatalf("read G35 image: %v", err)
	}
	if imageCount != 1 || imageURL != "/company/aboutus/appearance/tanzanite-carbon-rim-finish1.webp" ||
		imageAlt != "Carbon rim finish reference" || imageOrder != 0 {
		t.Fatalf(
			"unexpected G35 image: count=%d url=%q alt=%q order=%d",
			imageCount,
			imageURL,
			imageAlt,
			imageOrder,
		)
	}
}
