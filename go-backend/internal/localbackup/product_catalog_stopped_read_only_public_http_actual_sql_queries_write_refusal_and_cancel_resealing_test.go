package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func verifyStoppedProductReadOnlyActualPublicHTTPQueriesPreserveSQLRedisAndSealedDatabases(t *testing.T, manager *Manager, taskID string, redisOptions *redis.Options, invokeHTTP func(string, any, int) *httptest.ResponseRecorder) {
	t.Helper()
	ctx := context.Background()
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	baseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	var active ProductCatalogStoppedStorefrontRouteExpectation
	for _, route := range baseline.FinalRoutes {
		if route.PublicActive && route.Slug == "product-backup-original" && route.Locale == "en" {
			active = route
			break
		}
	}
	require.NotEmpty(t, active.ProductID, "fixture must contain an active public product")
	client := redis.NewClient(redisOptions)
	keysBefore, err := client.Keys(ctx, "*").Result() // Disposable test Redis only.
	require.NoError(t, err)
	require.NoError(t, client.Close())
	for _, query := range []StoppedProductReadOnlyPublicHTTPRequest{
		{Operation: "detail", Slug: active.Slug, Locale: active.Locale},
		{Operation: "list", Locale: active.Locale},
		{Operation: "search", Locale: active.Locale, Keyword: "product"},
		{Operation: "detail", Slug: "synthetic-definitely-absent-read-only-product", Locale: active.Locale},
	} {
		query.ContractSHA256 = baseline.ContractSHA256
		var response struct {
			Result StoppedProductReadOnlyPublicHTTPResult
		}
		require.NoError(t, json.Unmarshal(invokeHTTP("product-catalog-stopped-read-only-product-http-query", query, http.StatusOK).Body.Bytes(), &response))
		expectedStatus := http.StatusOK
		if strings.Contains(query.Slug, "definitely-absent") {
			expectedStatus = http.StatusNotFound
		}
		require.Equal(t, expectedStatus, response.Result.HTTPStatus, string(response.Result.Response))
		require.Equal(t, baseline.ContractSHA256, response.Result.ContractSHA256)
		require.Equal(t, query, response.Result.Request)
		require.Equal(t, productCatalogBytesSHA256(response.Result.Response), response.Result.ResponseSHA256)
		require.True(t, response.Result.SQLReadOnly)
		require.True(t, response.Result.ApplicationRemainsStopped)
		require.False(t, response.Result.OnlineOverwriteAvailable)
		require.NotEmpty(t, response.Result.RemainingChecks)
		if query.Operation == "detail" && expectedStatus == http.StatusOK {
			var detail struct {
				Data struct {
					ID                       json.Number
					Name, Slug, Availability string
					PriceDecimal             string `json:"price_decimal"`
				}
			}
			// The public contract deliberately excludes exact inventory counters.
			var publicData map[string]any
			require.NoError(t, json.Unmarshal(response.Result.Response, &publicData))
			require.NoError(t, json.Unmarshal(response.Result.Response, &detail))
			require.Equal(t, active.ProductID, detail.Data.ID.String())
			require.Equal(t, active.Slug, detail.Data.Slug)
			require.Equal(t, "Original product", detail.Data.Name, "must read the recovered candidate rather than the rollback database")
			require.Equal(t, "123.45", detail.Data.PriceDecimal)
			require.Equal(t, "in_stock", detail.Data.Availability)
			require.NotContains(t, publicData["data"].(map[string]any), "view_count")
			require.NotContains(t, publicData["data"].(map[string]any), "stock")
		}
	}
	stale := StoppedProductReadOnlyPublicHTTPRequest{ContractSHA256: strings.Repeat("b", 64), Operation: "list", Locale: active.Locale}
	invokeHTTP("product-catalog-stopped-read-only-product-http-query", stale, http.StatusConflict)
	assertSealed := func() {
		control, err := manager.Configuration.connectDatabase(ctx, "postgres")
		require.NoError(t, err)
		defer control.Close()
		var allowed int
		require.NoError(t, control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE oid=ANY($1) AND datallowconn`, pq.Array([]int64{before.Journal.OldDatabaseOID, before.Journal.NewDatabaseOID})).Scan(&allowed))
		require.Zero(t, allowed)
	}
	assertSealed()
	// A real UPDATE is rejected by PostgreSQL, not simply omitted by the UI.
	err = manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, before.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE products SET view_count=view_count+1`)
		var databaseError *pq.Error
		require.ErrorAs(t, err, &databaseError)
		require.Equal(t, pq.ErrorCode("25006"), databaseError.Code)
		return err
	})
	require.Error(t, err)
	assertSealed()
	canceled, cancel := context.WithCancel(ctx)
	err = manager.withBoundStoppedProductReadOnlyTransaction(canceled, manager.Configuration.DatabaseName, before.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error { cancel(); return canceled.Err() })
	require.ErrorIs(t, err, context.Canceled)
	assertSealed()
	require.Panics(t, func() {
		_ = manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, before.ProductCatalogPreviewBinding.Candidate, func(*sql.Tx) error { panic("synthetic read callback interruption") })
	})
	assertSealed()
	after, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, after, "read-only queries must not save task state, ownership or approval")
	current, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, baseline.ContractSHA256, current.ContractSHA256, "all SQL/media/views/sequences must remain unchanged")
	client = redis.NewClient(redisOptions)
	keysAfter, err := client.Keys(ctx, "*").Result()
	require.NoError(t, err)
	require.NoError(t, client.Close())
	require.ElementsMatch(t, keysBefore, keysAfter)
	client = redis.NewClient(redisOptions)
	require.NoError(t, client.Set(ctx, "product:7654322", "synthetic-read-only-query-boundary-change", 0).Err())
	require.NoError(t, client.Close())
	stale.ContractSHA256 = baseline.ContractSHA256
	invokeHTTP("product-catalog-stopped-read-only-product-http-query", stale, http.StatusConflict)
	client = redis.NewClient(redisOptions)
	require.Equal(t, "synthetic-read-only-query-boundary-change", client.Get(ctx, "product:7654322").Val())
	require.NoError(t, client.Del(ctx, "product:7654322").Err())
	require.NoError(t, client.Close())
	assertSealed()
}

