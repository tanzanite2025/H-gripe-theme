package admin

import "github.com/gin-gonic/gin"

const adminAuditResourceOpsProjectBinding = "ops_project_binding"

func (h *OpsProjectBindingHandler) recordOpsProjectBindingAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}
