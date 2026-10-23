package localbackup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

func seedStoppedStorefrontDatabaseMarketSettingsAndWeightedProductShippingBeforeSnapshots(t *testing.T, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO settings(key,value,type,locale,"group",is_public) VALUES
('site_name','Recovered storefront from SQL','string','en','site',true),
('recovery_private_configuration_marker','synthetic-private-setting-must-not-leak','string','en','site',false)
ON CONFLICT (key,locale) DO UPDATE SET value=excluded.value,is_public=excluded.is_public;
INSERT INTO storefront_markets(code,name,default_locale,supported_locales,default_currency,display_currencies,enabled,priority)
VALUES('RECOVERY_SQL','Recovered SQL market','en','["en"]','USD','["USD"]',true,1);
INSERT INTO storefront_market_countries(market_id,code) SELECT id,'AQ' FROM storefront_markets WHERE code='RECOVERY_SQL';
UPDATE product_variants SET weight_grams=900 WHERE sku='product-backup-sku';`)
	require.NoError(t, err)
	bound, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	template := shippingdomain.ShippingTemplate{Name: "Stopped recovery real shipping template", Type: "weight", Currency: "USD", Enabled: true, DefaultFeeMinor: 9900,
		CustomerBillingMode: "actual_weight", CustomerFirstWeightGrams: 500, CustomerAdditionalWeightGrams: 500, CustomerEtaMinDays: 3, CustomerEtaMaxDays: 7,
		Rules: []shippingdomain.ShippingRule{{Region: "AQ", MinValue: 0, MaxValue: 2, FeeMinor: 500, Currency: "USD"}},
	}
	require.NoError(t, bound.Create(&template).Error)
	carrier := shippingdomain.Carrier{Name: "Recovery fixture carrier", Code: "RECOVERY_SQL", Enabled: true}
	require.NoError(t, bound.Create(&carrier).Error)
	carrierService := shippingdomain.CarrierService{CarrierID: carrier.ID, ServiceCode: "RECOVERY-SQL", ServiceName: "Recovery SQL published route", Countries: `["AQ"]`, Currency: "USD", Enabled: true}
	require.NoError(t, bound.Create(&carrierService).Error)
	routeGroup := shippingdomain.ShippingRouteGroup{Name: "Recovery SQL published route", Enabled: true}
	shippingRepository := repository.NewShippingRepository(bound)
	require.NoError(t, shippingRepository.CreateShippingRouteGroup(&routeGroup, []uint{carrierService.ID}))
	template.RouteGroupID = &routeGroup.ID
	require.NoError(t, shippingRepository.UpdateTemplateWithRules(&template, template.Rules))
	var productID, variantID uint
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT product_id,id FROM product_variants WHERE sku='product-backup-sku'`).Scan(&productID, &variantID))
	require.NoError(t, repository.NewProductRepository(bound).ReplaceSkuShippingTemplateBindings(productID, variantID, []uint{template.ID}))
}

