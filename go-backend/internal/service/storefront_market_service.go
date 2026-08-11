package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/currency"
	marketdomain "commerce-platform/internal/domain/market"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"
)

var ErrInvalidStorefrontMarket = errors.New("invalid storefront market")

type StorefrontMarketService struct {
	repo *repository.StorefrontMarketRepository
}

type StorefrontMarketInput struct {
	Code                string   `json:"code"`
	Name                string   `json:"name"`
	Countries           []string `json:"countries"`
	DefaultLocale       string   `json:"default_locale"`
	SupportedLocales    []string `json:"supported_locales"`
	DefaultCurrency     string   `json:"default_currency"`
	DisplayCurrencies   []string `json:"display_currencies"`
	PaymentMethodPolicy string   `json:"payment_method_policy"`
	LogisticsPolicy     string   `json:"logistics_policy"`
	TaxPolicy           string   `json:"tax_policy"`
	Enabled             *bool    `json:"enabled"`
	Priority            int      `json:"priority"`
}

type StorefrontMarketOptionSet struct {
	AvailableLanguageLocales []locales.Language        `json:"available_locales"`
	AvailableCurrencies      []currency.CurrencyOption `json:"available_currencies"`
}

func NewStorefrontMarketService(repo *repository.StorefrontMarketRepository) *StorefrontMarketService {
	return &StorefrontMarketService{repo: repo}
}

func (s *StorefrontMarketService) List(enabledOnly bool) ([]StorefrontMarket, error) {
	if s == nil || s.repo == nil {
		return defaultStorefrontMarkets(), nil
	}
	records, err := s.repo.List(enabledOnly)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return defaultStorefrontMarkets(), nil
	}
	return serviceMarketsFromDomain(records), nil
}

func (s *StorefrontMarketService) ListDomain(enabledOnly bool) ([]marketdomain.StorefrontMarket, error) {
	if s == nil || s.repo == nil {
		return domainMarketsFromService(defaultStorefrontMarkets()), nil
	}
	records, err := s.repo.List(enabledOnly)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return domainMarketsFromService(defaultStorefrontMarkets()), nil
	}
	return records, nil
}

func (s *StorefrontMarketService) Get(id uint) (*StorefrontMarket, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("storefront market service is not configured")
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	converted := serviceMarketFromDomain(*record)
	return &converted, nil
}

func (s *StorefrontMarketService) Create(input StorefrontMarketInput) (*StorefrontMarket, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("storefront market service is not configured")
	}
	record, err := normalizeStorefrontMarketInput(input)
	if err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindByCode(record.Code); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: market code already exists", ErrInvalidStorefrontMarket)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Create(&record); err != nil {
		return nil, err
	}
	converted := serviceMarketFromDomain(record)
	return &converted, nil
}

func (s *StorefrontMarketService) Update(id uint, input StorefrontMarketInput) (*StorefrontMarket, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("storefront market service is not configured")
	}
	record, err := normalizeStorefrontMarketInput(input)
	if err != nil {
		return nil, err
	}
	record.ID = id
	if existing, err := s.repo.FindByCode(record.Code); err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("%w: market code already exists", ErrInvalidStorefrontMarket)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Update(&record); err != nil {
		return nil, err
	}
	converted := serviceMarketFromDomain(record)
	return &converted, nil
}

func (s *StorefrontMarketService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return errors.New("storefront market service is not configured")
	}
	return s.repo.Delete(id)
}

func (s *StorefrontMarketService) Options() StorefrontMarketOptionSet {
	return StorefrontMarketOptionSet{
		AvailableLanguageLocales: locales.SupportedLanguages,
		AvailableCurrencies:      currency.Catalog(),
	}
}

