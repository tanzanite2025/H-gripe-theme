package currency

import (
	"encoding/json"
	"sort"
	"strings"
)

// Policy is the store-wide currency policy. Gateway capability is evaluated
// against this policy; it is not the source of the policy itself.
type Policy struct {
	AccountingCurrency   string           `json:"accounting_currency"`
	DefaultOrderCurrency string           `json:"default_order_currency"`
	AcceptedCurrencies   []string         `json:"accepted_currencies"`
	AvailableCurrencies  []CurrencyOption `json:"available_currencies"`
}

type CurrencyOption struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	MinorUnits int    `json:"minor_units"`
}

func (p *Policy) UnmarshalJSON(data []byte) error {
	var raw struct {
		AccountingCurrency       string           `json:"accounting_currency"`
		DefaultOrderCurrency     string           `json:"default_order_currency"`
		DefaultCheckoutCurrency  string           `json:"default_checkout_currency"`
		AcceptedCurrencies       []string         `json:"accepted_currencies"`
		LegacyCheckoutCurrencies []string         `json:"checkout_currencies"`
		AvailableCurrencies      []CurrencyOption `json:"available_currencies"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.AccountingCurrency = raw.AccountingCurrency
	p.DefaultOrderCurrency = raw.DefaultOrderCurrency
	if p.DefaultOrderCurrency == "" {
		p.DefaultOrderCurrency = raw.DefaultCheckoutCurrency
	}
	p.AcceptedCurrencies = raw.AcceptedCurrencies
	if len(p.AcceptedCurrencies) == 0 && len(raw.LegacyCheckoutCurrencies) > 0 {
		p.AcceptedCurrencies = raw.LegacyCheckoutCurrencies
	}
	p.AvailableCurrencies = raw.AvailableCurrencies
	return nil
}

func NormalizeCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func IsValidCode(value string) bool {
	value = NormalizeCode(value)
	if len(value) != 3 {
		return false
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func IsCatalogCode(value string) bool {
	_, ok := MinorUnits(value)
	return ok
}

func NormalizeCodes(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		code := NormalizeCode(value)
		if !IsValidCode(code) {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	sort.Strings(result)
	return result
}

func MinorUnits(value string) (int, bool) {
	code := NormalizeCode(value)
	for _, option := range Catalog() {
		if option.Code == code {
			return option.MinorUnits, true
		}
	}
	return 0, false
}

// Catalog is deliberately a product-facing catalog, not a claim that every
// enabled gateway/account can settle every listed currency.
func Catalog() []CurrencyOption {
	return []CurrencyOption{
		{Code: "USD", Name: "US Dollar", MinorUnits: 2},
		{Code: "EUR", Name: "Euro", MinorUnits: 2},
		{Code: "GBP", Name: "British Pound", MinorUnits: 2},
		{Code: "CAD", Name: "Canadian Dollar", MinorUnits: 2},
		{Code: "AUD", Name: "Australian Dollar", MinorUnits: 2},
		{Code: "NZD", Name: "New Zealand Dollar", MinorUnits: 2},
		{Code: "JPY", Name: "Japanese Yen", MinorUnits: 0},
		{Code: "CNY", Name: "Chinese Yuan", MinorUnits: 2},
		{Code: "HKD", Name: "Hong Kong Dollar", MinorUnits: 2},
		{Code: "SGD", Name: "Singapore Dollar", MinorUnits: 2},
		{Code: "CHF", Name: "Swiss Franc", MinorUnits: 2},
		{Code: "SEK", Name: "Swedish Krona", MinorUnits: 2},
		{Code: "NOK", Name: "Norwegian Krone", MinorUnits: 2},
		{Code: "DKK", Name: "Danish Krone", MinorUnits: 2},
		{Code: "PLN", Name: "Polish Zloty", MinorUnits: 2},
		{Code: "CZK", Name: "Czech Koruna", MinorUnits: 2},
		{Code: "HUF", Name: "Hungarian Forint", MinorUnits: 2},
		{Code: "MXN", Name: "Mexican Peso", MinorUnits: 2},
		{Code: "BRL", Name: "Brazilian Real", MinorUnits: 2},
		{Code: "INR", Name: "Indian Rupee", MinorUnits: 2},
		{Code: "AED", Name: "UAE Dirham", MinorUnits: 2},
		{Code: "SAR", Name: "Saudi Riyal", MinorUnits: 2},
		{Code: "ZAR", Name: "South African Rand", MinorUnits: 2},
		{Code: "KRW", Name: "South Korean Won", MinorUnits: 0},
		{Code: "TWD", Name: "New Taiwan Dollar", MinorUnits: 2},
		{Code: "THB", Name: "Thai Baht", MinorUnits: 2},
		{Code: "MYR", Name: "Malaysian Ringgit", MinorUnits: 2},
		{Code: "IDR", Name: "Indonesian Rupiah", MinorUnits: 2},
		{Code: "PHP", Name: "Philippine Peso", MinorUnits: 2},
		{Code: "TRY", Name: "Turkish Lira", MinorUnits: 2},
		{Code: "ILS", Name: "Israeli New Shekel", MinorUnits: 2},
		{Code: "RON", Name: "Romanian Leu", MinorUnits: 2},
		{Code: "BGN", Name: "Bulgarian Lev", MinorUnits: 2},
		{Code: "ISK", Name: "Icelandic Krona", MinorUnits: 2},
		{Code: "CLP", Name: "Chilean Peso", MinorUnits: 0},
		{Code: "COP", Name: "Colombian Peso", MinorUnits: 2},
	}
}
