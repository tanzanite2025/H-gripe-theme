package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func seedStoppedStorefrontApprovedPendingRejectedAndDeletedReviewPaginationBeforeSnapshots(t *testing.T, database *sql.DB) (uint, uint) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO users(email,username,password)
SELECT 'product-backup@example.invalid','product-backup-customer','synthetic-password'
WHERE NOT EXISTS(SELECT 1 FROM users WHERE email='product-backup@example.invalid');
INSERT INTO products(name,slug,locale,price_minor,currency)
SELECT 'Original product','product-backup-original','en',12345,'USD'
WHERE NOT EXISTS(SELECT 1 FROM products WHERE slug='product-backup-original' AND locale='en');
INSERT INTO products(name,slug,locale,parent_id,price_minor,currency)
SELECT 'Original translated product','product-backup-original','zh',id,12345,'USD' FROM products p
WHERE slug='product-backup-original' AND locale='en'
AND NOT EXISTS(SELECT 1 FROM products WHERE slug=p.slug AND locale='zh');
INSERT INTO reviews(product_id,user_id,rating,title,content,images,status,verified,reply_content,moderation_reason,created_at)
SELECT p.id,u.id,5,'<b>Recovered approved review '||number||'</b>','Recovered approved content',
CASE WHEN number=6 THEN '{broken-json' ELSE '["/uploads/products/product-original.png"]' END,
'approved',true,'Recovered merchant reply','private-review-metadata-must-not-leak','2026-01-01T00:00:00Z'
FROM products p CROSS JOIN users u CROSS JOIN generate_series(1,7) number
WHERE p.slug='product-backup-original' AND p.locale='en' AND u.email='product-backup@example.invalid';
INSERT INTO reviews(product_id,user_id,rating,title,content,status,deleted_at)
SELECT p.id,u.id,1,'private-review-metadata-must-not-leak',state,state,
CASE WHEN state='approved' THEN now() ELSE NULL END
FROM products p CROSS JOIN users u CROSS JOIN (VALUES('pending'),('rejected'),('approved')) states(state)
WHERE p.slug='product-backup-original' AND p.locale='en' AND u.email='product-backup@example.invalid';
INSERT INTO reviews(product_id,user_id,rating,title,content,status)
SELECT p.id,u.id,1,'other-product-review-must-not-leak','Other product','approved'
FROM products p CROSS JOIN users u WHERE p.slug='product-backup-original' AND p.locale='zh' AND u.email='product-backup@example.invalid';
INSERT INTO review_summaries(product_id,total_reviews,average_rating,rating_5_count)
SELECT id,7,5,7 FROM products WHERE slug='product-backup-original' AND locale='en'
ON CONFLICT(product_id) DO UPDATE SET total_reviews=7,average_rating=5,rating_5_count=7;
DELETE FROM review_summaries WHERE product_id IN (SELECT id FROM products WHERE slug='product-backup-original' AND locale='zh');`)
	require.NoError(t, err)
	var productID, missingSummaryProductID uint
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT id FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&productID))
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT id FROM products WHERE slug='product-backup-original' AND locale='zh'`).Scan(&missingSummaryProductID))
	return productID, missingSummaryProductID
}

