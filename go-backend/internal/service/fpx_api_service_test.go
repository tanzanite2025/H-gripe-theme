package service

import (
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFpxAPISyncChannelsFetchesSignedOfficialDirectoryAndPreservesEnabledState(t *testing.T) {
	t.Setenv(FpxAPIMasterKeyEnv, "test-fpx-master-key")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.FpxChannel{}, &shipping.FpxAPIConfig{}))

	const appKey = "test-app-key"
	const appSecret = "test-app-secret"
	const accessToken = "test-access-token"
	fixedTime := time.UnixMilli(1790000000123)
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCount++
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, fpxChannelMethod, request.URL.Query().Get("method"))
		require.Equal(t, appKey, request.URL.Query().Get("app_key"))
		require.Equal(t, accessToken, request.URL.Query().Get("access_token"))
		require.Equal(t, "json", request.URL.Query().Get("format"))
		require.Equal(t, fpxAPIVersion, request.URL.Query().Get("v"))
		require.Equal(t, "application/json;charset=utf-8", request.Header.Get("Content-Type"))

		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		var payload map[string]string
		require.NoError(t, json.Unmarshal(body, &payload))
		require.Equal(t, "1", payload["transport_mode"])

		params := map[string]string{
			"app_key":   appKey,
			"format":    "json",
			"method":    fpxChannelMethod,
			"timestamp": request.URL.Query().Get("timestamp"),
			"v":         fpxAPIVersion,
		}
		require.Len(t, request.URL.Query().Get("timestamp"), 13)
		require.Equal(t, strconv.FormatInt(fixedTime.UnixMilli(), 10), request.URL.Query().Get("timestamp"))
		require.Equal(t, signFpxRequest(params, body, appSecret), request.URL.Query().Get("sign"))

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"result":true,"data":[{"logistics_product_code":"ECO-TEST","logistics_product_name_cn":"4PX经济线"},{"logistics_product_code":"EXP-TEST","logistics_product_name_cn":"4PX快捷线"},{"logistics_product_code":"INCOMPLETE-TEST"},"unparseable-record"]}`))
	}))
	defer server.Close()

	configRepo := repository.NewFpxAPIConfigRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	apiService := NewFpxAPIService(configRepo, shippingRepo)
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	apiService.gateway.httpClient = &http.Client{Transport: fpxTestRoundTripper{target: target}}
	apiService.gateway.now = func() time.Time { return fixedTime }
	apiService.now = func() time.Time { return fixedTime }

	enabledExisting := &shipping.FpxChannel{ServiceCode: "ECO-TEST", DisplayName: "Old name", Enabled: true}
	require.NoError(t, shippingRepo.CreateFpxChannel(enabledExisting))

	require.NoError(t, apiService.Save(FpxAPIConfigInput{
		Environment: "production",
		Endpoint:    fpxDefaultEndpoint,
		AppKey:      appKey,
		AppSecret:   appSecret,
		AccessToken: accessToken,
		Enabled:     true,
	}))

	stored, err := configRepo.FindByEnvironment("production")
	require.NoError(t, err)
	require.NotContains(t, stored.AppKeyEncrypted, appKey)
	require.NotContains(t, stored.AppSecretEncrypted, appSecret)
	require.NotContains(t, stored.AccessTokenEncrypted, accessToken)

	summary, err := apiService.SyncChannels(context.Background(), FpxAPIConfigInput{Environment: "production"})
	require.NoError(t, err)
	require.Equal(t, FpxChannelSyncSummary{Scanned: 4, Added: 1, Updated: 1, PreservedEnabled: 1}, summary)
	require.Equal(t, 1, requestCount)

	channels, err := shippingRepo.FindAllFpxChannels(false)
	require.NoError(t, err)
	require.Len(t, channels, 2)
	require.Equal(t, "4PX经济线", channels[0].DisplayName)
	require.True(t, channels[0].Enabled, "sync must preserve a previously approved service")
	require.Equal(t, "EXP-TEST", channels[1].ServiceCode)
	require.False(t, channels[1].Enabled, "new official services must require explicit approval")

	stored, err = configRepo.FindByEnvironment("production")
	require.NoError(t, err)
	require.Equal(t, "success", stored.LastSyncStatus)
	require.NotNil(t, stored.LastSyncedAt)
	require.True(t, stored.LastSyncedAt.Equal(fixedTime.UTC()))
	require.Empty(t, stored.LastError)
	require.Equal(t, 4, stored.LastSyncScanned)
	require.Equal(t, 1, stored.LastSyncAdded)
	require.Equal(t, 1, stored.LastSyncUpdated)
	require.Equal(t, 1, stored.LastSyncPreservedEnabled)
}

func TestParseFpxChannelResponseRejectsGatewayAndMalformedResponses(t *testing.T) {
	_, err := parseFpxChannelResponse([]byte(`{"result":false,"message":"invalid signature","data":[]}`))
	require.ErrorContains(t, err, "invalid signature")
	_, err = parseFpxChannelResponse([]byte(`{"result":0,"error_code":"AUTH_102","error_msg":"IP not allowed","data":[]}`))
	require.ErrorContains(t, err, "AUTH_102")
	require.ErrorContains(t, err, "IP not allowed")
	_, err = parseFpxChannelResponse([]byte(`{"data":[]}`))
	require.ErrorContains(t, err, "unsuccessful result")
	_, err = parseFpxChannelResponse([]byte(`{"result":1}`))
	require.ErrorContains(t, err, "missing data")

	_, err = parseFpxChannelResponse([]byte(`{"errors":[{"code":"AUTH","message":"bad credentials"}],"data":[]}`))
	require.ErrorContains(t, err, "bad credentials")

	_, err = parseFpxChannelResponse([]byte(`not-json`))
	require.ErrorContains(t, err, "decode 4PX response")
}

func TestFpxAPIPingDoesNotPersistSyncStateOrChannels(t *testing.T) {
	t.Setenv(FpxAPIMasterKeyEnv, "test-fpx-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.FpxChannel{}, &shipping.FpxAPIConfig{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"result":1,"msg":"ok","errors":[]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewFpxAPIConfigRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	apiService := NewFpxAPIService(configRepo, shippingRepo)
	apiService.gateway.httpClient = &http.Client{Transport: fpxTestRoundTripper{target: target}}
	apiService.gateway.now = func() time.Time { return time.UnixMilli(1790000000123) }
	require.NoError(t, apiService.Save(FpxAPIConfigInput{
		Environment: "test",
		Endpoint:    fpxTestEndpoint,
		AppKey:      "test-app-key",
		AppSecret:   "test-app-secret",
		Enabled:     true,
	}))

	result, err := apiService.Ping(context.Background(), FpxAPIConfigInput{Environment: "test"})
	require.NoError(t, err)
	require.True(t, result.OK)
	require.GreaterOrEqual(t, result.LatencyMS, int64(0))

	stored, err := configRepo.FindByEnvironment("test")
	require.NoError(t, err)
	require.Empty(t, stored.LastSyncStatus)
	require.Nil(t, stored.LastSyncedAt)
	channels, err := shippingRepo.FindAllFpxChannels(false)
	require.NoError(t, err)
	require.Empty(t, channels)
}

func TestFpxConfigEndpointDefaultsByEnvironment(t *testing.T) {
	require.Equal(t, fpxDefaultEndpoint, fpxEndpointForEnvironment("production"))
	require.Equal(t, fpxTestEndpoint, fpxEndpointForEnvironment("test"))
	parsed, err := url.Parse(fpxTestEndpoint)
	require.NoError(t, err)
	require.Equal(t, "https", parsed.Scheme)
}

func TestFpxGatewayLogHelpersMaskCredentials(t *testing.T) {
	credentials := fpxGatewayCredentials{
		appKey:      "test-app-key",
		appSecret:   "test-app-secret",
		accessToken: "test-access-token",
	}
	message, gatewayErrors := fpxResponseLogDetails([]byte(`{"msg":"test-access-token","errors":[{"error_code":"AUTH_2","error_msg":"test-app-secret"}]}`))
	message = redactFpxCredentials(message, credentials)
	gatewayErrors = redactFpxCredentials(gatewayErrors, credentials)
	require.Equal(t, "[REDACTED]", message)
	require.Equal(t, "AUTH_2: [REDACTED]", gatewayErrors)
	require.Equal(t, "****", maskFpxCredential(credentials.appSecret))
	require.Equal(t, "****", maskFpxCredential(credentials.accessToken))
}

func TestFpxConfigSaveCanExplicitlyDisableSavedEnvironment(t *testing.T) {
	t.Setenv(FpxAPIMasterKeyEnv, "test-fpx-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.FpxAPIConfig{}))

	configRepo := repository.NewFpxAPIConfigRepository(db)
	apiService := NewFpxAPIService(configRepo, nil)
	require.NoError(t, apiService.Save(FpxAPIConfigInput{
		Environment: "test",
		Endpoint:    fpxTestEndpoint,
		AppKey:      "test-key",
		AppSecret:   "test-secret",
		Enabled:     true,
	}))
	require.NoError(t, apiService.Save(FpxAPIConfigInput{
		Environment: "test",
		Endpoint:    fpxTestEndpoint,
		Enabled:     false,
	}))

	stored, err := configRepo.FindByEnvironment("test")
	require.NoError(t, err)
	require.False(t, stored.Enabled)
	require.NotEmpty(t, stored.AppKeyEncrypted)
	require.NotEmpty(t, stored.AppSecretEncrypted)
}

func TestValidateFpxEndpointRejectsInvalidScheme(t *testing.T) {
	require.Error(t, validateFpxEndpoint("file:///etc/passwd"))
	require.Error(t, validateFpxEndpoint("http:/missing-host"))
	require.Error(t, validateFpxEndpoint("https://localhost/router/api/service"))
	require.NoError(t, validateFpxEndpoint("https://open.4px.com/router/api/service"))
}

type fpxTestRoundTripper struct {
	target *url.URL
}

func (r fpxTestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	cloned := request.Clone(request.Context())
	cloned.URL.Scheme = r.target.Scheme
	cloned.URL.Host = r.target.Host
	cloned.Host = r.target.Host
	cloned.RequestURI = ""
	return http.DefaultTransport.RoundTrip(cloned)
}