func normalizeStorefrontMarketInput(input StorefrontMarketInput) (marketdomain.StorefrontMarket, error) {
	code := normalizeMarketCode(input.Code)
	if code == "" {
		return marketdomain.StorefrontMarket{}, fmt.Errorf("%w: code is required", ErrInvalidStorefrontMarket)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = code
	}

	supportedLanguageLocales, err := normalizeStorefrontMarketInputSupportedLanguageLocales(input.SupportedLocales)
	if err != nil {
		return marketdomain.StorefrontMarket{}, err
	}
	defaultLanguageLocale, err := requireSupportedLocale(input.DefaultLocale)
	if err != nil {
		return marketdomain.StorefrontMarket{}, fmt.Errorf("%w: default_locale is invalid: %v", ErrInvalidStorefrontMarket, err)
	}
	if !stringSliceContains(supportedLanguageLocales, defaultLanguageLocale) {
		return marketdomain.StorefrontMarket{}, fmt.Errorf("%w: default_locale must be included in supported_locales", ErrInvalidStorefrontMarket)
	}

	displayCurrencies := normalizeStorefrontMarketDisplayCurrencies(input.DisplayCurrencies, nil)
	defaultCurrency := currency.NormalizeCode(input.DefaultCurrency)
	if defaultCurrency == "" || !currency.IsCatalogCode(defaultCurrency) {
		defaultCurrency = displayCurrencies[0]
	}
	if !stringSliceContains(displayCurrencies, defaultCurrency) {
		displayCurrencies = append([]string{defaultCurrency}, displayCurrencies...)
	}

	countries, err := normalizeMarketCountryList(input.Countries)
	if err != nil {
		return marketdomain.StorefrontMarket{}, err
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	priority := input.Priority
	if priority <= 0 {
		priority = 100
	}

	record := marketdomain.StorefrontMarket{
		Code:                code,
		Name:                name,
		DefaultLocale:       defaultLanguageLocale,
		SupportedLocales:    marketdomain.StringList(supportedLanguageLocales),
		DefaultCurrency:     defaultCurrency,
		DisplayCurrencies:   marketdomain.StringList(displayCurrencies),
		PaymentMethodPolicy: strings.TrimSpace(input.PaymentMethodPolicy),
		LogisticsPolicy:     strings.TrimSpace(input.LogisticsPolicy),
		TaxPolicy:           strings.TrimSpace(input.TaxPolicy),
		Enabled:             enabled,
		Priority:            priority,
		Countries:           make([]marketdomain.MarketCountry, 0, len(countries)),
	}
	for _, countryCode := range countries {
		record.Countries = append(record.Countries, marketdomain.MarketCountry{Code: countryCode})
	}
	return record, nil
}

func normalizeStorefrontMarketInputSupportedLanguageLocales(values []string) ([]string, error) {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		locale, err := requireSupportedLocale(value)
		if err != nil {
			return nil, fmt.Errorf("%w: supported locale %q is invalid: %v", ErrInvalidStorefrontMarket, strings.TrimSpace(value), err)
		}
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		result = append(result, locale)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: supported_locales is required", ErrInvalidStorefrontMarket)
	}
	return result, nil
}

func serviceMarketsFromDomain(records []marketdomain.StorefrontMarket) []StorefrontMarket {
	markets := make([]StorefrontMarket, 0, len(records))
	for _, record := range records {
		markets = append(markets, serviceMarketFromDomain(record))
	}
	return markets
}

func serviceMarketFromDomain(record marketdomain.StorefrontMarket) StorefrontMarket {
	countries := make([]string, 0, len(record.Countries))
	for _, country := range record.Countries {
		countries = append(countries, normalizeStorefrontCountry(country.Code))
	}
	return StorefrontMarket{
		ID:                  record.ID,
		Code:                normalizeMarketCode(record.Code),
		Name:                strings.TrimSpace(record.Name),
		Countries:           countries,
		DefaultLocale:       locales.ResolveSupported(record.DefaultLocale),
		SupportedLocales:    normalizeStorefrontMarketSupportedLanguageLocales(record.SupportedLocales.Slice(), record.DefaultLocale),
		DefaultCurrency:     currency.NormalizeCode(record.DefaultCurrency),
		DisplayCurrencies:   normalizeStorefrontMarketDisplayCurrencies(record.DisplayCurrencies.Slice(), nil),
		PaymentMethodPolicy: strings.TrimSpace(record.PaymentMethodPolicy),
		LogisticsPolicy:     strings.TrimSpace(record.LogisticsPolicy),
		TaxPolicy:           strings.TrimSpace(record.TaxPolicy),
		Enabled:             record.Enabled,
		Priority:            record.Priority,
	}
}

func domainMarketsFromService(markets []StorefrontMarket) []marketdomain.StorefrontMarket {
	records := make([]marketdomain.StorefrontMarket, 0, len(markets))
	for _, market := range markets {
		records = append(records, domainMarketFromService(market))
	}
	return records
}

func domainMarketFromService(input StorefrontMarket) marketdomain.StorefrontMarket {
	record := marketdomain.StorefrontMarket{
		Code:                input.Code,
		Name:                input.Code,
		DefaultLocale:       input.DefaultLocale,
		SupportedLocales:    marketdomain.StringList(input.SupportedLocales),
		DefaultCurrency:     input.DefaultCurrency,
		DisplayCurrencies:   marketdomain.StringList(input.DisplayCurrencies),
		PaymentMethodPolicy: input.PaymentMethodPolicy,
		LogisticsPolicy:     input.LogisticsPolicy,
		TaxPolicy:           input.TaxPolicy,
		Enabled:             true,
		Priority:            100,
		Countries:           make([]marketdomain.MarketCountry, 0, len(input.Countries)),
	}
	for _, countryCode := range input.Countries {
		record.Countries = append(record.Countries, marketdomain.MarketCountry{Code: countryCode})
	}
	return record
}

func normalizeMarketCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeMarketCountryList(values []string) ([]string, error) {
	seen := map[string]struct{}{}
	countries := make([]string, 0, len(values))
	for _, value := range values {
		country := normalizeStorefrontCountry(value)
		if country == storefrontUnknownCountry {
			return nil, fmt.Errorf("%w: invalid country code %s", ErrInvalidStorefrontMarket, strings.TrimSpace(value))
		}
		if _, ok := seen[country]; ok {
			continue
		}
		seen[country] = struct{}{}
		countries = append(countries, country)
	}
	return countries, nil
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
