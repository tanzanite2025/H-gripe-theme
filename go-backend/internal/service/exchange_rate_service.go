package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	exchangeRateSyncLeaseKey    = "exchange-rate-api-sync"
	exchangeRateSyncLeaseTTL    = 30 * time.Minute
)

var (
	ErrExchangeRateDisabled       = errors.New("exchange rate API is disabled")
	ErrExchangeRateNotConfigured  = errors.New("exchange rate API is not configured")
	ErrExchangeRateMissing        = errors.New("exchange rate is missing")
	ErrExchangeRateSyncInProgress = errors.New("exchange rate sync is already running")
)

type ExchangeRateService struct {
	repo            *repository.ExchangeRateRepository
	settings        *repository.SettingRepository
	currencyPolicy  *CurrencyPolicyService
	marketService   *StorefrontMarketService
	productService  *ProductService
	shippingService *ShippingService
	client          *http.Client
	syncOwnerID     string
	syncLeaseTTL    time.Duration
	syncMu          sync.Mutex
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
	Config                      ExchangeRateConfig                 `json:"config"`
	Rates                       []currency.ExchangeRate            `json:"rates"`
	FetchedAt                   time.Time                          `json:"fetched_at"`
	ExpiresAt                   *time.Time                         `json:"expires_at,omitempty"`
	DisplayPriceRefresh         *ProductDisplayPriceRefreshResult  `json:"display_price_refresh,omitempty"`
	ShippingDisplayPriceRefresh *ShippingDisplayPriceRefreshResult `json:"shipping_display_price_refresh,omitempty"`
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
		repo:         repo,
		settings:     settings,
		client:       &http.Client{Timeout: 12 * time.Second},
		syncOwnerID:  exchangeRateSyncOwnerID(),
		syncLeaseTTL: exchangeRateSyncLeaseTTL,
	}
}

func (s *ExchangeRateService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
	if s == nil {
		return
	}
	s.currencyPolicy = policy
}

func (s *ExchangeRateService) ConfigureStorefrontMarkets(markets *StorefrontMarketService) {
	if s == nil {
		return
	}
	s.marketService = markets
}

func (s *ExchangeRateService) ConfigureProductService(productService *ProductService) {
	if s == nil {
		return
	}
	s.productService = productService
}

func (s *ExchangeRateService) ConfigureShippingService(shippingService *ShippingService) {
	if s == nil {
		return
	}
	s.shippingService = shippingService
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
	if backendEntryCurrency, ok := exchangeRateBackendEntryCurrencyFromPolicy(s); ok {
		config.BaseCurrency = backendEntryCurrency
	}
	if marketQuoteCurrencies, ok := exchangeRateQuoteCurrenciesFromStorefrontMarkets(s, config.BaseCurrency); ok {
		config.QuoteCurrencies = marketQuoteCurrencies
	}
	config.QuoteCurrencies = removeCurrency(config.QuoteCurrencies, config.BaseCurrency)
	return config, nil
}

func exchangeRateDefaultBase(s *ExchangeRateService) string {
	if primary, ok := exchangeRateBackendEntryCurrencyFromPolicy(s); ok {
		return primary
	}
	return defaultExchangeRateBase
}

func exchangeRateBackendEntryCurrencyFromPolicy(s *ExchangeRateService) (string, bool) {
	if s == nil || s.currencyPolicy == nil {
		return "", false
	}
	primary, err := s.currencyPolicy.BackendEntryCurrency()
	if err != nil {
		return "", false
	}
	primary = currency.NormalizeCode(primary)
	if !currency.IsCatalogCode(primary) {
		return "", false
	}
	return primary, true
}

func exchangeRateQuoteCurrenciesFromStorefrontMarkets(s *ExchangeRateService, baseCurrency string) ([]string, bool) {
	markets := (*StorefrontMarketService)(nil)
	if s != nil {
		markets = s.marketService
	}
	if markets == nil {
		markets = &StorefrontMarketService{}
	}
	values, err := markets.ListStorefrontDisplayCurrencies(true)
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
	if !s.syncMu.TryLock() {
		return nil, ErrExchangeRateSyncInProgress
	}
	defer s.syncMu.Unlock()

	now := time.Now().UTC()
	ownerID := strings.TrimSpace(s.syncOwnerID)
	if ownerID == "" {
		ownerID = exchangeRateSyncOwnerID()
	}
	leaseTTL := s.syncLeaseTTL
	if leaseTTL <= 0 {
		leaseTTL = exchangeRateSyncLeaseTTL
	}
	acquired, err := s.repo.TryAcquireSyncLease(exchangeRateSyncLeaseKey, ownerID, now, leaseTTL)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, ErrExchangeRateSyncInProgress
	}
	defer func() {
		_ = s.repo.ReleaseSyncLease(exchangeRateSyncLeaseKey, ownerID, time.Now().UTC())
	}()

	cacheFetchedAt := time.Now().UTC()
	apiResponse, err := s.fetch(config)
	if err != nil {
		return nil, err
	}
	fetchedAt := cacheFetchedAt
	if apiResponse.TimeLastUpdateUnix > 0 {
		fetchedAt = time.Unix(apiResponse.TimeLastUpdateUnix, 0).UTC()
	}
	expiresAt := cacheFetchedAt.Add(time.Duration(config.RefreshMinutes) * time.Minute)
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
	result := &ExchangeRateSyncResult{
		Config:    config,
		Rates:     rates,
		FetchedAt: fetchedAt,
		ExpiresAt: &expiresAt,
	}
	if s.productService != nil {
		displayPriceRefresh, err := s.productService.RefreshDisplayPriceSnapshots(
			config.BaseCurrency,
			config.QuoteCurrencies,
			rates,
		)
		if err != nil {
			return nil, err
		}
		result.DisplayPriceRefresh = &displayPriceRefresh
	}
	if s.shippingService != nil {
		shippingDisplayPriceRefresh, err := s.shippingService.RefreshDisplayPriceSnapshots(
			config.BaseCurrency,
			config.QuoteCurrencies,
			rates,
		)
		if err != nil {
			return nil, err
		}
		result.ShippingDisplayPriceRefresh = &shippingDisplayPriceRefresh
	}
	return result, nil
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
	rate, err := s.repo.FindFresh(base, quote, time.Now().UTC())
	if err != nil || rate.Rate <= 0 {
		return CurrencyConversion{}, false
	}
	return conversionFromRate(amount, quote, rate.Rate, "direct_rate", rate), true
}

func (s *ExchangeRateService) reverseConversion(amount float64, base string, quote string) (CurrencyConversion, bool) {
	rate, err := s.repo.FindFresh(quote, base, time.Now().UTC())
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
	now := time.Now().UTC()
	baseRate, err := s.repo.FindFresh(anchor, base, now)
	if err != nil || baseRate.Rate <= 0 {
		return CurrencyConversion{}, false
	}
	quoteRate, err := s.repo.FindFresh(anchor, quote, now)
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

var exchangeRateSyncInstanceSequence uint64

func exchangeRateSyncOwnerID() string {
	host, _ := os.Hostname()
	host = strings.TrimSpace(host)
	if host == "" {
		host = "unknown-host"
	}
	instanceID := atomic.AddUint64(&exchangeRateSyncInstanceSequence, 1)
	return fmt.Sprintf("%s:%d:%d", host, os.Getpid(), instanceID)
}
