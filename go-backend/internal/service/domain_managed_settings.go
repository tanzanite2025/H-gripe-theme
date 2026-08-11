package service

import (
	"strings"

	analyticsdomain "tanzanite/internal/domain/analytics"
	seodomain "tanzanite/internal/domain/seo"
	"tanzanite/internal/domain/setting"
)

var domainManagedSettingGroups = map[string]struct{}{
	"loyalty":             {},
	"redeem":              {},
	"currency":            {},
	"payment_secret":      {},
	setting.PayPalDisputeInvoiceSellerProfileGroup: {},
	seodomain.Group:       {},
	analyticsdomain.Group: {},
}

func IsDomainManagedSettingGroup(group string) bool {
	_, managed := domainManagedSettingGroups[strings.ToLower(strings.TrimSpace(group))]
	return managed
}

func IsDomainManagedSettingKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.HasPrefix(normalized, "tz_loyalty_") ||
		strings.HasPrefix(normalized, "tz_redeem_") ||
		strings.HasPrefix(normalized, "currency_") ||
		strings.HasPrefix(normalized, "payment_gateway_") ||
		strings.HasPrefix(normalized, "paypal_dispute_invoice_seller_") ||
		normalized == seodomain.HomeKeys.MetaTitle ||
		normalized == seodomain.HomeKeys.MetaDescription ||
		normalized == "google_analytics" ||
		normalized == "google_tag_manager"
}

func FilterDomainManagedSettings(settings []setting.Setting) []setting.Setting {
	filtered := make([]setting.Setting, 0, len(settings))
	for _, item := range settings {
		if IsDomainManagedSettingGroup(item.Group) || IsDomainManagedSettingKey(item.Key) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func FilterDomainManagedSettingGroups(groups []string) []string {
	filtered := make([]string, 0, len(groups))
	for _, group := range groups {
		if IsDomainManagedSettingGroup(group) {
			continue
		}
		filtered = append(filtered, group)
	}
	return filtered
}
