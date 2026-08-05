package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/repository"
)

const currencyPolicyLocale = "en"

const (
	currencyAccountingKey            = "currency_accounting_currency"
	currencyDefaultOrderKey          = "currency_default_order_currency"
	currencyAcceptedKey              = "currency_accepted_currencies"
	legacyCurrencyDefaultCheckoutKey = "currency_default_checkout_currency"
	legacyCurrencyCheckoutKey        = "currency_checkout_currencies"
)

var ErrInvalidCurrencyPolicy = errors.New("invalid currency policy")

type CurrencyPolicyService struct {
	settings *repository.SettingRepository
}

func NewCurrencyPolicyService(settings *repository.SettingRepository) *CurrencyPolicyService {
	return &CurrencyPolicyService{settings: settings}
}

func (s *CurrencyPolicyService) GetPolicy() (*currency.Policy, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("currency policy service is not configured")
	}

	records, err := s.settings.GetByGroup("currency", currencyPolicyLocale)
	if err != nil {
		return nil, err
	}

	policy := currency.Policy{
		AvailableCurrencies: currency.Catalog(),
	}
	values := make(map[string]string, len(records))
	for _, record := range records {
		switch record.Key {
		case currencyAccountingKey, currencyDefaultOrderKey, currencyAcceptedKey, legacyCurrencyDefaultCheckoutKey, legacyCurrencyCheckoutKey:
			values[record.Key] = record.Value
		}
	}
	if accounting, ok := currencySettingValue(values, currencyAccountingKey); ok {
		policy.AccountingCurrency = accounting
	} else {
		return nil, fmt.Errorf("%w: missing required setting %s", ErrInvalidCurrencyPolicy, currencyAccountingKey)
	}
	if defaultOrder, ok := currencySettingValue(values, currencyDefaultOrderKey, legacyCurrencyDefaultCheckoutKey); ok {
		policy.DefaultOrderCurrency = defaultOrder
	} else {
		return nil, fmt.Errorf("%w: missing required setting %s", ErrInvalidCurrencyPolicy, currencyDefaultOrderKey)
	}
	if accepted, ok := currencySettingValue(values, currencyAcceptedKey, legacyCurrencyCheckoutKey); ok {
		policy.AcceptedCurrencies = splitCurrencySetting(accepted)
	} else {
		return nil, fmt.Errorf("%w: missing required setting %s", ErrInvalidCurrencyPolicy, currencyAcceptedKey)
	}

	normalized, err := normalizeCurrencyPolicy(policy)
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
		return nil, errors.New("currency policy service is not configured")
	}

	settings := []setting.Setting{
		{
			Key:         currencyAccountingKey,
			Value:       normalized.AccountingCurrency,
			Type:        "string",
			Locale:      currencyPolicyLocale,
			Group:       "currency",
			IsPublic:    true,
			Description: "Internal accounting/base currency",
		},
		{
			Key:         currencyDefaultOrderKey,
			Value:       normalized.DefaultOrderCurrency,
			Type:        "string",
			Locale:      currencyPolicyLocale,
			Group:       "currency",
			IsPublic:    true,
			Description: "Default currency locked onto new orders",
		},
		{
			Key:         currencyAcceptedKey,
			Value:       strings.Join(normalized.AcceptedCurrencies, ","),
			Type:        "string",
			Locale:      currencyPolicyLocale,
			Group:       "currency",
			IsPublic:    true,
			Description: "Business currencies accepted for order payment collection",
		},
	}
	if err := s.settings.BatchSet(settings); err != nil {
		return nil, err
	}
	return &normalized, nil
}

func (s *CurrencyPolicyService) DefaultOrderCurrency() (string, error) {
	policy, err := s.GetPolicy()
	if err != nil {
		return "", err
	}

	defaultOrder := currency.NormalizeCode(policy.DefaultOrderCurrency)
	if defaultOrder == "" {
		return "", fmt.Errorf("%w: default order currency is required", ErrInvalidCurrencyPolicy)
	}
	if !containsCurrency(policy.AcceptedCurrencies, defaultOrder) {
		return "", fmt.Errorf("%w: default order currency must be included in accepted currencies", ErrInvalidCurrencyPolicy)
	}
	return defaultOrder, nil
}

