package admin

import "github.com/gin-gonic/gin"

const adminAuditResourceOpsConnector = "ops_connector"

func (h *OpsConnectorHandler) recordOpsConnectorAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}