func TestStoppedProductReadOnlyPublicHTTPQueriesRejectArbitraryOperationsAndUnboundedArguments(t *testing.T) {
	valid := StoppedProductReadOnlyPublicHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: "detail", Locale: "en", Slug: "actual-product"}
	require.NoError(t, validateStoppedProductReadOnlyPublicHTTPRequest(valid))
	for _, scenario := range []string{"sql", "url", "missing_slug", "missing_locale", "bad_contract", "bad_locale", "oversized_locale", "oversized_slug", "oversized_keyword", "control", "search_missing_keyword", "list_extra_slug", "detail_extra_keyword"} {
		t.Run(scenario, func(t *testing.T) {
			request := valid
			switch scenario {
			case "sql":
				request.Operation = "UPDATE products SET stock=0"
			case "url":
				request.Operation = "https://example.invalid/products"
			case "missing_slug":
				request.Slug = ""
			case "missing_locale":
				request.Locale = ""
			case "bad_contract":
				request.ContractSHA256 = strings.Repeat("0", 64)
			case "bad_locale":
				request.Locale = "en/../private"
			case "oversized_locale":
				request.Locale = strings.Repeat("e", 33)
			case "oversized_slug":
				request.Slug = strings.Repeat("a", 513)
			case "oversized_keyword":
				request.Operation, request.Slug, request.Keyword = "search", "", strings.Repeat("a", 201)
			case "control":
				request.Slug = "product\n"
			case "search_missing_keyword":
				request.Operation, request.Slug = "search", ""
			case "list_extra_slug":
				request.Operation = "list"
			case "detail_extra_keyword":
				request.Keyword = "extra"
			}
			require.Error(t, validateStoppedProductReadOnlyPublicHTTPRequest(request))
		})
	}
}
