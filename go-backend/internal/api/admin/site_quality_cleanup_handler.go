package admin

import (
	"errors"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const adminAuditResourceSiteQualityJob = "preflight_site_quality_job"

func (h *SiteQualityHandler) CleanupSiteQualityJobs(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.siteQualityEngine == nil {
		err := errors.New("site quality engine is not configured")
		if h != nil {
			h.recordSiteQualityJobAudit(c, adminAuditEvent{
				StartedAt:    startedAt,
				Action:       adminAuditActionDelete,
				Resource:     adminAuditResourceSiteQualityJob,
				Status:       adminAuditStatusFailed,
				ErrorMessage: err.Error(),
			})
		}
		apierror.RespondInternalError(c, err)
		return
	}

	result, err := h.siteQualityEngine.CleanupTerminalJobs()
	if err != nil {
		h.recordSiteQualityJobAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceSiteQualityJob,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondInternalError(c, err)
		return
	}

	h.recordSiteQualityJobAudit(c, adminAuditEvent{
		StartedAt: startedAt,
		Action:    adminAuditActionDelete,
		Resource:  adminAuditResourceSiteQualityJob,
		Status:    adminAuditStatusSuccess,
		Changes:   result,
		NewValue:  result,
	})
	response.Success(c, result)
}

func (h *SiteQualityHandler) recordSiteQualityJobAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}
