package admin

import "github.com/gin-gonic/gin"

const adminAuditResourceTaxRateRule = "tax_rate_rule"

func (h *TaxRateRuleHandler) recordTaxRateRuleAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	_ = recordAdminAudit(h.auditService, c, event)
}