func verifyStoppedStorefrontProductReviewResponseFiltersPrivateFieldsAndPreservesPagination(t *testing.T, request StoppedProductReadOnlyStorefrontHTTPRequest, body []byte) {
	t.Helper()
	require.NotContains(t, string(body), "private-review-metadata-must-not-leak")
	if request.Operation == "product_review_summary" {
		var summary struct {
			ProductID     uint    `json:"product_id"`
			TotalReviews  int     `json:"total_reviews"`
			AverageRating float64 `json:"average_rating"`
		}
		require.NoError(t, json.Unmarshal(body, &summary))
		require.Equal(t, request.ProductID, summary.ProductID)
		require.Contains(t, []int{0, 7}, summary.TotalReviews)
		if summary.TotalReviews == 7 {
			require.Equal(t, float64(5), summary.AverageRating)
		}
		return
	}
	require.NotContains(t, string(body), "other-product-review-must-not-leak")
	var result struct {
		Data       []map[string]json.RawMessage `json:"data"`
		Pagination struct {
			Page      int `json:"page"`
			PageSize  int `json:"page_size"`
			Total     int `json:"total"`
			TotalPage int `json:"total_page"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	require.NotNil(t, result.Data)
	require.Equal(t, request.Page, result.Pagination.Page)
	require.Equal(t, request.PageSize, result.Pagination.PageSize)
	require.Equal(t, 7, result.Pagination.Total)
	require.Equal(t, (7+request.PageSize-1)/request.PageSize, result.Pagination.TotalPage)
	expectedCount := 7 - (request.Page-1)*request.PageSize
	if expectedCount < 0 {
		expectedCount = 0
	}
	if expectedCount > request.PageSize {
		expectedCount = request.PageSize
	}
	require.Len(t, result.Data, expectedCount)
	for index, item := range result.Data {
		for _, field := range []string{"user_id", "order_id", "user", "product", "moderation_reason", "moderated_by", "moderated_at", "replied_by", "deleted_at"} {
			require.NotContains(t, item, field)
		}
		var title string
		require.NoError(t, json.Unmarshal(item["title"], &title))
		number := 7 - (request.Page-1)*request.PageSize - index
		require.Equal(t, fmt.Sprintf("Recovered approved review %d", number), title)
		if number == 6 {
			require.JSONEq(t, `"[]"`, string(item["images"]))
		}
		if number == 7 {
			require.Contains(t, string(item["images"]), "/uploads/products/product-original.png")
		}
	}
}

func TestStoppedStorefrontProductReviewSelectorsAreBoundedAndUnavailableForOtherOperations(t *testing.T) {
	valid := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: "product_reviews", Locale: "en", Country: "AQ", Currency: "USD", ProductID: 1, Page: 1, PageSize: 5}
	require.NoError(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(valid))
	for name, change := range map[string]func(*StoppedProductReadOnlyStorefrontHTTPRequest){
		"missing_product":     func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.ProductID = 0 },
		"zero_page":           func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.Page = 0 },
		"negative_page":       func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.Page = -1 },
		"excessive_page":      func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.Page = 1001 },
		"zero_page_size":      func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.PageSize = 0 },
		"excessive_page_size": func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.PageSize = 21 },
		"summary_with_pagination": func(request *StoppedProductReadOnlyStorefrontHTTPRequest) {
			request.Operation = "product_review_summary"
		},
		"other_operation_with_selectors": func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.Operation = "site_settings" },
		"review_with_slug":               func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.Slug = "wrong" },
		"review_with_keyword":            func(request *StoppedProductReadOnlyStorefrontHTTPRequest) { request.Keyword = "wrong" },
	} {
		t.Run(name, func(t *testing.T) {
			request := valid
			change(&request)
			require.Error(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(request))
		})
	}
	valid.Operation, valid.Page, valid.PageSize = "product_review_summary", 0, 0
	require.NoError(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(valid))
}

func TestPostgreSQLStoppedStorefrontProductReviewPagesAndSummariesPreserveAllTablesAndSequences(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and complete application migrations")
	}
	ctx := context.Background()
	configuration := createIndependentRecoveryTestConfiguration(t)
	configuration.DatabaseURL, configuration.DatabaseName = connectionURL, "local_backup_test_"+newTaskID()[:12]
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
	productID, missingSummaryProductID := seedStoppedStorefrontApprovedPendingRejectedAndDeletedReviewPaginationBeforeSnapshots(t, database)
	configuration.StoppedProductLocalMediaVerification = &StoppedProductLocalMediaVerificationConfiguration{SharedBaseURL: "https://synthetic.example.test", LogoBaseURL: "https://synthetic.example.test"}
	manager := &Manager{Configuration: configuration}
	beforeCounts, beforeFingerprint, beforeMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	for _, scenario := range []struct {
		Name      string
		Operation string
		ProductID uint
		Page      int
		PageSize  int
	}{
		{"first_page", "product_reviews", productID, 1, 5}, {"second_page_equal_timestamps", "product_reviews", productID, 2, 5},
		{"empty_page", "product_reviews", productID, 3, 5}, {"maximum_page_and_size", "product_reviews", productID, 1000, 20},
		{"existing_summary", "product_review_summary", productID, 0, 0}, {"missing_summary_no_insert", "product_review_summary", missingSummaryProductID, 0, 0},
	} {
		t.Run(scenario.Name, func(t *testing.T) {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			request := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: scenario.Operation, Locale: "en", Country: "AQ", Currency: "USD", ProductID: scenario.ProductID, Page: scenario.Page, PageSize: scenario.PageSize}
			require.NoError(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(request))
			status, body, err := manager.invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx, transaction, request)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			verifyStoppedStorefrontProductReviewResponseFiltersPrivateFieldsAndPreservesPagination(t, request, body)
			if scenario.Name == "missing_summary_no_insert" {
				require.Contains(t, string(body), `"total_reviews":0`)
			}
			if scenario.Name == "existing_summary" {
				require.Contains(t, string(body), `"total_reviews":7`)
			}
			require.NoError(t, transaction.Commit())
		})
	}
	afterCounts, afterFingerprint, afterMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	require.Equal(t, beforeCounts, afterCounts)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	require.Equal(t, beforeMigration, afterMigration)
}
