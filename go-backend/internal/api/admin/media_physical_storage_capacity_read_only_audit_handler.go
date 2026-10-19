package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *MediaHandler) AuditReadOnlyPhysicalStorageCapacity(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	report, err := h.mediaService.AuditReadOnlyMediaPhysicalCapacity(c.Request.Context())
	if errors.Is(err, service.ErrMediaPhysicalCapacityAuditAlreadyRunning) {
		c.JSON(http.StatusConflict, gin.H{"error": "容量核对正在进行，请稍后重试", "code": "media_capacity_audit_running"})
		return
	}
	if errors.Is(err, repository.ErrMediaPhysicalCapacitySnapshotLimitExceeded) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "登记记录超过本次核对上限，未生成容量结论", "code": "media_capacity_snapshot_limit_exceeded"})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "容量核对未完成，请稍后重试", "code": "media_capacity_audit_unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"report": report}})
}
