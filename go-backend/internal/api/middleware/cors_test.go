package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCORSAllowsIdempotencyKeyAndExposesReplayMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS(config.CORSConfig{
		AllowedOrigins: []string{"https://store.example.test"},
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
		AllowedHeaders: []string{
			"Content-Type",
		},
		AllowCredentials: true,
	}))
	router.POST("/api/v1/orders", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/orders", nil)
	request.Header.Set("Origin", "https://store.example.test")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type, idempotency-key, x-anonymous-id")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "https://store.example.test", response.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, response.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key")
	require.Contains(t, response.Header().Get("Access-Control-Allow-Headers"), "X-Anonymous-ID")
	require.Contains(t, response.Header().Get("Access-Control-Expose-Headers"), "Idempotency-Replayed")
	require.Contains(t, response.Header().Get("Access-Control-Expose-Headers"), "Retry-After")
}
