package service

import (
	"strings"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/pkg/locales"
)

const storefrontFallbackLocale = "en"
const storefrontUnknownCountry = "ZZ"

type StorefrontContextService struct {
	currencyPolicy *CurrencyPolicyService
	marketService  *StorefrontMarketService
	markets        []StorefrontMarket
}

type StorefrontContextInput struct {
	Country           string
	CountrySource     string
	RequestedLocale   string
	CookieLocale      string
	AcceptLanguage    string
	RequestedCurrency string
	CookieCurrency    string
}

type StorefrontMarket struct {
	ID                  uint     `json:"id,omitempty"`
	Code                string   `json:"code"`
	Name                string   `json:"name,omitempty"`
	Countries           []string `json:"countries,omitempty"`
	DefaultLocale       string   `json:"default_locale"`
	SupportedLocales    []string `json:"supported_locales"`
	DefaultCurrency     string   `json:"default_currency"`
	DisplayCurrencies   []string `json:"display_currencies"`
	PaymentMethodPolicy string   `json:"payment_method_policy,omitempty"`
	LogisticsPolicy     string   `json:"logistics_policy,omitempty"`
	TaxPolicy           string   `json:"tax_policy,omitempty"`
	Enabled             bool     `json:"enabled,omitempty"`
	Priority            int      `json:"priority,omitempty"`
}

type StorefrontContext struct {
	Country  StorefrontCountryContext  `json:"country"`
	Market   StorefrontMarketContext   `json:"market"`
	Locale   StorefrontLocaleContext   `json:"locale"`
	Currency StorefrontCurrencyContext `json:"currency"`
}

type StorefrontCountryContext struct {
	Code   string `json:"code"`
	Source string `json:"source"`
}

type StorefrontMarketContext struct {
	Code                string   `json:"code"`
	DefaultLocale       string   `json:"default_locale"`
	SupportedLocales    []string `json:"supported_locales"`
	DefaultCurrency     string   `json:"default_currency"`
	DisplayCurrencies   []string `json:"display_currencies"`
	PaymentMethodPolicy string   `json:"payment_method_policy,omitempty"`
}

type StorefrontLocaleContext struct {
	Requested string `json:"requested,omitempty"`
	Resolved  string `json:"resolved"`
	Fallback  string `json:"fallback"`
	Source    string `json:"source"`
}

type StorefrontCurrencyContext struct {
	Requested string `json:"requested,omitempty"`
	Resolved  string `json:"resolved"`
	Base      string `json:"base"`
	Source    string `json:"source"`
}

func NewStorefrontContextService(currencyPolicy *CurrencyPolicyService) *StorefrontContextService {
	return &StorefrontContextService{
		currencyPolicy: currencyPolicy,
		markets:        defaultStorefrontMarkets(),
	}
}

func NewStorefrontContextServiceWithMarkets(currencyPolicy *CurrencyPolicyService, marketService *StorefrontMarketService) *StorefrontContextService {
	return &StorefrontContextService{
		currencyPolicy: currencyPolicy,
		marketService:  marketService,
		markets:        defaultStorefrontMarkets(),
	}
}

func (s *StorefrontContextService) Resolve(input StorefrontContextInput) (*StorefrontContext, error) {
	primaryCurrency, err := s.primaryPricingCurrencyForAdminEnteredAmounts()
	if err != nil {
		return nil, err
	}
	displayCurrencies, err := s.configuredSecondaryDisplayCurrenciesForStorefront(primaryCurrency)
	if err != nil {
		return nil, err
	}
	country := normalizeStorefrontCountry(input.Country)
	countrySource := strings.TrimSpace(input.CountrySource)
	if countrySource == "" {
		countrySource = "fallback"
	}
	market, err := s.resolveStorefrontMarketForDetectedCountry(country, displayCurrencies, primaryCurrency)
	if err != nil {
		return nil, err
	}
	resolvedLanguageLocale, languageLocaleSource, requestedLanguageLocale := resolveStorefrontLanguageLocaleFromRequestCookieHeaderOrMarket(input, market)
	displayCurrency, currencySource, requestedCurrency := resolveStorefrontDisplayCurrencyFromRequestCookieOrMarket(input, market, displayCurrencies, primaryCurrency)

	return &StorefrontContext{
		Country: StorefrontCountryContext{Code: country, Source: countrySource},
		Market: StorefrontMarketContext{
			Code:                market.Code,
			DefaultLocale:       market.DefaultLocale,
			SupportedLocales:    append([]string(nil), market.SupportedLocales...),
			DefaultCurrency:     market.DefaultCurrency,
			DisplayCurrencies:   append([]string(nil), market.DisplayCurrencies...),
			PaymentMethodPolicy: strings.TrimSpace(market.PaymentMethodPolicy),
		},
		Locale: StorefrontLocaleContext{
			Requested: requestedLanguageLocale,
			Resolved:  resolvedLanguageLocale,
			Fallback:  storefrontFallbackLocale,
			Source:    languageLocaleSource,
		},
		Currency: StorefrontCurrencyContext{
			Requested: requestedCurrency,
			Resolved:  displayCurrency,
			Base:      primaryCurrency,
			Source:    currencySource,
		},
	}, nil
}

