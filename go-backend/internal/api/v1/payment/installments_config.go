package payment

import (
	"strings"

	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

func (h *Handler) resolveStripePaymentMethodTypes(orderCountry, orderCurrency string, orderAmount float64, fallback []string) ([]string, error) {
	defaultTypes := finalizeStripePaymentMethodTypes(fallback, orderCurrency)
	if h == nil || h.settingsService == nil {
		return defaultTypes, nil
	}

	record, err := h.settingsService.GetDomainManagedSetting(settingdomain.PaymentInstallmentsStripeKey, "global")
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return defaultTypes, nil
		}
		return nil, err
	}

	settings, err := settingdomain.PaymentProviderInstallmentsSettingsFromValue("stripe", record.Value)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled || len(settings.PaymentMethodTypes) == 0 {
		return defaultTypes, nil
	}
	if len(settings.Countries) > 0 && !containsNormalizedString(settings.Countries, orderCountry, true) {
		return defaultTypes, nil
	}
	if len(settings.Currencies) > 0 && !containsNormalizedString(settings.Currencies, orderCurrency, true) {
		return defaultTypes, nil
	}
	if settings.MinAmount > 0 && orderAmount < settings.MinAmount {
		return defaultTypes, nil
	}
	if settings.MaxAmount > 0 && orderAmount > settings.MaxAmount {
		return defaultTypes, nil
	}

	return finalizeStripePaymentMethodTypes(settings.PaymentMethodTypes, orderCurrency), nil
}

func normalizeStripePaymentMethodTypes(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.ToLower(strings.TrimSpace(value))
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func finalizeStripePaymentMethodTypes(values []string, currency string) []string {
	return filterStripePaymentMethodTypesForCurrency(ensureStripeCardPaymentMethodType(normalizeStripePaymentMethodTypes(values)), currency)
}

func ensureStripeCardPaymentMethodType(values []string) []string {
	if containsStripePaymentMethodType(values, "card") {
		return values
	}
	result := make([]string, 0, len(values)+1)
	result = append(result, "card")
	result = append(result, values...)
	return result
}

func containsStripePaymentMethodType(values []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == target {
			return true
		}
	}
	return false
}

func filterStripePaymentMethodTypesForCurrency(values []string, currency string) []string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if stripePaymentMethodTypeSupportsCurrency(value, currency) {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return []string{"card"}
	}
	return result
}

func stripePaymentMethodTypeSupportsCurrency(methodType, currency string) bool {
	if currency == "" {
		return true
	}
	supportedCurrencies, restricted := stripePaymentMethodCurrencySupport()[strings.ToLower(strings.TrimSpace(methodType))]
	if !restricted {
		return true
	}
	_, ok := supportedCurrencies[currency]
	return ok
}

func stripePaymentMethodCurrencySupport() map[string]map[string]struct{} {
	return map[string]map[string]struct{}{
		"affirm": {
			"CAD": {},
			"USD": {},
		},
		"afterpay_clearpay": {
			"AUD": {},
			"CAD": {},
			"GBP": {},
			"NZD": {},
			"USD": {},
		},
		"klarna": {
			"AUD": {},
			"CAD": {},
			"CHF": {},
			"CZK": {},
			"DKK": {},
			"EUR": {},
			"GBP": {},
			"NOK": {},
			"NZD": {},
			"PLN": {},
			"RON": {},
			"SEK": {},
			"USD": {},
		},
	}
}

func containsNormalizedString(values []string, target string, upper bool) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	if upper {
		target = strings.ToUpper(target)
	} else {
		target = strings.ToLower(target)
	}
	for _, value := range values {
		item := strings.TrimSpace(value)
		if upper {
			item = strings.ToUpper(item)
		} else {
			item = strings.ToLower(item)
		}
		if item == target {
			return true
		}
	}
	return false
}
