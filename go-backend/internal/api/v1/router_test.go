package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tanzanite/internal/app"
	"tanzanite/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesBuildsCompleteRouteTree(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	deps := &app.Dependencies{}

	RegisterRoutes(router, deps, cfg)
}

func TestAnonymousProfileProbeReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	RegisterRoutes(router, &app.Dependencies{}, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/profile", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected anonymous profile probe to return 204, got %d: %s", w.Code, w.Body.String())
	}
}
