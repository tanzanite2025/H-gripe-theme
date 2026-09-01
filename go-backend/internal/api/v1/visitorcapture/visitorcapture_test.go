package visitorcapture

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/service"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVisitorProfileTouchInputUsesExplicitLocaleBeforeHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := newVisitorCaptureTestContext()
	c.Request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	input := BuildVisitorProfileTouchInput(c, TouchOptions{Locale: "fr"})

	assert.Equal(t, "fr", input.Locale)
	assert.Equal(t, "request", input.LocaleSource)
}

func TestBuildVisitorProfileTouchInputUsesMarketCountryFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := newVisitorCaptureTestContext()
	c.Request.Header.Set("X-Market-Country", "CA")

	input := BuildVisitorProfileTouchInput(c, TouchOptions{})

	assert.Equal(t, "CA", input.CountryCode)
}

func TestBuildVisitorProfileTouchInputPrefersTrustedEdgeCountry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	input := buildVisitorCaptureInputThroughTrustedEdge(
		t,
		[]string{"10.0.0.0/8"},
		"10.1.2.3:12345",
		map[string]string{
			"CF-IPCountry":     "US",
			"X-Market-Country": "CA",
		},
	)

	assert.Equal(t, "US", input.CountryCode)
}

func TestBuildVisitorProfileTouchInputUsesLocationHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := newVisitorCaptureTestContext()
	c.Request.Header.Set("X-Region", "California")
	c.Request.Header.Set("X-City", "Los Angeles")
	c.Request.Header.Set("X-Timezone", "America/Los_Angeles")

	input := BuildVisitorProfileTouchInput(c, TouchOptions{})

	assert.Equal(t, "California", input.Region)
	assert.Equal(t, "Los Angeles", input.City)
	assert.Equal(t, "America/Los_Angeles", input.Timezone)
}

func newVisitorCaptureTestContext() *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	return c
}

func buildVisitorCaptureInputThroughTrustedEdge(t *testing.T, trustedProxies []string, remoteAddr string, headers map[string]string) service.VisitorProfileTouchInput {
	t.Helper()

	edgeMiddleware, err := middleware.NewTrustedEdgeMetadata(trustedProxies)
	require.NoError(t, err)

	var input service.VisitorProfileTouchInput
	router := gin.New()
	router.Use(edgeMiddleware)
	router.GET("/", func(c *gin.Context) {
		input = BuildVisitorProfileTouchInput(c, TouchOptions{})
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	request.RemoteAddr = remoteAddr
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusNoContent, recorder.Code)
	return input
}
