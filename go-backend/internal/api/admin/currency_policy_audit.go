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
	backendEntryCurrency := currency.NormalizeCode(policy.PrimaryCurrency)
	details := map[string]interface{}{
		"primary_currency":          backendEntryCurrency,
		"backend_entry_currency":    backendEntryCurrency,
		"available_currency_count":  len(policy.AvailableCurrencies),
		"display_currencies_source": "storefront_markets",
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
