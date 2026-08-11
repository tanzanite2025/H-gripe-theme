package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"commerce-platform/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWaitPaymentRiskDelayStopsWhenRequestIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	startedAt := time.Now()
	ok := waitPaymentRiskDelay(ctx, time.Minute)

	require.False(t, ok)
	require.Less(t, time.Since(startedAt), 100*time.Millisecond)
}

func TestStripeRequestCountryRequiresTrustedEdgeMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	trustedEdge, err := middleware.NewTrustedEdgeMetadata([]string{"10.0.0.0/8"})
	require.NoError(t, err)

	router := gin.New()
	router.Use(trustedEdge)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"country": stripeRequestCountry(c)})
	})

	untrusted := httptest.NewRequest(http.MethodGet, "/", nil)
	untrusted.RemoteAddr = "203.0.113.20:443"
	untrusted.Header.Set("CF-IPCountry", "US")
	untrustedRecorder := httptest.NewRecorder()
	router.ServeHTTP(untrustedRecorder, untrusted)

	require.Equal(t, http.StatusOK, untrustedRecorder.Code)
	require.JSONEq(t, `{"country":""}`, untrustedRecorder.Body.String())

	trusted := httptest.NewRequest(http.MethodGet, "/", nil)
	trusted.RemoteAddr = "10.1.0.8:443"
	trusted.Header.Set("CF-IPCountry", "US")
	trustedRecorder := httptest.NewRecorder()
	router.ServeHTTP(trustedRecorder, trusted)

	require.Equal(t, http.StatusOK, trustedRecorder.Code)
	require.JSONEq(t, `{"country":"US"}`, trustedRecorder.Body.String())
}
