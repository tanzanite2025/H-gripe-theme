package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenGatewaySignsCountryListRequestAndAcceptsSuccessfulResponse(t *testing.T) {
	const userID = "yanwen-user"
	const apiToken = "yanwen-token"
	fixedTime := time.UnixMilli(1790000000123)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, userID, request.URL.Query().Get("user_id"))
		require.Equal(t, yanwenCountryListMethod, request.URL.Query().Get("method"))
		require.Equal(t, "json", request.URL.Query().Get("format"))
		require.Equal(t, yanwenAPIVersion, request.URL.Query().Get("version"))
		require.Equal(t, strconv.FormatInt(fixedTime.UnixMilli(), 10), request.URL.Query().Get("timestamp"))

		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{}`, string(body))
		raw := userID + string(body) + "json" + yanwenCountryListMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"code":0,"message":"ok","data":[]}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	client.now = func() time.Time { return fixedTime }
	latency, err := client.PingYanwenGateway(context.Background(), yanwenGatewayCredentials{
		environment: "fat",
		endpoint:    server.URL + "/api/order",
		userID:      userID,
		apiToken:    apiToken,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, latency, time.Duration(0))
}

func TestYanwenAPIServiceSavesEncryptedCredentialsAndPingsUsingStoredValues(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}))

	const userID = "stored-yanwen-user"
	const apiToken = "stored-yanwen-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		timestamp := request.URL.Query().Get("timestamp")
		raw := userID + string(body) + "json" + yanwenCountryListMethod + timestamp + yanwenAPIVersion
		require.Equal(t, userID, request.URL.Query().Get("user_id"))
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok"}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	apiService.gateway.now = func() time.Time { return fixedTime }
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "fat",
		Endpoint:    yanwenFATEndpoint,
		UserID:      userID,
		APIToken:    apiToken,
		Enabled:     true,
	}))

	stored, err := configRepo.FindYanwenAPIConfigByEnvironment("fat")
	require.NoError(t, err)
	require.NotContains(t, stored.UserIDEncrypted, userID)
	require.NotContains(t, stored.APITokenEncrypted, apiToken)
	require.NotEqual(t, userID, stored.UserIDEncrypted)
	require.NotEqual(t, apiToken, stored.APITokenEncrypted)

	result, err := apiService.PingYanwenGateway(context.Background(), YanwenAPIConfigInput{Environment: "fat"})
	require.NoError(t, err)
	require.True(t, result.OK)
	require.GreaterOrEqual(t, result.LatencyMS, int64(0))
}

func TestYanwenAPIServiceRejectsNonOfficialEnvironmentEndpoint(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}))
	apiService := NewYanwenAPIService(repository.NewYanwenAPIConfigRepository(db), nil)

	require.Error(t, validateYanwenEndpoint("http://open.yw56.com.cn/api/order"))
	require.Error(t, validateYanwenEndpoint("https://localhost/api/order"))
	require.NoError(t, validateYanwenEndpoint(yanwenFATEndpoint))
	require.NoError(t, validateYanwenEndpoint(yanwenDefaultEndpoint))
	require.Error(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "fat",
		Endpoint:    yanwenDefaultEndpoint,
		UserID:      "user",
		APIToken:    "token",
	}))
}

func TestParseYanwenGatewayResponseFailsLoudlyOnGatewayErrors(t *testing.T) {
	require.ErrorContains(t, parseYanwenGatewayResponse([]byte(`{"success":false,"code":"AUTH_401","message":"invalid signature"}`)), "invalid signature")
	require.ErrorContains(t, parseYanwenGatewayResponse([]byte(`{"success":true,"code":"500","message":"upstream error"}`)), "unexpected success code")
	require.ErrorContains(t, parseYanwenGatewayResponse([]byte(`{"success":true}`)), "unexpected success code")
	require.ErrorContains(t, parseYanwenGatewayResponse([]byte(`not-json`)), "decode Yanwen response")
	require.NoError(t, parseYanwenGatewayResponse([]byte(`{"success":true,"code":"0"}`)))
}

func TestParseYanwenProductResponseRequiresOfficialProductIdentity(t *testing.T) {
	products, err := parseYanwenProductResponse([]byte(`{"success":true,"code":"0","message":"ok","data":[{"id":"481","nameCh":"燕文专线追踪-普货","nameEn":"Direct Line Tracked Packet-P"},{"id":484,"nameCh":"燕文专线追踪-特货","nameEn":""}]}`))
	require.NoError(t, err)
	require.Equal(t, []shipping.YanwenProductCatalogEntry{
		{ProductID: "481", NameChinese: "燕文专线追踪-普货", NameEnglish: "Direct Line Tracked Packet-P"},
		{ProductID: "484", NameChinese: "燕文专线追踪-特货", NameEnglish: ""},
	}, products)

	_, err = parseYanwenProductResponse([]byte(`{"success":true,"code":"0","data":[{"nameCh":"缺少产品编号"}]}`))
	require.ErrorContains(t, err, "missing id")
	_, err = parseYanwenProductResponse([]byte(`{"success":true,"code":"0","data":[{"id":"481"}]}`))
	require.ErrorContains(t, err, "missing name")
}

func TestYanwenAPIServiceSyncsOfficialProductsAndUpdatesExistingCatalogEntries(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenProductCatalogEntry{}))

	const userID = "product-sync-user"
	const apiToken = "product-sync-token"
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCount++
		require.Equal(t, yanwenProductListMethod, request.URL.Query().Get("method"))
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		timestamp := request.URL.Query().Get("timestamp")
		raw := userID + string(body) + "json" + yanwenProductListMethod + timestamp + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		if requestCount == 1 {
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"id":"481","nameCh":"普货旧名称","nameEn":"Tracked Packet"},{"id":"484","nameCh":"特货产品","nameEn":"Sensitive Packet"}]}`))
			return
		}
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"id":"481","nameCh":"普货新名称","nameEn":"Tracked Packet Updated"},{"id":"484","nameCh":"特货产品","nameEn":"Sensitive Packet"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	productRepo := repository.NewYanwenProductCatalogRepository(db)
	apiService := NewYanwenAPIService(configRepo, productRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "fat",
		Endpoint:    yanwenFATEndpoint,
		UserID:      userID,
		APIToken:    apiToken,
		Enabled:     true,
	}))

	first, err := apiService.SyncYanwenOfficialProducts(context.Background(), YanwenAPIConfigInput{Environment: "fat"})
	require.NoError(t, err)
	require.Equal(t, YanwenProductSyncSummary{Environment: "fat", Scanned: 2, Added: 2, Updated: 0}, YanwenProductSyncSummary{
		Environment: first.Environment,
		Scanned:     first.Scanned,
		Added:       first.Added,
		Updated:     first.Updated,
	})

	second, err := apiService.SyncYanwenOfficialProducts(context.Background(), YanwenAPIConfigInput{Environment: "fat"})
	require.NoError(t, err)
	require.Equal(t, 2, second.Scanned)
	require.Equal(t, 0, second.Added)
	require.Equal(t, 2, second.Updated)
	require.Equal(t, 2, requestCount)

	entries, err := productRepo.FindYanwenProductCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "普货新名称", entries[0].NameChinese)
	require.Equal(t, "production", normalizeYanwenEnvironment("production"))
	productionEntries, err := productRepo.FindYanwenProductCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	require.Empty(t, productionEntries)
}

type yanwenTestRoundTripper struct {
	target *url.URL
}

func (r yanwenTestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	cloned := request.Clone(request.Context())
	cloned.URL.Scheme = r.target.Scheme
	cloned.URL.Host = r.target.Host
	cloned.Host = r.target.Host
	cloned.RequestURI = ""
	return http.DefaultTransport.RoundTrip(cloned)
}
