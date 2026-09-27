package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFpxAPIConfigGetFailsLoudlyOnDatabaseError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	handler := NewFpxAPIHandler(service.NewFpxAPIService(
		repository.NewFpxAPIConfigRepository(db),
		nil,
	))
	router := gin.New()
	router.GET("/config", handler.Get)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/config?environment=production", nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "no such table")
	require.NotContains(t, response.Body.String(), `"enabled":false`)
}

func TestFpxAPIConfigGetReturnsEmptyViewWhenNotConfigured(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.FpxAPIConfig{}))

	handler := NewFpxAPIHandler(service.NewFpxAPIService(
		repository.NewFpxAPIConfigRepository(db),
		nil,
	))
	router := gin.New()
	router.GET("/config", handler.Get)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/config?environment=test", nil))

	require.Equal(t, http.StatusOK, response.Code)
	var result struct {
		Data shipping.FpxAPIConfigView `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, "test", result.Data.Environment)
	require.Equal(t, "https://open-test.4px.com/router/api/service", result.Data.Endpoint)
	require.False(t, result.Data.Enabled)
}
