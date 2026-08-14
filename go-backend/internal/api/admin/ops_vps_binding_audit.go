package admin

import "github.com/gin-gonic/gin"

const adminAuditResourceOpsVPSBinding = "ops_vps_binding"

func (h *OpsVPSBindingHandler) recordOpsVPSBindingAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}
