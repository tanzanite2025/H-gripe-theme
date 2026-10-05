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

func TestParseYanwenWarehouseResponseRequiresOfficialWarehouseIdentity(t *testing.T) {
	warehouses, err := parseYanwenWarehouseResponse([]byte(`{"success":true,"code":"0","message":"ok","data":[{"code":"01","name":"北京燕文","area":"华北"},{"code":"02","name":"上海燕文","area":"华东"}]}`))
	require.NoError(t, err)
	require.Equal(t, []shipping.YanwenWarehouseCatalogEntry{
		{WarehouseCode: "01", Name: "北京燕文", Area: "华北"},
		{WarehouseCode: "02", Name: "上海燕文", Area: "华东"},
	}, warehouses)

	_, err = parseYanwenWarehouseResponse([]byte(`{"success":true,"code":"0","data":[{"name":"缺少仓库代码"}]}`))
	require.ErrorContains(t, err, "missing code")
	_, err = parseYanwenWarehouseResponse([]byte(`{"success":true,"code":"0","data":[{"code":"01","name":"北京燕文"},{"code":"01","name":"重复仓库"}]}`))
	require.ErrorContains(t, err, "duplicate code")
}

func TestYanwenAPIServiceSyncsOfficialWarehousesUsingCurrentEnvironmentCredentials(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWarehouseCatalogEntry{}))

	const userID = "warehouse-sync-user"
	const apiToken = "warehouse-sync-token"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenWarehouseListMethod, request.URL.Query().Get("method"))
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		require.JSONEq(t, `{}`, string(body))
		raw := userID + string(body) + "json" + yanwenWarehouseListMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"code":"01","name":"北京燕文","area":"华北"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	warehouseRepo := repository.NewYanwenWarehouseCatalogRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWarehouseCatalogRepository(warehouseRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "fat",
		Endpoint:    yanwenFATEndpoint,
		UserID:      userID,
		APIToken:    apiToken,
		Enabled:     true,
	}))

	summary, err := apiService.SyncYanwenOfficialWarehouses(context.Background(), YanwenAPIConfigInput{Environment: "fat"})
	require.NoError(t, err)
	require.Equal(t, 1, summary.Scanned)
	require.Equal(t, 1, summary.Added)
	entries, err := warehouseRepo.FindYanwenWarehouseCatalogEntriesByEnvironment("fat")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "01", entries[0].WarehouseCode)
}
