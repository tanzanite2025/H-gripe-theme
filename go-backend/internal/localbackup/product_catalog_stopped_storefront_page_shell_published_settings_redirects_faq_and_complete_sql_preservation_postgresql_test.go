package localbackup

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	faqdomain "commerce-platform/internal/domain/faq"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func seedStoppedStorefrontPublishedPageShellSettingsRedirectRulesAndProductFAQ(t *testing.T, database *sql.DB) {
	t.Helper()
	_, err := database.Exec(`
INSERT INTO settings(key,value,type,locale,"group",is_public) VALUES
('home_meta_title','Recovered SQL default SEO','string','en','seo',true),
('home_meta_description','private-page-shell-marker-must-not-leak','string','en','seo',false),
('youtube','https://social.example.invalid/recovered-sql-channel','string','en','social',true),
('title','Recovered SQL website profile','string','en','website_profile',true),
('website_name_title','Recovered SQL website name','string','en','website_name',true)
ON CONFLICT(key,locale) DO UPDATE SET value=excluded.value,"group"=excluded."group",is_public=excluded.is_public;
INSERT INTO product_specification_templates(name,slug,is_enabled) VALUES
('Recovered SQL specification navigation','recovery-public-navigation',true),
('private-page-shell-marker-must-not-leak','recovery-disabled-navigation',false)
ON CONFLICT(slug) DO UPDATE SET name=excluded.name,is_enabled=excluded.is_enabled;
UPDATE faq_pages SET route_path='' WHERE route_path='/products/:slug';`)
	require.NoError(t, err)
	bound, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	for _, rule := range []urlmanagementdomain.StorefrontRedirectRule{
		{SourcePath: "/recovered-published-product", TargetPath: "/products/product-backup-original", StatusCode: 308, State: "published", Reason: "private-page-shell-marker-must-not-leak"},
		{SourcePath: "/private-page-shell-marker-must-not-leak", TargetPath: "/products/product-backup-original", StatusCode: 301, State: "draft"},
	} {
		require.NoError(t, bound.Create(&rule).Error)
	}
	page := faqdomain.FAQPage{PageID: "recovery-product-page", RoutePath: "/products/:slug", Locale: "en", Title: "Recovered SQL product FAQ", Status: "active", RouteStatus: "current"}
	require.NoError(t, bound.Create(&page).Error)
	for _, item := range []faqdomain.FAQ{
		{PageID: page.PageID, Locale: "en", Question: "Recovered SQL public question", Answer: "<p>Recovered SQL published answer</p>", Status: "published"},
		{PageID: page.PageID, Locale: "en", Question: "private-page-shell-marker-must-not-leak", Answer: "private-page-shell-marker-must-not-leak", Status: "draft"},
	} {
		require.NoError(t, bound.Create(&item).Error)
	}
}

func verifyStoppedStorefrontPageShellPublicResponseUsesSeededSQLAndFiltersPrivateValues(t *testing.T, operation string, body []byte) {
	t.Helper()
	require.NotContains(t, string(body), "private-page-shell-marker-must-not-leak")
	expected := map[string]string{
		"redirects": "/recovered-published-product", "seo_home": "Recovered SQL default SEO",
		"social_settings": "recovered-sql-channel", "website_profile": "Recovered SQL website profile",
		"website_name": "Recovered SQL website name", "specification_templates": "recovery-public-navigation",
		"currency_policy": `"primary_currency":"USD"`, "product_page_faq": "Recovered SQL published answer",
	}
	if value, present := expected[operation]; present {
		require.Contains(t, string(body), value, operation)
	}
}

func TestStoppedStorefrontPageShellOperationsRejectAdditionalSelectorsAndUnsupportedServices(t *testing.T) {
	for operation := range stoppedStorefrontPageShellReadOnlyGETPaths {
		t.Run(operation, func(t *testing.T) {
			request := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: operation, Locale: "en", Country: "AQ", Currency: "USD"}
			require.NoError(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(request))
			request.Slug = "unbound-selector"
			require.Error(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(request))
			request.Slug, request.Keyword = "", "unbound-selector"
			require.Error(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(request))
		})
	}
	for _, operation := range []string{"payment_methods", "recommendations", "auth_profile", "analytics", "arbitrary_settings_group"} {
		require.Error(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: operation, Locale: "en", Country: "AQ", Currency: "USD"}))
	}
}

func TestPostgreSQLStoppedStorefrontPageShellDependenciesUsePublishedDatabaseValuesWithoutMutatingAnyTableOrSequence(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and the complete application migrations")
	}
	ctx := context.Background()
	configuration := createIndependentRecoveryTestConfiguration(t)
	configuration.DatabaseURL = connectionURL
	configuration.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	control, err := configuration.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(configuration.DatabaseName))
	require.NoError(t, err)
	defer func() {
		_, _ = control.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(configuration.DatabaseName)+` WITH (FORCE)`)
	}()
	database, err := configuration.connectDatabase(ctx, configuration.DatabaseName)
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(database, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDirectory), "postgres", driver)
	require.NoError(t, err)
	if err := migrator.Up(); err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	sourceCloseError, databaseCloseError := migrator.Close()
	require.NoError(t, sourceCloseError)
	require.NoError(t, databaseCloseError)
	database, err = configuration.connectDatabase(ctx, configuration.DatabaseName)
	require.NoError(t, err)
	defer database.Close()
	seedStoppedStorefrontPublishedPageShellSettingsRedirectRulesAndProductFAQ(t, database)
	configuration.StoppedProductLocalMediaVerification = &StoppedProductLocalMediaVerificationConfiguration{SharedBaseURL: "https://synthetic.example.test", LogoBaseURL: "https://synthetic.example.test"}
	manager := &Manager{Configuration: configuration}
	beforeCounts, beforeFingerprint, beforeMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	for operation := range stoppedStorefrontPageShellReadOnlyGETPaths {
		t.Run(operation, func(t *testing.T) {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			_, err = transaction.ExecContext(ctx, `SET LOCAL statement_timeout='15s'`)
			require.NoError(t, err)
			request := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: operation, Locale: "en", Country: "AQ", Currency: "USD"}
			status, body, err := manager.invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx, transaction, request)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			verifyStoppedStorefrontPageShellPublicResponseUsesSeededSQLAndFiltersPrivateValues(t, operation, body)
			require.NoError(t, transaction.Commit())
		})
	}
	afterCounts, afterFingerprint, afterMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	require.Equal(t, beforeCounts, afterCounts)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	require.Equal(t, beforeMigration, afterMigration)
	// A configured but hidden product FAQ must keep its genuine 404 semantics.
	_, err = database.Exec(`UPDATE faq_pages SET status='hidden' WHERE page_id='recovery-product-page'`)
	require.NoError(t, err)
	transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	require.NoError(t, err)
	defer transaction.Rollback()
	status, _, err := manager.invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx, transaction, StoppedProductReadOnlyStorefrontHTTPRequest{Operation: "product_page_faq", Locale: "en", Country: "AQ", Currency: "USD"})
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, status)
}
