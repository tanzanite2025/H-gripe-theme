package setting

import (
	"encoding/json"
	"strings"
)

const PaymentInstallmentsGroup = "payment_installments"

const (
	PaymentInstallmentsStripeKey = "payment_installments_stripe"
	PaymentInstallmentsPayPalKey = "payment_installments_paypal"
	PaymentInstallmentsWeChatKey = "payment_installments_wechat"
	PaymentInstallmentsAlipayKey = "payment_installments_alipay"
)

type PaymentProviderInstallmentsSettings struct {
	Provider           string   `json:"provider"`
	Enabled            bool     `json:"enabled"`
	PaymentMethodTypes []string `json:"payment_method_types,omitempty"`
	Countries          []string `json:"countries,omitempty"`
	Currencies         []string `json:"currencies,omitempty"`
	MinAmount          float64  `json:"min_amount,omitempty"`
	MaxAmount          float64  `json:"max_amount,omitempty"`
	Notes              string   `json:"notes,omitempty"`
}

type PaymentProviderInstallmentsUpdateRequest struct {
	Enabled            bool     `json:"enabled"`
	PaymentMethodTypes []string `json:"payment_method_types"`
	Countries          []string `json:"countries"`
	Currencies         []string `json:"currencies"`
	MinAmount          float64  `json:"min_amount"`
	MaxAmount          float64  `json:"max_amount"`
	Notes              string   `json:"notes"`
}

func PaymentInstallmentsSettingKey(provider string) string {
	return PaymentInstallmentsKeyPrefix + normalizeInstallmentsProvider(provider)
}

const PaymentInstallmentsKeyPrefix = "payment_installments_"

func normalizeInstallmentsProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func normalizeInstallmentsList(values []string, upper bool) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if upper {
			item = strings.ToUpper(item)
		} else {
			item = strings.ToLower(item)
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		items = append(items, item)
	}
	return items
}

func normalizeInstallmentsAmount(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeInstallmentsPaymentMethodTypes(provider string, values []string) []string {
	items := normalizeInstallmentsList(values, false)
	if normalizeInstallmentsProvider(provider) != "stripe" {
		return items
	}

	allowed := map[string]struct{}{
		"card":              {},
		"klarna":            {},
		"affirm":            {},
		"afterpay_clearpay": {},
	}
	filtered := make([]string, 0, len(items)+1)
	seen := map[string]struct{}{}
	for _, item := range items {
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		filtered = append(filtered, item)
	}
	if len(values) > 0 {
		if _, ok := seen["card"]; !ok {
			filtered = append([]string{"card"}, filtered...)
		} else if len(filtered) > 0 && filtered[0] != "card" {
			reordered := []string{"card"}
			for _, item := range filtered {
				if item != "card" {
					reordered = append(reordered, item)
				}
			}
			filtered = reordered
		}
	}
	return filtered
}

func (request PaymentProviderInstallmentsUpdateRequest) Settings(provider string) PaymentProviderInstallmentsSettings {
	normalizedProvider := normalizeInstallmentsProvider(provider)
	return PaymentProviderInstallmentsSettings{
		Provider:           normalizedProvider,
		Enabled:            request.Enabled,
		PaymentMethodTypes: normalizeInstallmentsPaymentMethodTypes(normalizedProvider, request.PaymentMethodTypes),
		Countries:          normalizeInstallmentsList(request.Countries, true),
		Currencies:         normalizeInstallmentsList(request.Currencies, true),
		MinAmount:          normalizeInstallmentsAmount(request.MinAmount),
		MaxAmount:          normalizeInstallmentsAmount(request.MaxAmount),
		Notes:              strings.TrimSpace(request.Notes),
	}
}

func (settings PaymentProviderInstallmentsSettings) Normalize() PaymentProviderInstallmentsSettings {
	settings.Provider = normalizeInstallmentsProvider(settings.Provider)
	settings.PaymentMethodTypes = normalizeInstallmentsPaymentMethodTypes(settings.Provider, settings.PaymentMethodTypes)
	settings.Countries = normalizeInstallmentsList(settings.Countries, true)
	settings.Currencies = normalizeInstallmentsList(settings.Currencies, true)
	settings.MinAmount = normalizeInstallmentsAmount(settings.MinAmount)
	settings.MaxAmount = normalizeInstallmentsAmount(settings.MaxAmount)
	settings.Notes = strings.TrimSpace(settings.Notes)
	return settings
}

func (settings PaymentProviderInstallmentsSettings) Configured() bool {
	settings = settings.Normalize()
	return settings.Enabled ||
		len(settings.PaymentMethodTypes) > 0 ||
		len(settings.Countries) > 0 ||
		len(settings.Currencies) > 0 ||
		settings.MinAmount > 0 ||
		settings.MaxAmount > 0 ||
		settings.Notes != ""
}

func PaymentProviderInstallmentsSettingsFromValue(provider, value string) (PaymentProviderInstallmentsSettings, error) {
	settings := PaymentProviderInstallmentsSettings{
		Provider: normalizeInstallmentsProvider(provider),
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return settings.Normalize(), nil
	}
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return PaymentProviderInstallmentsSettings{}, err
	}
	if settings.Provider == "" {
		settings.Provider = provider
	}
	return settings.Normalize(), nil
}

func (settings PaymentProviderInstallmentsSettings) Value() (string, error) {
	settings = settings.Normalize()
	payload, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
