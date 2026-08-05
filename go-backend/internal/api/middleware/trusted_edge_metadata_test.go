package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTrustedEdgeMetadataAcceptsCountryFromTrustedPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, err := NewTrustedEdgeMetadata([]string{"10.0.0.0/8"})
	require.NoError(t, err)

	router := gin.New()
	router.Use(handler)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"country": TrustedEdgeCountry(c)})
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "10.12.0.7:4567"
	request.Header.Set("CF-IPCountry", "us")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"country":"US"}`, recorder.Body.String())
}

func TestTrustedEdgeMetadataIgnoresCountryFromUntrustedPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, err := NewTrustedEdgeMetadata([]string{"10.0.0.0/8"})
	require.NoError(t, err)

	router := gin.New()
	router.Use(handler)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"country": TrustedEdgeCountry(c)})
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "203.0.113.7:4567"
	request.Header.Set("CF-IPCountry", "US")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"country":""}`, recorder.Body.String())
}

func TestNewTrustedEdgeMetadataRejectsInvalidProxy(t *testing.T) {
	_, err := NewTrustedEdgeMetadata([]string{"not-a-network"})
	require.Error(t, err)
}