func (s *StorefrontContextService) primaryPricingCurrencyForAdminEnteredAmounts() (string, error) {
	if s == nil || s.currencyPolicy == nil {
		return currency.DefaultPrimaryCurrency, nil
	}
	value, err := s.currencyPolicy.PrimaryCurrency()
	if err != nil {
		return "", err
	}
	value = currency.NormalizeCode(value)
	if value == "" {
		return currency.DefaultPrimaryCurrency, nil
	}
	return value, nil
}

func (s *StorefrontContextService) configuredSecondaryDisplayCurrenciesForStorefront(primaryCurrency string) ([]string, error) {
	primaryCurrency = currency.NormalizeCode(primaryCurrency)
	if primaryCurrency == "" {
		primaryCurrency = currency.DefaultPrimaryCurrency
	}
	if s == nil || s.currencyPolicy == nil {
		return []string{primaryCurrency}, nil
	}
	values, err := s.currencyPolicy.DisplayCurrencies()
	if err != nil {
		return nil, err
	}
	values = currency.NormalizeCodes(values)
	if len(values) == 0 {
		return []string{primaryCurrency}, nil
	}
	return values, nil
}

func (s *StorefrontContextService) resolveStorefrontMarketForDetectedCountry(country string, displayCurrencies []string, primaryCurrency string) (StorefrontMarket, error) {
	markets := defaultStorefrontMarkets()
	if s != nil && len(s.markets) > 0 {
		markets = s.markets
	}
	if s != nil && s.marketService != nil {
		configured, err := s.marketService.List(true)
		if err != nil {
			return StorefrontMarket{}, err
		}
		if len(configured) > 0 {
			markets = configured
		}
	}
	for _, market := range markets {
		for _, candidate := range market.Countries {
			if normalizeStorefrontCountry(candidate) == country {
				return normalizeStorefrontMarket(market, displayCurrencies), nil
			}
		}
	}
	return normalizeStorefrontMarket(globalStorefrontMarket(displayCurrencies, primaryCurrency), displayCurrencies), nil
}

func resolveStorefrontLanguageLocaleFromRequestCookieHeaderOrMarket(input StorefrontContextInput, market StorefrontMarket) (string, string, string) {
	if resolved := resolveLanguageLocaleAllowedByStorefrontMarket(input.RequestedLocale, market); resolved != "" {
		return resolved, "request", locales.ResolveSupported(input.RequestedLocale)
	}
	if resolved := resolveLanguageLocaleAllowedByStorefrontMarket(input.CookieLocale, market); resolved != "" {
		return resolved, "cookie", locales.ResolveSupported(input.CookieLocale)
	}
	if resolved := resolveLanguageLocaleAllowedByStorefrontMarket(input.AcceptLanguage, market); resolved != "" {
		return resolved, "accept_language", locales.ResolveSupported(input.AcceptLanguage)
	}
	if resolved := resolveLanguageLocaleAllowedByStorefrontMarket(market.DefaultLocale, market); resolved != "" {
		return resolved, "market_default", ""
	}
	return storefrontFallbackLocale, "fallback", ""
}

func resolveLanguageLocaleAllowedByStorefrontMarket(value string, market StorefrontMarket) string {
	resolved := locales.ResolveSupported(value)
	if resolved == "" {
		return ""
	}
	for _, supported := range market.SupportedLocales {
		if locales.ResolveSupported(supported) == resolved {
			return resolved
		}
	}
	return ""
}

func resolveStorefrontDisplayCurrencyFromRequestCookieOrMarket(input StorefrontContextInput, market StorefrontMarket, configured []string, primaryCurrency string) (string, string, string) {
	if resolved := resolveDisplayCurrencyAllowedByStorefrontMarket(input.RequestedCurrency, market); resolved != "" {
		return resolved, "request", currency.NormalizeCode(input.RequestedCurrency)
	}
	if resolved := resolveDisplayCurrencyAllowedByStorefrontMarket(input.CookieCurrency, market); resolved != "" {
		return resolved, "cookie", currency.NormalizeCode(input.CookieCurrency)
	}
	if resolved := resolveDisplayCurrencyAllowedByStorefrontMarket(market.DefaultCurrency, market); resolved != "" {
		return resolved, "market_default", ""
	}
	if len(configured) > 0 {
		return configured[0], "currency_policy", ""
	}
	if code := currency.NormalizeCode(primaryCurrency); code != "" {
		return code, "fallback", ""
	}
	return currency.DefaultPrimaryCurrency, "fallback", ""
}

