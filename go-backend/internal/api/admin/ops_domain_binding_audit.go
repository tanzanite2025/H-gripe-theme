package admin

import "github.com/gin-gonic/gin"

const adminAuditResourceOpsDomainBinding = "ops_domain_binding"

func (h *OpsDomainBindingHandler) recordOpsDomainBindingAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}
