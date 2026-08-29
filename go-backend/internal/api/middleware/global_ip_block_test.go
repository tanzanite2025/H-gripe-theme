package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGlobalIPBlockerRejectsBlockedIPAndAllowsOtherIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&security.IPBlockRule{}))

	blockService := service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	_, err = blockService.Block(context.Background(), service.IPBlockRuleInput{
		CIDR:   "203.0.113.0/24",
		Reason: "test block",
	})
	require.NoError(t, err)

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(GlobalIPBlocker(blockService))
	router.GET("/probe", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	blockedRequest := httptest.NewRequest(http.MethodGet, "/probe", nil)
	blockedRequest.RemoteAddr = "203.0.113.19:12345"
	blockedResponse := httptest.NewRecorder()
	router.ServeHTTP(blockedResponse, blockedRequest)
	require.Equal(t, http.StatusForbidden, blockedResponse.Code)
	require.Contains(t, blockedResponse.Body.String(), "ip_blocked")
	require.Equal(t, "ip", blockedResponse.Header().Get("X-Access-Block"))

	allowedRequest := httptest.NewRequest(http.MethodGet, "/probe", nil)
	allowedRequest.RemoteAddr = "198.51.100.19:12345"
	allowedResponse := httptest.NewRecorder()
	router.ServeHTTP(allowedResponse, allowedRequest)
	require.Equal(t, http.StatusNoContent, allowedResponse.Code)
	require.True(t, strings.Contains(blockedResponse.Header().Get("Cache-Control"), "no-store"))
}

func TestGlobalIPBlockerReturnsUnavailableWhenInitialCacheCannotLoad(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&security.IPBlockRule{}))
	blockService := service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	require.NoError(t, sqlDB.Close())

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(GlobalIPBlocker(blockService))
	router.GET("/probe", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.RemoteAddr = "198.51.100.19:12345"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.Contains(t, response.Body.String(), "ip_block_unavailable")
}

func TestGlobalIPBlockerAllowsRecoveryPathsButFailsClosedElsewhere(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(GlobalIPBlocker(nil))
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/api/admin/security/ip-blocks", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/probe", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, path := range []string{"/health", "/api/admin/security/ip-blocks"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusNoContent, response.Code, path)
	}

	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.Contains(t, response.Body.String(), "ip_block_unavailable")
}
