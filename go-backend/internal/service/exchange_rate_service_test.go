package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	marketdomain "commerce-platform/internal/domain/market"
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExchangeRateConvertUsesDirectRate(t *testing.T) {
	service, repo := newExchangeRateTestService(t)
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{
		exchangeRateRecord("USD", "EUR", 0.9),
	}))

	converted := service.Convert(100, "USD", "EUR")

	require.True(t, converted.Converted)
	require.Equal(t, "EUR", converted.Currency)
	require.Equal(t, "direct_rate", converted.Source)
	require.InDelta(t, 90, converted.Amount, 0.0001)
	require.InDelta(t, 0.9, converted.Rate, 0.0001)
}

func TestExchangeRateConvertUsesReverseRate(t *testing.T) {
	service, repo := newExchangeRateTestService(t)
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{
		exchangeRateRecord("USD", "EUR", 0.8),
	}))

	converted := service.Convert(100, "EUR", "USD")

	require.True(t, converted.Converted)
	require.Equal(t, "USD", converted.Currency)
	require.Equal(t, "reverse_rate", converted.Source)
	require.InDelta(t, 125, converted.Amount, 0.0001)
	require.InDelta(t, 1.25, converted.Rate, 0.0001)
}

func TestExchangeRateConvertUsesConfiguredBaseAsCrossRateAnchor(t *testing.T) {
	service, repo := newExchangeRateTestService(t)
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{
		exchangeRateRecord("USD", "EUR", 0.8),
		exchangeRateRecord("USD", "GBP", 0.7),
	}))

	converted := service.Convert(100, "EUR", "GBP")

	require.True(t, converted.Converted)
	require.Equal(t, "GBP", converted.Currency)
	require.Equal(t, "cross_rate", converted.Source)
	require.InDelta(t, 87.5, converted.Amount, 0.0001)
	require.InDelta(t, 0.875, converted.Rate, 0.0001)
}

func TestExchangeRateConvertFallsBackToCatalogCurrencyWhenRateMissing(t *testing.T) {
	service, _ := newExchangeRateTestService(t)

	converted := service.Convert(100, "USD", "EUR")

	require.False(t, converted.Converted)
	require.Equal(t, "USD", converted.Currency)
	require.Equal(t, "catalog_currency", converted.Source)
	require.Equal(t, ErrExchangeRateMissing.Error(), converted.FallbackReason)
	require.InDelta(t, 100, converted.Amount, 0.0001)
}

func TestExchangeRateConvertIgnoresExpiredRate(t *testing.T) {
	service, repo := newExchangeRateTestService(t)
	record := exchangeRateRecord("USD", "EUR", 0.9)
	expiredAt := time.Now().UTC().Add(-time.Minute)
	record.ExpiresAt = &expiredAt
	require.NoError(t, repo.UpsertRates([]currency.ExchangeRate{record}))

	converted := service.Convert(100, "USD", "EUR")

	require.False(t, converted.Converted)
	require.Equal(t, ErrExchangeRateMissing.Error(), converted.FallbackReason)
	require.InDelta(t, 100, converted.Amount, 0.0001)
}

