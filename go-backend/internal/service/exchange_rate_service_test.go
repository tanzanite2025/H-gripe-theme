package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
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

func TestExchangeRateConfigDefaultsBaseToPrimaryPricingCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &currency.ExchangeRate{}))

	settingRepo := repository.NewSettingRepository(db)
	policy := NewCurrencyPolicyService(settingRepo)
	_, err = policy.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "CNY",
		DisplayCurrencies: []string{"USD", "EUR"},
	})
	require.NoError(t, err)
	service := NewExchangeRateService(repository.NewExchangeRateRepository(db), settingRepo)
	service.ConfigureCurrencyPolicy(policy)

	config, err := service.GetConfig()

	require.NoError(t, err)
	require.Equal(t, "CNY", config.BaseCurrency)
	require.Equal(t, []string{"USD", "EUR"}, config.QuoteCurrencies)
	require.NotContains(t, config.QuoteCurrencies, "CNY")
}

func TestExchangeRateConfigLeavesQuoteCurrenciesEmptyWhenPricingPolicyHasNoSecondaryDisplays(t *testing.T) {
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
	require.Empty(t, config.QuoteCurrencies)
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
	fetchedAt := time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
	expiresAt := fetchedAt.Add(24 * time.Hour)
	return currency.ExchangeRate{
		BaseCurrency:  base,
		QuoteCurrency: quote,
		Rate:          rate,
		Source:        "test-rate",
		FetchedAt:     fetchedAt,
		ExpiresAt:     &expiresAt,
	}
}
