package service

import (
	"testing"

	"tanzanite/internal/domain/currency"

	"github.com/stretchr/testify/require"
)

func TestStorefrontContextResolvesExplicitLocaleAndCurrency(t *testing.T) {
	service := NewStorefrontContextService(nil)

	context, err := service.Resolve(StorefrontContextInput{
		Country:           "DE",
		CountrySource:     "cf_ip_country",
		RequestedLocale:   "fr-FR",
		AcceptLanguage:    "de-DE,de;q=0.9,en;q=0.8",
		RequestedCurrency: "EUR",
	})

	require.NoError(t, err)
	require.Equal(t, "DE", context.Country.Code)
	require.Equal(t, "EU", context.Market.Code)
	require.Equal(t, "fr", context.Locale.Resolved)
	require.Equal(t, "request", context.Locale.Source)
	require.Equal(t, "EUR", context.Currency.Resolved)
	require.Equal(t, "request", context.Currency.Source)
}

func TestStorefrontContextFallsBackToMarketLocale(t *testing.T) {
	service := NewStorefrontContextService(nil)

	context, err := service.Resolve(StorefrontContextInput{
		Country:        "CN",
		CountrySource:  "cf_ip_country",
		AcceptLanguage: "fr-FR,fr;q=0.9",
	})

	require.NoError(t, err)
	require.Equal(t, "CN", context.Market.Code)
	require.Equal(t, "zh_cn", context.Locale.Resolved)
	require.Equal(t, "market_default", context.Locale.Source)
	require.Equal(t, "CNY", context.Currency.Resolved)
}

func TestStorefrontContextUsesGlobalFallbackForUnknownCountry(t *testing.T) {
	service := NewStorefrontContextService(nil)

	context, err := service.Resolve(StorefrontContextInput{
		Country:        "",
		CountrySource:  "fallback",
		AcceptLanguage: "it-IT,it;q=0.9",
	})

	require.NoError(t, err)
	require.Equal(t, "ZZ", context.Country.Code)
	require.Equal(t, "GLOBAL", context.Market.Code)
	require.Equal(t, "it", context.Locale.Resolved)
	require.Equal(t, "accept_language", context.Locale.Source)
	require.Equal(t, "USD", context.Currency.Resolved)
}

func TestStorefrontContextGlobalFallbackKeepsPrimaryCurrencySelectable(t *testing.T) {
	policy := newTestCurrencyPolicyService(t)
	_, err := policy.UpdatePolicy(currency.Policy{
		PrimaryCurrency:   "CNY",
		DisplayCurrencies: []string{"USD", "EUR"},
	})
	require.NoError(t, err)
	service := NewStorefrontContextService(policy)

	context, err := service.Resolve(StorefrontContextInput{
		Country:       "ZZ",
		CountrySource: "fallback",
	})

	require.NoError(t, err)
	require.Equal(t, "GLOBAL", context.Market.Code)
	require.Equal(t, "CNY", context.Market.DefaultCurrency)
	require.Equal(t, []string{"CNY", "USD", "EUR"}, context.Market.DisplayCurrencies)
	require.Equal(t, "CNY", context.Currency.Base)
	require.Equal(t, "CNY", context.Currency.Resolved)
}
