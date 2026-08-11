package admin

import (
	"sort"

	"commerce-platform/internal/domain/currency"

	"github.com/gin-gonic/gin"
)

const adminAuditResourceCurrencyPolicy = "currency_policy"

func (h *CurrencyPolicyHandler) recordCurrencyPolicyAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}

func currencyPolicyAuditDetails(policy currency.Policy) map[string]interface{} {
	details := map[string]interface{}{
		"primary_currency":         currency.NormalizeCode(policy.PrimaryCurrency),
		"display_currencies":       currency.NormalizeCodes(policy.DisplayCurrencies),
		"display_currency_count":   len(currency.NormalizeCodes(policy.DisplayCurrencies)),
		"available_currency_count": len(policy.AvailableCurrencies),
	}
	if len(policy.AvailableCurrencies) > 0 {
		minorUnits := make(map[string]int, len(policy.AvailableCurrencies))
		codes := make([]string, 0, len(policy.AvailableCurrencies))
		seen := map[string]struct{}{}
		for _, option := range policy.AvailableCurrencies {
			code := currency.NormalizeCode(option.Code)
			if code == "" {
				continue
			}
			minorUnits[code] = option.MinorUnits
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			codes = append(codes, code)
		}
		sort.Strings(codes)
		details["available_currencies"] = codes
		details["minor_units"] = minorUnits
	}
	return details
}

func currencyPolicyOldValue(policy *currency.Policy) map[string]interface{} {
	if policy == nil {
		return nil
	}
	return currencyPolicyAuditDetails(*policy)
}