func TestExchangeRateSyncRejectsConcurrentCallsWithinService(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		close(started)
		<-release
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]interface{}{
			"result":           "success",
			"base_code":        "USD",
			"conversion_rates": map[string]float64{"EUR": 0.9},
		})
	}))
	t.Cleanup(server.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&settingdomain.Setting{},
		&currency.ExchangeRate{},
		&currency.ExchangeRateSyncLease{},
	))

	settingRepo := repository.NewSettingRepository(db)
	require.NoError(t, settingRepo.BatchSet([]settingdomain.Setting{
		{Key: "exchange_rate_enabled", Value: "true", Type: "bool", Locale: "en", Group: "api"},
		{Key: "exchange_rate_api_key", Value: "test-key", Type: "string", Locale: "en", Group: "api"},
		{Key: "exchange_rate_endpoint", Value: server.URL + "/latest/{base}", Type: "string", Locale: "en", Group: "api"},
	}))

	service := NewExchangeRateService(repository.NewExchangeRateRepository(db), settingRepo)
	service.client = server.Client()

	firstResult := make(chan error, 1)
	go func() {
		_, syncErr := service.Sync()
		firstResult <- syncErr
	}()

	<-started
	secondResult := make(chan error, 1)
	go func() {
		_, syncErr := service.Sync()
		secondResult <- syncErr
	}()

	select {
	case syncErr := <-secondResult:
		require.ErrorIs(t, syncErr, ErrExchangeRateSyncInProgress)
	case <-time.After(time.Second):
		t.Fatal("concurrent exchange-rate sync did not return promptly")
	}

	close(release)
	require.NoError(t, <-firstResult)
}

func TestExchangeRateConfigUsesBackendEntryCurrencyAndEnabledMarketTargets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &currency.ExchangeRate{}, &marketdomain.StorefrontMarket{}, &marketdomain.MarketCountry{}))

	settingRepo := repository.NewSettingRepository(db)
	policy := NewCurrencyPolicyService(settingRepo)
	_, err = policy.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "CNY",
		DisplayCurrencies: []string{"USD", "EUR"},
	})
	require.NoError(t, err)
	marketService := NewStorefrontMarketService(repository.NewStorefrontMarketRepository(db))
	_, err = marketService.Create(StorefrontMarketInput{
		Code:              "US",
		Name:              "United States",
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "USD",
		DisplayCurrencies: []string{"USD", "EUR", "CNY"},
	})
	require.NoError(t, err)
	_, err = marketService.Create(StorefrontMarketInput{
		Code:              "EU",
		Name:              "Europe",
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "EUR",
		DisplayCurrencies: []string{"EUR", "USD"},
	})
	require.NoError(t, err)
	service := NewExchangeRateService(repository.NewExchangeRateRepository(db), settingRepo)
	service.ConfigureCurrencyPolicy(policy)
	service.ConfigureStorefrontMarkets(marketService)

	config, err := service.GetConfig()

	require.NoError(t, err)
	require.Equal(t, "CNY", config.BaseCurrency)
	require.ElementsMatch(t, []string{"USD", "EUR"}, config.QuoteCurrencies)
	require.NotContains(t, config.QuoteCurrencies, "CNY")
}

func TestExchangeRateConfigUsesDefaultStorefrontMarketTargetsWhenNoMarketsConfigured(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &currency.ExchangeRate{}))

	settingRepo := repository.NewSettingRepository(db)
	policy := NewCurrencyPolicyService(settingRepo)
	service := NewExchangeRateService(repository.NewExchangeRateRepository(db), settingRepo)
	service.ConfigureCurrencyPolicy(policy)

	config, err := service.GetConfig()

	require.NoError(t, err)
	require.Equal(t, currency.DefaultPrimaryCurrency, config.BaseCurrency)
	require.ElementsMatch(t, []string{"EUR", "GBP", "CAD", "CNY"}, config.QuoteCurrencies)
	require.NotContains(t, config.QuoteCurrencies, currency.DefaultPrimaryCurrency)
}

func newExchangeRateTestService(t *testing.T) (*ExchangeRateService, *repository.ExchangeRateRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &currency.ExchangeRate{}))
	repo := repository.NewExchangeRateRepository(db)
	return NewExchangeRateService(repo, repository.NewSettingRepository(db)), repo
}

func exchangeRateRecord(base string, quote string, rate float64) currency.ExchangeRate {
	fetchedAt := time.Now().UTC().Add(-time.Minute)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	return currency.ExchangeRate{
		BaseCurrency:  base,
		QuoteCurrency: quote,
		Rate:          rate,
		Source:        "test-rate",
		FetchedAt:     fetchedAt,
		ExpiresAt:     &expiresAt,
	}
}
