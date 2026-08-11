package currency

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestListExchangeRatesReturnsCachedBackendRates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &currencydomain.ExchangeRate{}))

	fetchedAt := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&currencydomain.ExchangeRate{
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Rate:          0.91,
		Source:        "test-provider",
		FetchedAt:     fetchedAt,
	}).Error)

	handler := NewHandler(nil, service.NewExchangeRateService(
		repository.NewExchangeRateRepository(db),
		repository.NewSettingRepository(db),
	))
	router := gin.New()
	router.GET("/exchange-rates", handler.ListExchangeRates)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/exchange-rates?base=usd", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"base_currency":"USD"`)
	require.Contains(t, response.Body.String(), `"quote_currency":"EUR"`)
	require.Contains(t, response.Body.String(), `"rate":0.91`)
	require.Contains(t, response.Body.String(), `"provider":"ExchangeRate-API"`)
}

func TestListExchangeRatesDefaultsToPrimaryPricingCurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &currencydomain.ExchangeRate{}))

	settingRepo := repository.NewSettingRepository(db)
	policyService := service.NewCurrencyPolicyService(settingRepo)
	_, err = policyService.UpdatePolicy(currencydomain.Policy{
		PrimaryCurrency:   "CNY",
		DisplayCurrencies: []string{"USD"},
	})
	require.NoError(t, err)
	require.NoError(t, db.Create(&currencydomain.ExchangeRate{
		BaseCurrency:  "CNY",
		QuoteCurrency: "USD",
		Rate:          0.14,
		Source:        "test-provider",
		FetchedAt:     time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
	}).Error)

	exchangeRateService := service.NewExchangeRateService(
		repository.NewExchangeRateRepository(db),
		settingRepo,
	)
	exchangeRateService.ConfigureCurrencyPolicy(policyService)
	handler := NewHandler(policyService, exchangeRateService)
	router := gin.New()
	router.GET("/exchange-rates", handler.ListExchangeRates)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/exchange-rates", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"base_currency":"CNY"`)
	require.Contains(t, response.Body.String(), `"quote_currency":"USD"`)
}
