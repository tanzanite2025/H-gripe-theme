package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/repository"
)

const (
	exchangeRateLocale          = "en"
	exchangeRateGroup           = "api"
	defaultExchangeRateProvider = "ExchangeRate-API"
	defaultExchangeRateEndpoint = "https://v6.exchangerate-api.com/v6/{apiKey}/latest/{base}"
	defaultExchangeRateBase     = "USD"
	defaultExchangeRateRefresh  = 1440
)

var (
	ErrExchangeRateDisabled      = errors.New("exchange rate API is disabled")
	ErrExchangeRateNotConfigured = errors.New("exchange rate API is not configured")
	ErrExchangeRateMissing       = errors.New("exchange rate is missing")
)

type ExchangeRateService struct {
	repo           *repository.ExchangeRateRepository
	settings       *repository.SettingRepository
	currencyPolicy *CurrencyPolicyService
	client         *http.Client
}

type ExchangeRateConfig struct {
	Enabled         bool     `json:"enabled"`
	Provider        string   `json:"provider"`
	Endpoint        string   `json:"endpoint"`
	BaseCurrency    string   `json:"base_currency"`
	QuoteCurrencies []string `json:"quote_currencies"`
	RefreshMinutes  int      `json:"refresh_minutes"`
	APIKeySet       bool     `json:"api_key_set"`
	apiKey          string
}

type ExchangeRateSyncResult struct {
	Config    ExchangeRateConfig      `json:"config"`
	Rates     []currency.ExchangeRate `json:"rates"`
	FetchedAt time.Time               `json:"fetched_at"`
	ExpiresAt *time.Time              `json:"expires_at,omitempty"`
}

type CurrencyConversion struct {
	Amount         float64    `json:"amount"`
	Currency       string     `json:"currency"`
	Rate           float64    `json:"rate"`
	Source         string     `json:"source"`
	Converted      bool       `json:"converted"`
	FetchedAt      *time.Time `json:"fetched_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	FallbackReason string     `json:"fallback_reason,omitempty"`
}

type exchangeRateAPIResponse struct {
	Result             string             `json:"result"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	ErrorType          string             `json:"error-type"`
}

func NewExchangeRateService(repo *repository.ExchangeRateRepository, settings *repository.SettingRepository) *ExchangeRateService {
	return &ExchangeRateService{
		repo:     repo,
		settings: settings,
		client:   &http.Client{Timeout: 12 * time.Second},
	}
}

func (s *ExchangeRateService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
	if s == nil {
		return
	}
	s.currencyPolicy = policy
}

func (s *ExchangeRateService) GetConfig() (ExchangeRateConfig, error) {
	config := ExchangeRateConfig{
		Provider:       defaultExchangeRateProvider,
		Endpoint:       defaultExchangeRateEndpoint,
		BaseCurrency:   exchangeRateDefaultBase(s),
		RefreshMinutes: defaultExchangeRateRefresh,
	}
	if s == nil || s.settings == nil {
		return config, nil
	}
	records, err := s.settings.GetByGroup(exchangeRateGroup, exchangeRateLocale)
	if err != nil {
		return config, err
	}
	for _, record := range records {
		switch record.Key {
		case "exchange_rate_enabled":
			config.Enabled = parseSettingBool(record.Value)
		case "exchange_rate_provider":
			if value := strings.TrimSpace(record.Value); value != "" {
				config.Provider = value
			}
		case "exchange_rate_endpoint":
			if value := strings.TrimSpace(record.Value); value != "" {
				config.Endpoint = value
			}
		case "exchange_rate_refresh_minutes":
			if value, err := strconv.Atoi(strings.TrimSpace(record.Value)); err == nil && value > 0 {
				config.RefreshMinutes = value
			}
		case "exchange_rate_api_key":
			config.apiKey = strings.TrimSpace(record.Value)
			config.APIKeySet = config.apiKey != ""
		}
	}
	if policyBaseCurrency, ok := exchangeRatePrimaryCurrencyFromPricingPolicy(s); ok {
		config.BaseCurrency = policyBaseCurrency
	}
	if policyQuoteCurrencies, ok := exchangeRateQuoteCurrenciesFromPricingPolicy(s, config.BaseCurrency); ok {
		config.QuoteCurrencies = policyQuoteCurrencies
	}
	config.QuoteCurrencies = removeCurrency(config.QuoteCurrencies, config.BaseCurrency)
	return config, nil
}