func resolveDisplayCurrencyAllowedByStorefrontMarket(value string, market StorefrontMarket) string {
	code := currency.NormalizeCode(value)
	if code == "" || !currency.IsCatalogCode(code) {
		return ""
	}
	for _, supported := range market.DisplayCurrencies {
		if currency.NormalizeCode(supported) == code {
			return code
		}
	}
	return ""
}

func normalizeStorefrontCountry(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 2 {
		return storefrontUnknownCountry
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return storefrontUnknownCountry
		}
	}
	return value
}

func normalizeStorefrontMarket(market StorefrontMarket, configured []string) StorefrontMarket {
	market.Code = strings.ToUpper(strings.TrimSpace(market.Code))
	if market.Code == "" {
		market.Code = "GLOBAL"
	}
	market.DefaultLocale = locales.ResolveSupported(market.DefaultLocale)
	if market.DefaultLocale == "" {
		market.DefaultLocale = storefrontFallbackLocale
	}
	market.SupportedLocales = normalizeStorefrontMarketSupportedLanguageLocales(market.SupportedLocales, market.DefaultLocale)
	market.DisplayCurrencies = normalizeStorefrontMarketDisplayCurrencies(market.DisplayCurrencies, configured)
	market.DefaultCurrency = currency.NormalizeCode(market.DefaultCurrency)
	if resolveDisplayCurrencyAllowedByStorefrontMarket(market.DefaultCurrency, market) == "" {
		market.DefaultCurrency = market.DisplayCurrencies[0]
	}
	return market
}

func normalizeStorefrontMarketSupportedLanguageLocales(values []string, defaults ...string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values)+len(defaults)+1)
	for _, value := range append(values, defaults...) {
		resolved := locales.ResolveSupported(value)
		if resolved == "" {
			continue
		}
		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = struct{}{}
		result = append(result, resolved)
	}
	if _, ok := seen[storefrontFallbackLocale]; !ok {
		result = append(result, storefrontFallbackLocale)
	}
	return result
}

func normalizeStorefrontMarketDisplayCurrencies(values []string, configured []string) []string {
	result := make([]string, 0, len(values)+len(configured)+1)
	seen := map[string]struct{}{}
	for _, value := range append(values, configured...) {
		code := currency.NormalizeCode(value)
		if code == "" || !currency.IsCatalogCode(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	if len(result) == 0 {
		return []string{"USD"}
	}
	return result
}

func defaultStorefrontMarkets() []StorefrontMarket {
	return []StorefrontMarket{
		{
			Code:              "US",
			Countries:         []string{"US"},
			DefaultLocale:     "en",
			SupportedLocales:  []string{"en", "es"},
			DefaultCurrency:   "USD",
			DisplayCurrencies: []string{"USD"},
		},
		{
			Code:              "EU",
			Countries:         []string{"AT", "BE", "DE", "ES", "FI", "FR", "IE", "IT", "LU", "NL", "PT"},
			DefaultLocale:     "en",
			SupportedLocales:  []string{"en", "de", "fr", "es", "it", "nl"},
			DefaultCurrency:   "EUR",
			DisplayCurrencies: []string{"EUR", "USD", "GBP"},
		},
		{
			Code:              "UK",
			Countries:         []string{"GB"},
			DefaultLocale:     "en",
			SupportedLocales:  []string{"en"},
			DefaultCurrency:   "GBP",
			DisplayCurrencies: []string{"GBP", "USD", "EUR"},
		},
		{
			Code:              "CA",
			Countries:         []string{"CA"},
			DefaultLocale:     "en",
			SupportedLocales:  []string{"en", "fr"},
			DefaultCurrency:   "CAD",
			DisplayCurrencies: []string{"CAD", "USD", "EUR"},
		},
		{
			Code:              "CN",
			Countries:         []string{"CN"},
			DefaultLocale:     "zh_cn",
			SupportedLocales:  []string{"zh_cn", "en"},
			DefaultCurrency:   "CNY",
			DisplayCurrencies: []string{"CNY", "USD"},
		},
	}
}

func globalStorefrontMarket(displayCurrencies []string, primaryCurrency string) StorefrontMarket {
	primaryCurrency = currency.NormalizeCode(primaryCurrency)
	if primaryCurrency == "" {
		primaryCurrency = currency.DefaultPrimaryCurrency
	}
	displayCurrencies = append([]string{primaryCurrency}, displayCurrencies...)
	return StorefrontMarket{
		Code:              "GLOBAL",
		Countries:         nil,
		DefaultLocale:     storefrontFallbackLocale,
		SupportedLocales:  locales.EnabledLocaleCodes(),
		DefaultCurrency:   primaryCurrency,
		DisplayCurrencies: displayCurrencies,
	}
}
