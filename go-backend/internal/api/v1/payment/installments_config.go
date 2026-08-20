package payment

import (
	"strings"

	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

func (h *Handler) resolveStripePaymentMethodTypes(orderCountry, orderCurrency string, fallback []string) ([]string, error) {
	defaultTypes := normalizeStripePaymentMethodTypes(fallback)
	if len(defaultTypes) == 0 {
		defaultTypes = []string{"card"}
	}
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

	return settings.PaymentMethodTypes, nil
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
