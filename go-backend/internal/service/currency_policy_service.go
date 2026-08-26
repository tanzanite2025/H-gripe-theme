package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

const currencyPolicyLocale = "en"
const currencyPrimaryKey = "currency_primary_currency"

var ErrInvalidCurrencyPolicy = errors.New("invalid backend entry currency policy")

type CurrencyPolicyService struct {
	settings *repository.SettingRepository
}

func NewCurrencyPolicyService(settings *repository.SettingRepository) *CurrencyPolicyService {
	return &CurrencyPolicyService{settings: settings}
}

func (s *CurrencyPolicyService) GetPolicy() (*currency.Policy, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("backend entry currency policy service is not configured")
	}

	records, err := s.settings.GetByGroup("currency", currencyPolicyLocale)
	if err != nil {
		return nil, err
	}

	primaryCurrency := currency.DefaultPrimaryCurrency
	for _, record := range records {
		switch record.Key {
		case currencyPrimaryKey:
			if code := currency.NormalizeCode(record.Value); currency.IsCatalogCode(code) {
				primaryCurrency = code
			}
		}
	}

	normalized, err := normalizeCurrencyPolicy(currency.Policy{
		PrimaryCurrency:     primaryCurrency,
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
		return nil, errors.New("backend entry currency policy service is not configured")
	}

	settings := []setting.Setting{
		{
			Key:         currencyPrimaryKey,
			Value:       normalized.PrimaryCurrency,
			Type:        "string",
			Locale:      currencyPolicyLocale,
			Group:       "currency",
			IsPublic:    true,
			Description: "后台商品、SKU、运费和商业金额录入使用的唯一币种",
		},
	}
	if err := s.settings.BatchSet(settings); err != nil {
		return nil, err
	}
	// currency_display_currencies used to be stored globally. Storefront display
	// currencies now belong to enabled market settings, so remove the old global
	// value when this policy is saved to avoid stale configuration drift.
	if err := s.settings.Delete("currency_display_currencies", currencyPolicyLocale); err != nil {
		return nil, err
	}
	return &normalized, nil
}

// DisplayCurrencies is retained for old callers and public response
// compatibility. Storefront display currencies must be read from the market
// settings domain instead.
func (s *CurrencyPolicyService) DisplayCurrencies() ([]string, error) {
	return []string{}, nil
}

func (s *CurrencyPolicyService) PrimaryCurrency() (string, error) {
	return s.BackendEntryCurrency()
}

func (s *CurrencyPolicyService) BackendEntryCurrency() (string, error) {
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
	return currency.Policy{
		PrimaryCurrency:     primaryCurrency,
		DisplayCurrencies:   []string{},
		AvailableCurrencies: currency.Catalog(),
	}, nil
}