func exchangeRateDefaultBase(s *ExchangeRateService) string {
	if primary, ok := exchangeRatePrimaryCurrencyFromPricingPolicy(s); ok {
		return primary
	}
	return defaultExchangeRateBase
}

func exchangeRatePrimaryCurrencyFromPricingPolicy(s *ExchangeRateService) (string, bool) {
	if s == nil || s.currencyPolicy == nil {
		return "", false
	}
	primary, err := s.currencyPolicy.PrimaryCurrency()
	if err != nil {
		return "", false
	}
	primary = currency.NormalizeCode(primary)
	if !currency.IsCatalogCode(primary) {
		return "", false
	}
	return primary, true
}

func exchangeRateQuoteCurrenciesFromPricingPolicy(s *ExchangeRateService, baseCurrency string) ([]string, bool) {
	if s == nil || s.currencyPolicy == nil {
		return nil, false
	}
	values, err := s.currencyPolicy.DisplayCurrencies()
	if err != nil {
		return nil, false
	}
	return removeCurrency(values, baseCurrency), true
}

func (s *ExchangeRateService) Sync() (*ExchangeRateSyncResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("exchange rate service is not configured")
	}
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	if !config.Enabled {
		return nil, ErrExchangeRateDisabled
	}
	if !config.APIKeySet {
		return nil, ErrExchangeRateNotConfigured
	}

	apiResponse, err := s.fetch(config)
	if err != nil {
		return nil, err
	}
	fetchedAt := time.Now().UTC()
	if apiResponse.TimeLastUpdateUnix > 0 {
		fetchedAt = time.Unix(apiResponse.TimeLastUpdateUnix, 0).UTC()
	}
	expiresAt := fetchedAt.Add(time.Duration(config.RefreshMinutes) * time.Minute)
	rates := make([]currency.ExchangeRate, 0, len(config.QuoteCurrencies)+1)
	rates = append(rates, currency.ExchangeRate{
		BaseCurrency:  config.BaseCurrency,
		QuoteCurrency: config.BaseCurrency,
		Rate:          1,
		Source:        config.Provider,
		FetchedAt:     fetchedAt,
		ExpiresAt:     &expiresAt,
	})
	for _, quote := range config.QuoteCurrencies {
		rate := apiResponse.ConversionRates[quote]
		if rate <= 0 {
			continue
		}
		rates = append(rates, currency.ExchangeRate{
			BaseCurrency:  config.BaseCurrency,
			QuoteCurrency: quote,
			Rate:          rate,
			Source:        config.Provider,
			FetchedAt:     fetchedAt,
			ExpiresAt:     &expiresAt,
		})
	}
	if err := s.repo.UpsertRates(rates); err != nil {
		return nil, err
	}
	return &ExchangeRateSyncResult{Config: config, Rates: rates, FetchedAt: fetchedAt, ExpiresAt: &expiresAt}, nil
}

func (s *ExchangeRateService) List(baseCurrency string) ([]currency.ExchangeRate, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("exchange rate service is not configured")
	}
	return s.repo.List(currency.NormalizeCode(baseCurrency))
}

