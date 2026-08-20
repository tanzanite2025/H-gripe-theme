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
	Notes              string   `json:"notes,omitempty"`
}

type PaymentProviderInstallmentsUpdateRequest struct {
	Enabled            bool     `json:"enabled"`
	PaymentMethodTypes []string `json:"payment_method_types"`
	Countries          []string `json:"countries"`
	Currencies         []string `json:"currencies"`
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

func (request PaymentProviderInstallmentsUpdateRequest) Settings(provider string) PaymentProviderInstallmentsSettings {
	return PaymentProviderInstallmentsSettings{
		Provider:           normalizeInstallmentsProvider(provider),
		Enabled:            request.Enabled,
		PaymentMethodTypes: normalizeInstallmentsList(request.PaymentMethodTypes, false),
		Countries:          normalizeInstallmentsList(request.Countries, true),
		Currencies:         normalizeInstallmentsList(request.Currencies, true),
		Notes:              strings.TrimSpace(request.Notes),
	}
}

func (settings PaymentProviderInstallmentsSettings) Normalize() PaymentProviderInstallmentsSettings {
	settings.Provider = normalizeInstallmentsProvider(settings.Provider)
	settings.PaymentMethodTypes = normalizeInstallmentsList(settings.PaymentMethodTypes, false)
	settings.Countries = normalizeInstallmentsList(settings.Countries, true)
	settings.Currencies = normalizeInstallmentsList(settings.Currencies, true)
	settings.Notes = strings.TrimSpace(settings.Notes)
	return settings
}

func (settings PaymentProviderInstallmentsSettings) Configured() bool {
	settings = settings.Normalize()
	return settings.Enabled ||
		len(settings.PaymentMethodTypes) > 0 ||
		len(settings.Countries) > 0 ||
		len(settings.Currencies) > 0 ||
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
