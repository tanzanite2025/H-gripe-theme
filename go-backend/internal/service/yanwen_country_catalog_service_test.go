package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestParseYanwenCountryResponseRequiresOfficialCountryIdentity(t *testing.T) {
	countries, err := parseYanwenCountryResponse([]byte(`{"success":true,"code":"0","message":"ok","data":[{"id":"1","code":"aw","nameCh":"阿鲁巴","nameEn":"ARUBA"},{"id":45,"code":"CA","nameCh":"加拿大","nameEn":"CANADA"}]}`))
	require.NoError(t, err)
	require.Equal(t, []shipping.YanwenCountryCatalogEntry{
		{CountryID: "1", CountryCode: "AW", NameChinese: "阿鲁巴", NameEnglish: "ARUBA"},
		{CountryID: "45", CountryCode: "CA", NameChinese: "加拿大", NameEnglish: "CANADA"},
	}, countries)

	_, err = parseYanwenCountryResponse([]byte(`{"success":true,"code":"0","data":[{"id":"1","nameCh":"缺少国家代码"}]}`))
	require.ErrorContains(t, err, "missing code")
	_, err = parseYanwenCountryResponse([]byte(`{"success":true,"code":"0","data":[{"id":"1","code":"AW","nameCh":"阿鲁巴"},{"id":"2","code":"aw","nameCh":"重复代码"}]}`))
	require.ErrorContains(t, err, "duplicate code")
}

func TestYanwenAPIServiceSyncsOfficialCountriesUsingCurrentEnvironmentCredentials(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenCountryCatalogEntry{}))

	const userID = "country-sync-user"
	const apiToken = "country-sync-token"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenCountryListMethod, request.URL.Query().Get("method"))
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		raw := userID + string(body) + "json" + yanwenCountryListMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"id":"1","code":"AW","nameCh":"阿鲁巴","nameEn":"ARUBA"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	countryRepo := repository.NewYanwenCountryCatalogRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenCountryCatalogRepository(countryRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "fat",
		Endpoint:    yanwenFATEndpoint,
		UserID:      userID,
		APIToken:    apiToken,
		Enabled:     true,
	}))

	summary, err := apiService.SyncYanwenOfficialCountries(context.Background(), YanwenAPIConfigInput{Environment: "fat"})
	require.NoError(t, err)
	require.Equal(t, 1, summary.Scanned)
	require.Equal(t, 1, summary.Added)
	entries, err := countryRepo.FindYanwenCountryCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "AW", entries[0].CountryCode)
}