// Exercise newly composed dependencies before the longer backup/switch drill,
// including an absent review-summary row and an actual weighted shipping route.
func verifyStoppedStorefrontComposedDependenciesInReadOnlySQLBeforeBackupRoundTrip(t *testing.T, configuration Configuration, database *sql.DB) {
	t.Helper()
	configuration.StoppedProductLocalMediaVerification = &StoppedProductLocalMediaVerificationConfiguration{SharedBaseURL: "https://synthetic.example.test", LogoBaseURL: "https://synthetic.example.test"}
	manager := &Manager{Configuration: configuration}
	ctx := context.Background()
	tx, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `SET LOCAL statement_timeout='15s'`)
	require.NoError(t, err)
	var before, after int
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT (SELECT count(*) FROM review_summaries)+(SELECT count(*) FROM shipping_quote_snapshots)`).Scan(&before))
	for _, operation := range []string{"detail", "context", "site_settings", "categories", "list", "search", "redirects", "seo_home", "social_settings", "website_profile", "website_name", "specification_templates", "currency_policy", "product_page_faq"} {
		query := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: operation, Locale: "en", Country: "AQ", Currency: "USD"}
		if operation == "detail" {
			query.Slug = "product-backup-original"
		}
		if operation == "search" {
			query.Keyword = "product"
		}
		status, body, err := manager.invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx, tx, query)
		require.NoError(t, err, operation)
		require.Equal(t, http.StatusOK, status, operation)
		verifyStoppedStorefrontPageShellPublicResponseUsesSeededSQLAndFiltersPrivateValues(t, operation, body)
		if operation == "detail" {
			require.Contains(t, string(body), `"amount_decimal":"5.00"`)
		}
	}
	var missingSummaryProductID uint
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT id FROM products WHERE slug='product-backup-original' AND locale='zh'`).Scan(&missingSummaryProductID))
	status, body, err := manager.invokeStoppedProductStorefrontGETWithBoundReadOnlyTransaction(ctx, tx, StoppedProductReadOnlyStorefrontHTTPRequest{Operation: "product_review_summary", Locale: "en", Country: "AQ", Currency: "USD", ProductID: missingSummaryProductID})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, string(body), `"total_reviews":0`)
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT (SELECT count(*) FROM review_summaries)+(SELECT count(*) FROM shipping_quote_snapshots)`).Scan(&after))
	require.Equal(t, before, after)
}

func verifyStoppedStorefrontActualNetworkHTTPUsesRecoveredSQLWithoutWritingOrReleasing(t *testing.T, manager *Manager, taskID string) {
	t.Helper()
	ctx := context.Background()
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	baseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	key, err := readProtectedCredential(manager.Configuration.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, manager.Configuration.SiteID, manager.Configuration.Environment, "synthetic-storefront-network-verifier", []string{PermissionRestore})
	require.NoError(t, err)
	server := httptest.NewServer(manager.HTTPHandler())
	defer server.Close()
	client := &http.Client{Timeout: 2 * time.Minute}
	invokeStoppedStorefrontActualNetworkHTTPQuery := func(queryTest *testing.T, body any, expectedStatus int) StoppedProductReadOnlyStorefrontHTTPResult {
		queryTest.Helper()
		encoded, err := json.Marshal(body)
		require.NoError(queryTest, err)
		request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/api/tasks/"+taskID+"/product-catalog-stopped-read-only-storefront-http-query", bytes.NewReader(encoded))
		require.NoError(queryTest, err)
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		require.NoError(queryTest, err)
		defer response.Body.Close()
		bodyBytes, err := io.ReadAll(io.LimitReader(response.Body, 6<<20))
		require.NoError(queryTest, err)
		require.Equal(queryTest, expectedStatus, response.StatusCode, string(bodyBytes))
		var result struct {
			Result StoppedProductReadOnlyStorefrontHTTPResult
		}
		if expectedStatus == http.StatusOK {
			require.NoError(queryTest, json.Unmarshal(bodyBytes, &result))
		}
		return result.Result
	}
	for _, operation := range []string{"context", "site_settings", "categories", "detail", "list", "search", "missing_product", "redirects", "seo_home", "social_settings", "website_profile", "website_name", "specification_templates", "currency_policy", "product_page_faq", "product_reviews", "product_reviews_page_two", "product_reviews_empty_page", "product_review_summary", "product_review_summary_missing"} {
		t.Run("storefront_actual_network_"+operation, func(t *testing.T) {
			query := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: operation, Locale: "en", Country: "AQ", Currency: "USD"}
			if operation == "detail" {
				query.Slug = "product-backup-original"
			}
			if operation == "search" {
				query.Keyword = "product"
			}
			if operation == "missing_product" {
				query.Operation, query.Slug = "detail", "synthetic-missing-storefront-product"
			}
			if strings.HasPrefix(operation, "product_reviews") || strings.HasPrefix(operation, "product_review_summary") {
				locale := "en"
				if operation == "product_review_summary_missing" {
					locale = "zh"
				}
				for _, route := range baseline.FinalRoutes {
					if route.Slug == "product-backup-original" && route.Locale == locale {
						identity, err := strconv.ParseUint(route.ProductID, 10, 32)
						require.NoError(t, err)
						query.ProductID = uint(identity)
					}
				}
				require.NotZero(t, query.ProductID)
				if strings.HasPrefix(operation, "product_reviews") {
					query.Operation, query.Page, query.PageSize = "product_reviews", 1, 5
					if operation == "product_reviews_page_two" {
						query.Page = 2
					}
					if operation == "product_reviews_empty_page" {
						query.Page = 3
					}
				} else {
					query.Operation = "product_review_summary"
				}
			}
			result := invokeStoppedStorefrontActualNetworkHTTPQuery(t, query, http.StatusOK)
			require.Equal(t, query, result.Request)
			require.Equal(t, baseline.ContractSHA256, result.ContractSHA256)
			require.Equal(t, baseline.DatabaseOID, result.DatabaseOID)
			require.Equal(t, baseline.WindowID, result.WindowID)
			require.Equal(t, productCatalogBytesSHA256(result.ResponseBytesBase64), result.ResponseSHA256)
			require.True(t, result.SQLReadOnly)
			require.True(t, result.ApplicationRemainsStopped)
			require.False(t, result.OnlineOverwriteAvailable)
			require.NotContains(t, string(result.ResponseBytesBase64), "synthetic-private-setting-must-not-leak")
			if operation == "missing_product" {
				require.Equal(t, http.StatusNotFound, result.HTTPStatus)
				return
			}
			require.Equal(t, http.StatusOK, result.HTTPStatus, string(result.ResponseBytesBase64))
			verifyStoppedStorefrontPageShellPublicResponseUsesSeededSQLAndFiltersPrivateValues(t, operation, result.ResponseBytesBase64)
			if isStoppedStorefrontProductReviewOperation(query.Operation) {
				verifyStoppedStorefrontProductReviewResponseFiltersPrivateFieldsAndPreservesPagination(t, query, result.ResponseBytesBase64)
				if operation == "product_review_summary_missing" {
					require.Contains(t, string(result.ResponseBytesBase64), `"total_reviews":0`)
				}
				if operation == "product_review_summary" {
					require.Contains(t, string(result.ResponseBytesBase64), `"total_reviews":7`)
				}
			}
			if operation == "site_settings" {
				require.Contains(t, string(result.ResponseBytesBase64), "Recovered storefront from SQL")
			}
			if operation == "context" || operation == "detail" || operation == "list" {
				require.Contains(t, string(result.ResponseBytesBase64), "RECOVERY_SQL", "must use the actual configured market rather than the default global fallback")
			}
			if operation == "detail" {
				var detail struct {
					Data struct {
						ShippingDetails struct {
							Country       string
							AmountDecimal string `json:"amount_decimal"`
							EtaMinDays    int    `json:"eta_min_days"`
							EtaMaxDays    int    `json:"eta_max_days"`
						} `json:"shipping_details"`
						Breadcrumb    json.RawMessage `json:"breadcrumb"`
						ReviewSummary json.RawMessage `json:"review_summary"`
					}
				}
				require.NoError(t, json.Unmarshal(result.ResponseBytesBase64, &detail))
				require.Equal(t, "AQ", detail.Data.ShippingDetails.Country)
				require.Equal(t, "5.00", detail.Data.ShippingDetails.AmountDecimal)
				require.Equal(t, 3, detail.Data.ShippingDetails.EtaMinDays)
				require.Equal(t, 7, detail.Data.ShippingDetails.EtaMaxDays)
				require.NotEmpty(t, detail.Data.Breadcrumb)
				require.NotEmpty(t, detail.Data.ReviewSummary)
			}
		})
	}
	invokeStoppedStorefrontActualNetworkHTTPQuery(t, map[string]string{"contract_sha256": baseline.ContractSHA256, "sql": "UPDATE products SET stock=0"}, http.StatusConflict)
	invokeStoppedStorefrontActualNetworkHTTPQuery(t, StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("b", 64), Operation: "context", Locale: "en", Country: "AQ", Currency: "USD"}, http.StatusConflict)
	after, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, after)
	current, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, baseline.ContractSHA256, current.ContractSHA256, "including weighted shipping must preserve all SQL rows/sequences, media, Redis and the stopped lease")
}

func TestStoppedStorefrontQueryRequiresExplicitMarketAndRejectsUnboundedTargets(t *testing.T) {
	valid := StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: "context", Locale: "en", Country: "AQ", Currency: "USD"}
	require.NoError(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(valid))
	for _, scenario := range []string{"missing_country", "lowercase_country", "oversized_country", "missing_currency", "lowercase_currency", "extra_slug", "unknown_operation"} {
		t.Run(scenario, func(t *testing.T) {
			query := valid
			switch scenario {
			case "missing_country":
				query.Country = ""
			case "lowercase_country":
				query.Country = "aq"
			case "oversized_country":
				query.Country = "AQQ"
			case "missing_currency":
				query.Currency = ""
			case "lowercase_currency":
				query.Currency = "usd"
			case "extra_slug":
				query.Slug = "forged"
			case "unknown_operation":
				query.Operation = "orders"
			}
			require.Error(t, validateStoppedProductReadOnlyStorefrontHTTPRequest(query))
		})
	}
}
