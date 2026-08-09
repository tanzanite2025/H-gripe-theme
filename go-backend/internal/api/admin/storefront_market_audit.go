package admin

import "github.com/gin-gonic/gin"

const adminAuditResourceStorefrontMarket = "storefront_market"

func (h *StorefrontMarketHandler) recordStorefrontMarketAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}
