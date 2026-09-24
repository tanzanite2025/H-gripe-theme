package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
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

// ConvertMoneyStrict converts an exact Money value using a fresh exchange-rate
// record. Multiplication and minor-unit rounding remain inside Money.
func (s *ExchangeRateService) ConvertMoneyStrict(amount domainmoney.Money, quoteCurrency string) (domainmoney.Money, error) {
	base := currency.NormalizeCode(amount.Currency().String())
	quote := currency.NormalizeCode(quoteCurrency)
	if base == "" || quote == "" || !currency.IsCatalogCode(base) || !currency.IsCatalogCode(quote) {
		return domainmoney.Money{}, fmt.Errorf("invalid currency conversion %s to %s", base, quote)
	}
	if base == quote {
		return amount, nil
	}
	if s == nil || s.repo == nil {
		return domainmoney.Money{}, fmt.Errorf("%w: exchange-rate service unavailable", ErrExchangeRateMissing)
	}
	now := time.Now().UTC()
	if rate, err := s.repo.FindFresh(base, quote, now); err == nil && rate != nil {
		rat, parseErr := exchangeRateRatRecord(rate)
		if parseErr != nil {
			return domainmoney.Money{}, parseErr
		}
		return amount.ConvertAtRat(rat, quote)
	}
	if rate, err := s.repo.FindFresh(quote, base, now); err == nil && rate != nil {
		rat, parseErr := exchangeRateRatRecord(rate)
		if parseErr != nil {
			return domainmoney.Money{}, parseErr
		}
		return amount.ConvertAtRat(new(big.Rat).Inv(rat), quote)
	}
	config, err := s.GetConfig()
	if err == nil {
		anchor := currency.NormalizeCode(config.BaseCurrency)
		if anchor != "" && anchor != base && anchor != quote {
			baseRate, baseErr := s.repo.FindFresh(anchor, base, now)
			quoteRate, quoteErr := s.repo.FindFresh(anchor, quote, now)
			if baseErr == nil && quoteErr == nil && baseRate != nil && quoteRate != nil {
				baseRat, baseParseErr := exchangeRateRatRecord(baseRate)
				quoteRat, quoteParseErr := exchangeRateRatRecord(quoteRate)
				if baseParseErr != nil {
					return domainmoney.Money{}, baseParseErr
				}
				if quoteParseErr != nil {
					return domainmoney.Money{}, quoteParseErr
				}
				return amount.ConvertAtRat(new(big.Rat).Quo(quoteRat, baseRat), quote)
			}
		}
	}
	return domainmoney.Money{}, fmt.Errorf("%w for %s to %s", ErrExchangeRateMissing, base, quote)
}

func exchangeRateRatRecord(record *currency.ExchangeRate) (*big.Rat, error) {
	if record == nil {
		return nil, errors.New("exchange rate is required")
	}
	if value := strings.TrimSpace(record.RateDecimal); value != "" && value != "0" {
		rat, ok := new(big.Rat).SetString(value)
		if !ok || rat.Sign() <= 0 {
			return nil, errors.New("invalid exchange rate decimal")
		}
		return rat, nil
	}
	return nil, errors.New("exchange rate decimal is required")
}

// exchangeRateFloatRecord converts the canonical decimal rate only at a
// display/snapshot boundary that still exposes a float64 contract. Core rate
// storage and conversion always use RateDecimal through exchangeRateRatRecord.
func exchangeRateFloatRecord(record *currency.ExchangeRate) (float64, error) {
	rat, err := exchangeRateRatRecord(record)
	if err != nil {
		return 0, err
	}
	value, _ := rat.Float64()
	if value <= 0 || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, errors.New("exchange rate overflows float64")
	}
	return value, nil
}

type exchangeRateAPIResponse struct {
	Result             string                 `json:"result"`
	BaseCode           string                 `json:"base_code"`
	ConversionRates    map[string]json.Number `json:"conversion_rates"`
	TimeLastUpdateUnix int64                  `json:"time_last_update_unix"`
	ErrorType          string                 `json:"error-type"`
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
	heartbeatDone := make(chan struct{})
	heartbeatInterval := leaseTTL / 3
	if heartbeatInterval < time.Second {
		heartbeatInterval = time.Second
	}
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if renewed, renewErr := s.repo.RenewSyncLease(exchangeRateSyncLeaseKey, ownerID, time.Now().UTC(), leaseTTL); renewErr != nil || !renewed {
					return
				}
			case <-heartbeatDone:
				return
			}
		}
	}()
	defer func() {
		close(heartbeatDone)
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
		RateDecimal:   "1",
		Source:        config.Provider,
		FetchedAt:     fetchedAt,
		ExpiresAt:     &expiresAt,
	})
	for _, quote := range config.QuoteCurrencies {
		rateValue := strings.TrimSpace(apiResponse.ConversionRates[quote].String())
		rateRat, ok := new(big.Rat).SetString(rateValue)
		if !ok || rateRat.Sign() <= 0 {
			continue
		}
		rates = append(rates, currency.ExchangeRate{
			BaseCurrency:  config.BaseCurrency,
			QuoteCurrency: quote,
			RateDecimal:   rateValue,
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