func (s *ExchangeRateService) Convert(amount float64, baseCurrency string, quoteCurrency string) CurrencyConversion {
	base := currency.NormalizeCode(baseCurrency)
	quote := currency.NormalizeCode(quoteCurrency)
	if amount <= 0 || base == "" || quote == "" || !currency.IsCatalogCode(base) || !currency.IsCatalogCode(quote) {
		return CurrencyConversion{Amount: amount, Currency: base, Rate: 1, Source: "catalog_currency", FallbackReason: "invalid_input"}
	}
	if base == quote {
		return CurrencyConversion{Amount: amount, Currency: quote, Rate: 1, Source: "same_currency", Converted: true}
	}
	if s == nil || s.repo == nil {
		return CurrencyConversion{Amount: amount, Currency: base, Rate: 1, Source: "catalog_currency", FallbackReason: "service_unavailable"}
	}
	if converted, ok := s.directConversion(amount, base, quote); ok {
		return converted
	}
	if converted, ok := s.reverseConversion(amount, base, quote); ok {
		return converted
	}
	if converted, ok := s.crossConversion(amount, base, quote); ok {
		return converted
	}
	return CurrencyConversion{Amount: amount, Currency: base, Rate: 1, Source: "catalog_currency", FallbackReason: ErrExchangeRateMissing.Error()}
}

func (s *ExchangeRateService) directConversion(amount float64, base string, quote string) (CurrencyConversion, bool) {
	rate, err := s.repo.Find(base, quote)
	if err != nil || rate.Rate <= 0 {
		return CurrencyConversion{}, false
	}
	return conversionFromRate(amount, quote, rate.Rate, "direct_rate", rate), true
}

func (s *ExchangeRateService) reverseConversion(amount float64, base string, quote string) (CurrencyConversion, bool) {
	rate, err := s.repo.Find(quote, base)
	if err != nil || rate.Rate <= 0 {
		return CurrencyConversion{}, false
	}
	return conversionFromRate(amount, quote, 1/rate.Rate, "reverse_rate", rate), true
}

func (s *ExchangeRateService) crossConversion(amount float64, base string, quote string) (CurrencyConversion, bool) {
	config, err := s.GetConfig()
	if err != nil {
		return CurrencyConversion{}, false
	}
	anchor := currency.NormalizeCode(config.BaseCurrency)
	if anchor == "" || anchor == base || anchor == quote {
		return CurrencyConversion{}, false
	}
	baseRate, err := s.repo.Find(anchor, base)
	if err != nil || baseRate.Rate <= 0 {
		return CurrencyConversion{}, false
	}
	quoteRate, err := s.repo.Find(anchor, quote)
	if err != nil || quoteRate.Rate <= 0 {
		return CurrencyConversion{}, false
	}
	rate := quoteRate.Rate / baseRate.Rate
	return conversionFromRate(amount, quote, rate, "cross_rate", quoteRate), true
}

func (s *ExchangeRateService) fetch(config ExchangeRateConfig) (*exchangeRateAPIResponse, error) {
	endpoint := strings.ReplaceAll(config.Endpoint, "{apiKey}", config.apiKey)
	endpoint = strings.ReplaceAll(endpoint, "{base}", config.BaseCurrency)
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	client := s.client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("exchange rate API returned status %d", response.StatusCode)
	}
	var payload exchangeRateAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if strings.ToLower(payload.Result) != "success" {
		return nil, fmt.Errorf("exchange rate API error: %s", payload.ErrorType)
	}
	if currency.NormalizeCode(payload.BaseCode) != config.BaseCurrency {
		return nil, fmt.Errorf("exchange rate API returned unexpected base %s", payload.BaseCode)
	}
	return &payload, nil
}

func conversionFromRate(amount float64, currencyCode string, rate float64, source string, record *currency.ExchangeRate) CurrencyConversion {
	converted := CurrencyConversion{
		Amount:    amount * rate,
		Currency:  currencyCode,
		Rate:      rate,
		Source:    source,
		Converted: true,
	}
	if record != nil {
		converted.FetchedAt = &record.FetchedAt
		converted.ExpiresAt = record.ExpiresAt
	}
	return converted
}

func parseSettingBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on", "enabled":
		return true
	default:
		return false
	}
}

func removeCurrency(values []string, target string) []string {
	target = currency.NormalizeCode(target)
	result := make([]string, 0, len(values))
	for _, value := range values {
		code := currency.NormalizeCode(value)
		if code == "" || code == target || !currency.IsCatalogCode(code) {
			continue
		}
		result = append(result, code)
	}
	return result
}
