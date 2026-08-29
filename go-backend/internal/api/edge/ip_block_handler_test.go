package edge

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestIPBlockHandlerReturnsNoContentForAllowedClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, blockService := newTestIPBlockHandlerService(t)
	handler := NewIPBlockHandler(blockService)

	router := gin.New()
	router.GET(ipBlockCheckPath, handler.Check)
	request := httptest.NewRequest(http.MethodGet, ipBlockCheckPath, nil)
	request.RemoteAddr = "198.51.100.10:1234"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Empty(t, response.Body.String())
}

func TestIPBlockHandlerReturnsForbiddenForBlockedClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, blockService := newTestIPBlockHandlerService(t)
	_, err := blockService.Block(context.Background(), service.IPBlockRuleInput{
		CIDR:   "198.51.100.0/24",
		Reason: "edge check test",
	})
	require.NoError(t, err)
	handler := NewIPBlockHandler(blockService)

	router := gin.New()
	router.GET(ipBlockCheckPath, handler.Check)
	request := httptest.NewRequest(http.MethodGet, ipBlockCheckPath, nil)
	request.RemoteAddr = "198.51.100.10:1234"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusForbidden, response.Code)
	require.Equal(t, "ip", response.Header().Get("X-Access-Block"))
}

func TestIPBlockHandlerReturnsUnavailableWhenCheckCannotLoadCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, blockService := newTestIPBlockHandlerService(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	handler := NewIPBlockHandler(blockService)

	router := gin.New()
	router.GET(ipBlockCheckPath, handler.Check)
	request := httptest.NewRequest(http.MethodGet, ipBlockCheckPath, nil)
	request.RemoteAddr = "198.51.100.10:1234"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestIPBlockHandlerReturnsUnavailableWithoutService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewIPBlockHandler(nil)

	router := gin.New()
	router.GET(ipBlockCheckPath, handler.Check)
	request := httptest.NewRequest(http.MethodGet, ipBlockCheckPath, nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func newTestIPBlockHandlerService(t *testing.T) (*gorm.DB, *service.GlobalIPBlockService) {
	t.Helper()

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

	return db, service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
}
