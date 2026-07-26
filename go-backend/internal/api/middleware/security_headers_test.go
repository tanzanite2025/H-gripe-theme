package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeadersIncludesLaunchHardeningHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil))

	headers := recorder.Result().Header
	if headers.Get("Content-Security-Policy") == "" {
		t.Fatalf("expected Content-Security-Policy header")
	}
	if headers.Get("Referrer-Policy") == "" {
		t.Fatalf("expected Referrer-Policy header")
	}
	if headers.Get("Permissions-Policy") == "" {
		t.Fatalf("expected Permissions-Policy header")
	}
}
