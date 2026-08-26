package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWebsiteNameHandlerGetRejectsUnsupportedLocale(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewWebsiteNameHandler(
		service.NewWebsiteNameService(
			service.NewSettingService(nil, nil, 0),
		),
	)
	router.GET("/api/admin/settings/website-name", handler.Get)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/settings/website-name?locale=not-a-language",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "unsupported locale")
}
