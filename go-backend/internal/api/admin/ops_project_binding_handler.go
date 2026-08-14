package admin

import (
	"errors"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsProjectBindingHandler struct {
	projectService       *service.OpsProjectBindingService
	hostingerSyncService *service.OpsHostingerSyncService
	auditService         adminAuditRecorder
}

func NewOpsProjectBindingHandler(projectService *service.OpsProjectBindingService) *OpsProjectBindingHandler {
	return &OpsProjectBindingHandler{projectService: projectService}
}

func (h *OpsProjectBindingHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *OpsProjectBindingHandler) ConfigureSyncService(syncService *service.OpsHostingerSyncService) {
	if h == nil {
		return
	}
	h.hostingerSyncService = syncService
}

func (h *OpsProjectBindingHandler) List(c *gin.Context) {
	if h == nil || h.projectService == nil {
		apierror.RespondInternalError(c, errors.New("operations project service is not configured"))
		return
	}
	records, err := h.projectService.List()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"projects": records})
}

func (h *OpsProjectBindingHandler) Get(c *gin.Context) {
	if h == nil || h.projectService == nil {
		apierror.RespondInternalError(c, errors.New("operations project service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid project binding id")
	if err != nil {
		return
	}
	record, err := h.projectService.Get(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations project binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, record)
}

func (h *OpsProjectBindingHandler) Create(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req opsProjectBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsProjectBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.projectService == nil {
		err := errors.New("operations project service is not configured")
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsProjectBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondInternalError(c, err)
		return
	}
	record, err := h.projectService.Create(req.toServiceInput())
	if err != nil {
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsProjectBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondOpsProjectBindingError(c, err)
		return
	}
	h.recordOpsProjectBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceOpsProjectBinding,
		ResourceID: record.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		NewValue:   record,
	})
	response.Created(c, record)
}

func (h *OpsProjectBindingHandler) Update(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.projectService == nil {
		apierror.RespondInternalError(c, errors.New("operations project service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid project binding id")
	if err != nil {
		return
	}
	oldRecord, _ := h.projectService.Get(id)
	var req opsProjectBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsProjectBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldRecord,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	record, err := h.projectService.Update(id, req.toServiceInput())
	if err != nil {
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsProjectBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldRecord,
		})
		respondOpsProjectBindingError(c, err)
		return
	}
	h.recordOpsProjectBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsProjectBinding,
		ResourceID: record.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldRecord,
		NewValue:   record,
	})
	response.Success(c, record)
}

func (h *OpsProjectBindingHandler) UpdateStatus(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.projectService == nil {
		apierror.RespondInternalError(c, errors.New("operations project service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid project binding id")
	if err != nil {
		return
	}
	oldRecord, _ := h.projectService.Get(id)
	var req opsProjectBindingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	record, err := h.projectService.SetEnabled(id, req.Enabled)
	if err != nil {
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsProjectBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldRecord,
		})
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations project binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	h.recordOpsProjectBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsProjectBinding,
		ResourceID: record.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldRecord,
		NewValue:   record,
	})
	response.Success(c, record)
}

func (h *OpsProjectBindingHandler) Sync(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.hostingerSyncService == nil {
		err := errors.New("operations Hostinger sync service is not configured")
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionProbe,
			Resource:     adminAuditResourceOpsProjectBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := parseUintParam(c, "id", "invalid project binding id")
	if err != nil {
		return
	}

	result, syncErr := h.hostingerSyncService.SyncProject(c.Request.Context(), id)
	if result == nil {
		h.recordOpsProjectBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionProbe,
			Resource:     adminAuditResourceOpsProjectBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: errorMessage(syncErr),
		})
		if repository.IsRecordNotFound(syncErr) {
			apierror.RespondNotFound(c, "Operations project binding")
			return
		}
		apierror.RespondInternalError(c, syncErr)
		return
	}

	auditStatus := adminAuditStatusSuccess
	if result.ObservedError != "" || syncErr != nil && !errors.Is(syncErr, service.ErrOpsHostingerSync) {
		auditStatus = adminAuditStatusFailed
	}
	h.recordOpsProjectBindingAudit(c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionProbe,
		Resource:     adminAuditResourceOpsProjectBinding,
		ResourceID:   id,
		Status:       auditStatus,
		ErrorMessage: result.ObservedError,
		Changes: map[string]interface{}{
			"health_status":           result.HealthStatus,
			"remote_state":            result.RemoteState,
			"container_count":         result.ContainerCount,
			"running_container_count": result.RunningContainerCount,
			"healthy_container_count": result.HealthyContainerCount,
			"observed_source":         result.ObservedSource,
		},
		NewValue: result,
	})
	response.Success(c, result)
}

func respondOpsProjectBindingError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidOpsProjectBinding) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Operations project binding")
		return
	}
	apierror.RespondInternalError(c, err)
}
