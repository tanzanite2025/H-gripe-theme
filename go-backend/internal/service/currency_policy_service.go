package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

const currencyPolicyLocale = "en"
const currencyPrimaryKey = "currency_primary_currency"
const currencyDisplayKey = "currency_display_currencies"

var ErrInvalidCurrencyPolicy = errors.New("invalid currency display policy")

type CurrencyPolicyService struct {
	settings *repository.SettingRepository
}

func NewCurrencyPolicyService(settings *repository.SettingRepository) *CurrencyPolicyService {
	return &CurrencyPolicyService{settings: settings}
}

func (s *CurrencyPolicyService) GetPolicy() (*currency.Policy, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("currency display policy service is not configured")
	}

	records, err := s.settings.GetByGroup("currency", currencyPolicyLocale)
	if err != nil {
		return nil, err
	}

	primaryCurrency := currency.DefaultPrimaryCurrency
	displayCurrencies := []string{}
	for _, record := range records {
		switch record.Key {
		case currencyPrimaryKey:
			if code := currency.NormalizeCode(record.Value); currency.IsCatalogCode(code) {
				primaryCurrency = code
			}
		case currencyDisplayKey:
			displayCurrencies = splitCurrencyPolicyValue(record.Value)
		}
	}

	normalized, err := normalizeCurrencyPolicy(currency.Policy{
		PrimaryCurrency:     primaryCurrency,
		DisplayCurrencies:   displayCurrencies,
		AvailableCurrencies: currency.Catalog(),
	})
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func (s *CurrencyPolicyService) UpdatePolicy(input currency.Policy) (*currency.Policy, error) {
	normalized, err := normalizeCurrencyPolicy(input)
	if err != nil {
		return nil, err
	}
	if s == nil || s.settings == nil {
		return nil, errors.New("currency display policy service is not configured")
	}

	settings := []setting.Setting{
		{
			Key:         currencyPrimaryKey,
			Value:       normalized.PrimaryCurrency,
			Type:        "string",
			Locale:      currencyPolicyLocale,
			Group:       "currency",
			IsPublic:    true,
			Description: "后台商品、SKU、运费和商业金额录入使用的主基准币种",
		},
		{
			Key:         currencyDisplayKey,
			Value:       strings.Join(normalized.DisplayCurrencies, ","),
			Type:        "string",
			Locale:      currencyPolicyLocale,
			Group:       "currency",
			IsPublic:    true,
			Description: "后台明确添加的次展示币种，用于缓存汇率和前台价格标签",
		},
	}
	if err := s.settings.BatchSet(settings); err != nil {
		return nil, err
	}
	return &normalized, nil
}

func (s *CurrencyPolicyService) DisplayCurrencies() ([]string, error) {
	policy, err := s.GetPolicy()
	if err != nil {
		return nil, err
	}
	return append([]string(nil), policy.DisplayCurrencies...), nil
}

func (s *CurrencyPolicyService) PrimaryCurrency() (string, error) {
	policy, err := s.GetPolicy()
	if err != nil {
		return "", err
	}
	return policy.PrimaryCurrency, nil
}

func normalizeCurrencyPolicy(input currency.Policy) (currency.Policy, error) {
	primaryCurrency := currency.NormalizeCode(input.PrimaryCurrency)
	if primaryCurrency == "" {
		primaryCurrency = currency.DefaultPrimaryCurrency
	}
	if !currency.IsValidCode(primaryCurrency) || !currency.IsCatalogCode(primaryCurrency) {
		return currency.Policy{}, fmt.Errorf("%w: unsupported primary currency %s", ErrInvalidCurrencyPolicy, primaryCurrency)
	}
	displayCurrencies := currency.NormalizeCodes(input.DisplayCurrencies)
	displayCurrencies = removeCurrencyCode(displayCurrencies, primaryCurrency)

	for _, code := range displayCurrencies {
		if !currency.IsCatalogCode(code) {
			return currency.Policy{}, fmt.Errorf("%w: unsupported display currency %s", ErrInvalidCurrencyPolicy, code)
		}
	}

	return currency.Policy{
		PrimaryCurrency:     primaryCurrency,
		DisplayCurrencies:   displayCurrencies,
		AvailableCurrencies: currency.Catalog(),
	}, nil
}

func removeCurrencyCode(values []string, target string) []string {
	target = currency.NormalizeCode(target)
	if target == "" {
		return values
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if currency.NormalizeCode(value) == target {
			continue
		}
		result = append(result, value)
	}
	return result
}

func splitCurrencyPolicyValue(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case ',', ';', '，', '；':
			return true
		default:
			return strings.TrimSpace(string(r)) == ""
		}
	})
}
