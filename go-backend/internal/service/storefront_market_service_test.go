package service

import (
	"errors"
	"testing"

	marketdomain "commerce-platform/internal/domain/market"
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStorefrontContextUsesConfiguredMarket(t *testing.T) {
	db := newStorefrontMarketTestDB(t)
	marketService := NewStorefrontMarketService(repository.NewStorefrontMarketRepository(db))

	enabled := true
	created, err := marketService.Create(StorefrontMarketInput{
		Code:              "AU",
		Name:              "Australia",
		Countries:         []string{"AU"},
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "AUD",
		DisplayCurrencies: []string{"AUD", "USD"},
		Enabled:           &enabled,
		Priority:          10,
	})
	require.NoError(t, err)
	require.Equal(t, "AU", created.Code)

	currencyPolicy := NewCurrencyPolicyService(repository.NewSettingRepository(db))
	contextService := NewStorefrontContextServiceWithMarkets(currencyPolicy, marketService)
	context, err := contextService.Resolve(StorefrontContextInput{Country: "AU", CountrySource: "cf_ip_country"})

	require.NoError(t, err)
	require.Equal(t, "AU", context.Country.Code)
	require.Equal(t, "AU", context.Market.Code)
	require.Equal(t, "AUD", context.Currency.Resolved)
	require.Equal(t, "market_default", context.Currency.Source)
}

func TestStorefrontMarketRejectsDuplicateCountry(t *testing.T) {
	db := newStorefrontMarketTestDB(t)
	marketService := NewStorefrontMarketService(repository.NewStorefrontMarketRepository(db))
	enabled := true

	_, err := marketService.Create(StorefrontMarketInput{
		Code:              "ONE",
		Countries:         []string{"NZ"},
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "NZD",
		DisplayCurrencies: []string{"NZD"},
		Enabled:           &enabled,
	})
	require.NoError(t, err)

	_, err = marketService.Create(StorefrontMarketInput{
		Code:              "TWO",
		Countries:         []string{"NZ"},
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "USD",
		DisplayCurrencies: []string{"USD"},
		Enabled:           &enabled,
	})
	require.Error(t, err)
}

func TestStorefrontMarketRejectsUnsupportedLocales(t *testing.T) {
	db := newStorefrontMarketTestDB(t)
	marketService := NewStorefrontMarketService(repository.NewStorefrontMarketRepository(db))

	_, err := marketService.Create(StorefrontMarketInput{
		Code:              "BAD_DEFAULT",
		Countries:         []string{"BR"},
		DefaultLocale:     "zz",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "USD",
		DisplayCurrencies: []string{"USD"},
	})
	require.True(t, errors.Is(err, ErrInvalidStorefrontMarket))

	_, err = marketService.Create(StorefrontMarketInput{
		Code:              "BAD_SUPPORTED",
		Countries:         []string{"CL"},
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en", "zz"},
		DefaultCurrency:   "USD",
		DisplayCurrencies: []string{"USD"},
	})
	require.True(t, errors.Is(err, ErrInvalidStorefrontMarket))
}

func TestStorefrontMarketRequiresDefaultLocaleInSupportedLocales(t *testing.T) {
	db := newStorefrontMarketTestDB(t)
	marketService := NewStorefrontMarketService(repository.NewStorefrontMarketRepository(db))

	_, err := marketService.Create(StorefrontMarketInput{
		Code:              "DEFAULT_MISSING",
		Countries:         []string{"AR"},
		DefaultLocale:     "fr",
		SupportedLocales:  []string{"en"},
		DefaultCurrency:   "USD",
		DisplayCurrencies: []string{"USD"},
	})
	require.True(t, errors.Is(err, ErrInvalidStorefrontMarket))
}

func TestStorefrontMarketNormalizesLocaleAliases(t *testing.T) {
	db := newStorefrontMarketTestDB(t)
	marketService := NewStorefrontMarketService(repository.NewStorefrontMarketRepository(db))

	created, err := marketService.Create(StorefrontMarketInput{
		Code:              "ALIAS",
		Countries:         []string{"PE"},
		DefaultLocale:     "en-US",
		SupportedLocales:  []string{"en-US", "fr-FR"},
		DefaultCurrency:   "USD",
		DisplayCurrencies: []string{"USD"},
	})
	require.NoError(t, err)
	require.Equal(t, "en", created.DefaultLocale)
	require.Equal(t, []string{"en", "fr"}, created.SupportedLocales)
}

func newStorefrontMarketTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &marketdomain.StorefrontMarket{}, &marketdomain.MarketCountry{}))
	return db
}
