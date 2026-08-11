package payment

import (
	"fmt"

	"commerce-platform/internal/domain/currency"
)

var paypalSupportedCurrencies = []string{
	"AUD",
	"BRL",
	"CAD",
	"CNY",
	"CZK",
	"DKK",
	"EUR",
	"GBP",
	"HKD",
	"ILS",
	"JPY",
	"MXN",
	"MYR",
	"NOK",
	"NZD",
	"PHP",
	"PLN",
	"SEK",
	"SGD",
	"THB",
	"USD",
}

func SupportedCurrenciesForProvider(provider GatewayType) []string {
	switch provider {
	case GatewayStripe:
		return catalogCurrencyCodes()
	case GatewayPayPal:
		return cloneCurrencyCodes(paypalSupportedCurrencies)
	case GatewayAlipay, GatewayWechat:
		return []string{"CNY"}
	default:
		return nil
	}
}

func GatewaySupportsCurrency(provider GatewayType, code string) bool {
	code = currency.NormalizeCode(code)
	if code == "" {
		return false
	}
	for _, supported := range SupportedCurrenciesForProvider(provider) {
		if supported == code {
			return true
		}
	}
	return false
}

func ValidateGatewayCurrency(provider GatewayType, code string) error {
	code = currency.NormalizeCode(code)
	if !currency.IsValidCode(code) || !currency.IsCatalogCode(code) {
		return fmt.Errorf("unsupported currency code")
	}
	if GatewaySupportsCurrency(provider, code) {
		return nil
	}
	if len(SupportedCurrenciesForProvider(provider)) == 0 {
		return fmt.Errorf("unsupported payment provider %s", provider)
	}
	return fmt.Errorf("%s does not support currency %s", provider, code)
}

func catalogCurrencyCodes() []string {
	options := currency.Catalog()
	codes := make([]string, 0, len(options))
	for _, option := range options {
		codes = append(codes, option.Code)
	}
	return codes
}

func cloneCurrencyCodes(values []string) []string {
	result := make([]string, len(values))
	copy(result, values)
	return result
}
