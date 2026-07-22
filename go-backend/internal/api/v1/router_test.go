package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tanzanite/internal/app"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/auth/profile", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected anonymous profile probe to return 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExternalWebhooksBypassCSRFProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newWebhookCSRFRouter(t)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "payment webhook",
			path: "/api/v1/payment/webhook/unsupported",
		},
		{
			name: "shipping webhook",
			path: "/api/v1/shipping/webhook/mock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, tt.path, strings.NewReader(`{}`))
			req.Header.Set("Sec-Fetch-Site", "cross-site")

			router.ServeHTTP(w, req)

			if w.Code == http.StatusForbidden && strings.Contains(w.Body.String(), "CSRF_VALIDATION_FAILED") {
				t.Fatalf("expected webhook to bypass CSRF, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestNonWebhookUnsafeRoutesRemainCSRFProtected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newWebhookCSRFRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/subscriptions", strings.NewReader(`{"email":"test@example.com"}`))
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "CSRF_VALIDATION_FAILED") {
		t.Fatalf("expected regular unsafe route to remain CSRF protected, got %d: %s", w.Code, w.Body.String())
	}
}

func newWebhookCSRFRouter(t *testing.T) *gin.Engine {
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

	require.NoError(t, db.AutoMigrate(&shippingdomain.TrackingProviderConfig{}))

	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	deps := &app.Dependencies{
		Services: app.Services{
			Shipping: service.NewShippingService(repository.NewShippingRepository(db)),
		},
	}
	RegisterRoutes(router, deps, cfg)
	return router
}
