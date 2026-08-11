package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/currency"
	settingdomain "commerce-platform/internal/domain/setting"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

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

func TestCurrencyPolicyRouteUsesDomainHandler(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))
	seedCurrencyPolicySetting := func(key, value string) {
		require.NoError(t, db.Create(&settingdomain.Setting{
			Key:      key,
			Value:    value,
			Type:     "string",
			Locale:   "en",
			Group:    "currency",
			IsPublic: true,
		}).Error)
	}
	seedCurrencyPolicySetting("currency_primary_currency", "CNY")
	seedCurrencyPolicySetting("currency_display_currencies", "USD,EUR")

	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	settingRepo := repository.NewSettingRepository(db)
	deps := &app.Dependencies{
		Services: app.Services{
			CurrencyPolicy: service.NewCurrencyPolicyService(settingRepo),
		},
	}
	RegisterRoutes(router, deps, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/settings/currency-policy", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"primary_currency":"CNY"`)
	require.Contains(t, w.Body.String(), `"display_currencies":["USD","EUR"]`)
	require.Contains(t, w.Body.String(), `"available_currencies"`)
	require.NotContains(t, w.Body.String(), `"accounting_currency"`)
	require.NotContains(t, w.Body.String(), `"default_order_currency"`)
	require.NotContains(t, w.Body.String(), `"accepted_currencies"`)
	require.NotContains(t, w.Body.String(), "Setting not found")
}

func TestCurrencyPolicyRouteLeavesDisplayCurrenciesEmptyWhenUnset(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))
	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	settingRepo := repository.NewSettingRepository(db)
	deps := &app.Dependencies{
		Services: app.Services{
			CurrencyPolicy: service.NewCurrencyPolicyService(settingRepo),
		},
	}
	RegisterRoutes(router, deps, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/settings/currency-policy", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"primary_currency":"USD"`)
	require.Contains(t, w.Body.String(), `"display_currencies":[]`)
	require.NotContains(t, w.Body.String(), `"accounting_currency"`)
	require.NotContains(t, w.Body.String(), `"default_order_currency"`)
	require.NotContains(t, w.Body.String(), `"accepted_currencies"`)
	require.NotContains(t, w.Body.String(), `"default_checkout_currency"`)
	require.NotContains(t, w.Body.String(), `"checkout_currencies"`)
}

func TestStorefrontContextRouteResolvesMarketLocaleAndCurrency(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))
	settingRepo := repository.NewSettingRepository(db)
	currencyPolicy := service.NewCurrencyPolicyService(settingRepo)
	_, err = currencyPolicy.UpdatePolicy(currency.Policy{DisplayCurrencies: []string{"USD", "EUR", "GBP"}})
	require.NoError(t, err)

	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	deps := &app.Dependencies{
		Services: app.Services{
			CurrencyPolicy:    currencyPolicy,
			StorefrontContext: service.NewStorefrontContextService(currencyPolicy),
		},
	}
	RegisterRoutes(router, deps, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/storefront/context?locale=de&currency=EUR", nil)
	req.Header.Set("CF-IPCountry", "DE")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"code":"DE"`)
	require.Contains(t, w.Body.String(), `"code":"EU"`)
	require.Contains(t, w.Body.String(), `"resolved":"de"`)
	require.Contains(t, w.Body.String(), `"resolved":"EUR"`)
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