func (s *CurrencyPolicyService) ValidateAcceptedCurrency(requested string) (string, error) {
	policy, err := s.GetPolicy()
	if err != nil {
		return "", err
	}

	requested = currency.NormalizeCode(requested)
	if requested == "" {
		return "", fmt.Errorf("%w: accepted currency is required", ErrInvalidCurrencyPolicy)
	}
	for _, allowed := range policy.AcceptedCurrencies {
		if requested == allowed {
			return requested, nil
		}
	}
	return "", fmt.Errorf("%w: currency %s is not accepted", ErrInvalidCurrencyPolicy, requested)
}

func normalizeCurrencyPolicy(input currency.Policy) (currency.Policy, error) {
	accounting := currency.NormalizeCode(input.AccountingCurrency)
	if accounting == "" {
		return currency.Policy{}, fmt.Errorf("%w: accounting currency is required", ErrInvalidCurrencyPolicy)
	}
	if !currency.IsValidCode(accounting) {
		return currency.Policy{}, fmt.Errorf("%w: invalid accounting currency", ErrInvalidCurrencyPolicy)
	}
	if !currency.IsCatalogCode(accounting) {
		return currency.Policy{}, fmt.Errorf("%w: unsupported accounting currency", ErrInvalidCurrencyPolicy)
	}

	defaultOrder := currency.NormalizeCode(input.DefaultOrderCurrency)
	if defaultOrder == "" {
		return currency.Policy{}, fmt.Errorf("%w: default order currency is required", ErrInvalidCurrencyPolicy)
	}
	if !currency.IsValidCode(defaultOrder) {
		return currency.Policy{}, fmt.Errorf("%w: invalid default order currency", ErrInvalidCurrencyPolicy)
	}
	if !currency.IsCatalogCode(defaultOrder) {
		return currency.Policy{}, fmt.Errorf("%w: unsupported default order currency", ErrInvalidCurrencyPolicy)
	}

	acceptedCurrencies, err := normalizeCurrencyCodeList(input.AcceptedCurrencies, "accepted currencies")
	if err != nil {
		return currency.Policy{}, err
	}
	if len(acceptedCurrencies) == 0 {
		return currency.Policy{}, fmt.Errorf("%w: accepted currencies are required", ErrInvalidCurrencyPolicy)
	}
	if !containsCurrency(acceptedCurrencies, defaultOrder) {
		return currency.Policy{}, fmt.Errorf("%w: default order currency must be included in accepted currencies", ErrInvalidCurrencyPolicy)
	}

	return currency.Policy{
		AccountingCurrency:   accounting,
		DefaultOrderCurrency: defaultOrder,
		AcceptedCurrencies:   acceptedCurrencies,
		AvailableCurrencies:  currency.Catalog(),
	}, nil
}

func currencySettingValue(values map[string]string, keys ...string) (string, bool) {
	for _, key := range keys {
		value, ok := values[key]
		if ok {
			return value, true
		}
	}
	return "", false
}

func splitCurrencySetting(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '，' || r == '；' || r == ' ' || r == '\n' || r == '\r' || r == '\t'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		code := currency.NormalizeCode(part)
		if code != "" {
			result = append(result, code)
		}
	}
	return result
}

func normalizeCurrencyCodeList(values []string, field string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		code := currency.NormalizeCode(value)
		if code == "" {
			continue
		}
		if !currency.IsValidCode(code) {
			return nil, fmt.Errorf("%w: invalid %s code %s", ErrInvalidCurrencyPolicy, field, code)
		}
		if !currency.IsCatalogCode(code) {
			return nil, fmt.Errorf("%w: unsupported %s code %s", ErrInvalidCurrencyPolicy, field, code)
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	sort.Strings(result)
	return result, nil
}

func containsCurrency(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
