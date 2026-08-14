package admin

import (
	"errors"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsDomainBindingHandler struct {
	domainService  *service.OpsDomainBindingService
	diffService    *service.OpsDomainDiffService
	previewService *service.OpsDomainPreviewService
	syncService    *service.OpsDomainSyncService
	auditService   adminAuditRecorder
}

func NewOpsDomainBindingHandler(
	domainService *service.OpsDomainBindingService,
	diffService *service.OpsDomainDiffService,
	previewService *service.OpsDomainPreviewService,
	syncService *service.OpsDomainSyncService,
) *OpsDomainBindingHandler {
	return &OpsDomainBindingHandler{
		domainService:  domainService,
		diffService:    diffService,
		previewService: previewService,
		syncService:    syncService,
	}
}

func (h *OpsDomainBindingHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *OpsDomainBindingHandler) List(c *gin.Context) {
	if h == nil || h.domainService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain service is not configured"))
		return
	}
	domains, err := h.domainService.List()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"domains": domains})
}

func (h *OpsDomainBindingHandler) Get(c *gin.Context) {
	if h == nil || h.domainService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid domain binding id")
	if err != nil {
		return
	}
	domain, err := h.domainService.Get(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations domain binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, domain)
}

func (h *OpsDomainBindingHandler) Preview(c *gin.Context) {
	if h == nil || h.previewService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain preview service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid domain binding id")
	if err != nil {
		return
	}
	preview, err := h.previewService.Preview(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations domain binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, preview)
}

func (h *OpsDomainBindingHandler) Diff(c *gin.Context) {
	if h == nil || h.diffService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain diff service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid domain binding id")
	if err != nil {
		return
	}
	diff, err := h.diffService.Diff(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations domain binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, diff)
}

func (h *OpsDomainBindingHandler) Create(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req opsDomainBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsDomainBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.domainService == nil {
		err := errors.New("operations domain service is not configured")
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsDomainBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondInternalError(c, err)
		return
	}
	domain, err := h.domainService.Create(req.toServiceInput())
	if err != nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsDomainBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondOpsDomainBindingError(c, err)
		return
	}
	h.recordOpsDomainBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceOpsDomainBinding,
		ResourceID: domain.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    domain,
		NewValue:   domain,
	})
	response.Created(c, domain)
}

func (h *OpsDomainBindingHandler) Update(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.domainService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid domain binding id")
	if err != nil {
		return
	}
	oldDomain, _ := h.domainService.Get(id)

	var req opsDomainBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsDomainBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldDomain,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	domain, err := h.domainService.Update(id, req.toServiceInput())
	if err != nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsDomainBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldDomain,
		})
		respondOpsDomainBindingError(c, err)
		return
	}
	h.recordOpsDomainBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsDomainBinding,
		ResourceID: domain.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    domain,
		OldValue:   oldDomain,
		NewValue:   domain,
	})
	response.Success(c, domain)
}

func (h *OpsDomainBindingHandler) UpdateStatus(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.domainService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid domain binding id")
	if err != nil {
		return
	}
	oldDomain, _ := h.domainService.Get(id)

	var req opsDomainBindingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsDomainBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldDomain,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	domain, err := h.domainService.SetEnabled(id, req.Enabled)
	if err != nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsDomainBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldDomain,
		})
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations domain binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	h.recordOpsDomainBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsDomainBinding,
		ResourceID: domain.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldDomain,
		NewValue:   domain,
	})
	response.Success(c, domain)
}

func (h *OpsDomainBindingHandler) Sync(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.syncService == nil {
		apierror.RespondInternalError(c, errors.New("operations domain sync service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid domain binding id")
	if err != nil {
		return
	}

	result, syncErr := h.syncService.Sync(c.Request.Context(), id)
	if result == nil {
		h.recordOpsDomainBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionProbe,
			Resource:     adminAuditResourceOpsDomainBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: errorMessage(syncErr),
		})
		if repository.IsRecordNotFound(syncErr) {
			apierror.RespondNotFound(c, "Operations domain binding")
			return
		}
		apierror.RespondInternalError(c, syncErr)
		return
	}

	auditStatus := adminAuditStatusSuccess
	if result.ObservedStatus == "error" || syncErr != nil && !errors.Is(syncErr, service.ErrOpsDomainSync) {
		auditStatus = adminAuditStatusFailed
	}
	h.recordOpsDomainBindingAudit(c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionProbe,
		Resource:     adminAuditResourceOpsDomainBinding,
		ResourceID:   id,
		Status:       auditStatus,
		ErrorMessage: result.ObservedError,
		Changes: map[string]interface{}{
			"observed_status":     result.ObservedStatus,
			"observed_target":     result.ObservedTarget,
			"observed_proxy_mode": result.ObservedProxyMode,
			"observed_tls_mode":   result.ObservedTLSMode,
			"observed_source":     result.ObservedSource,
			"dns_record_count":    result.DNSRecordCount,
		},
		NewValue: result,
	})
	response.Success(c, result)
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func respondOpsDomainBindingError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidOpsDomainBinding) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Operations domain binding")
		return
	}
	apierror.RespondInternalError(c, err)
}
