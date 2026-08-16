package admin

import (
	"errors"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsVPSBindingHandler struct {
	vpsService           *service.OpsVPSBindingService
	hostingerSyncService *service.OpsHostingerSyncService
	auditService         adminAuditRecorder
}

func NewOpsVPSBindingHandler(vpsService *service.OpsVPSBindingService) *OpsVPSBindingHandler {
	return &OpsVPSBindingHandler{vpsService: vpsService}
}

func (h *OpsVPSBindingHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *OpsVPSBindingHandler) ConfigureSyncService(syncService *service.OpsHostingerSyncService) {
	if h == nil {
		return
	}
	h.hostingerSyncService = syncService
}

func (h *OpsVPSBindingHandler) List(c *gin.Context) {
	if h == nil || h.vpsService == nil {
		apierror.RespondInternalError(c, errors.New("operations VPS service is not configured"))
		return
	}
	records, err := h.vpsService.ListForEnvironment(strings.TrimSpace(c.Query("environment")))
	if err != nil {
		if errors.Is(err, service.ErrInvalidOpsVPSEnvironment) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"vps": records})
}

func (h *OpsVPSBindingHandler) Get(c *gin.Context) {
	if h == nil || h.vpsService == nil {
		apierror.RespondInternalError(c, errors.New("operations VPS service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid VPS binding id")
	if err != nil {
		return
	}
	record, err := h.vpsService.Get(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations VPS binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, record)
}

func (h *OpsVPSBindingHandler) Create(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req opsVPSBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsVPSBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.vpsService == nil {
		err := errors.New("operations VPS service is not configured")
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsVPSBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondInternalError(c, err)
		return
	}
	record, err := h.vpsService.Create(req.toServiceInput())
	if err != nil {
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsVPSBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondOpsVPSBindingError(c, err)
		return
	}
	h.recordOpsVPSBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceOpsVPSBinding,
		ResourceID: record.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		NewValue:   record,
	})
	response.Created(c, record)
}

func (h *OpsVPSBindingHandler) Update(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.vpsService == nil {
		apierror.RespondInternalError(c, errors.New("operations VPS service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid VPS binding id")
	if err != nil {
		return
	}
	oldRecord, _ := h.vpsService.Get(id)
	var req opsVPSBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsVPSBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldRecord,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	record, err := h.vpsService.Update(id, req.toServiceInput())
	if err != nil {
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsVPSBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldRecord,
		})
		respondOpsVPSBindingError(c, err)
		return
	}
	h.recordOpsVPSBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsVPSBinding,
		ResourceID: record.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldRecord,
		NewValue:   record,
	})
	response.Success(c, record)
}

func (h *OpsVPSBindingHandler) UpdateStatus(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.vpsService == nil {
		apierror.RespondInternalError(c, errors.New("operations VPS service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid VPS binding id")
	if err != nil {
		return
	}
	oldRecord, _ := h.vpsService.Get(id)
	var req opsVPSBindingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	record, err := h.vpsService.SetEnabled(id, req.Enabled)
	if err != nil {
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsVPSBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldRecord,
		})
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations VPS binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	h.recordOpsVPSBindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsVPSBinding,
		ResourceID: record.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldRecord,
		NewValue:   record,
	})
	response.Success(c, record)
}

func (h *OpsVPSBindingHandler) Sync(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.hostingerSyncService == nil {
		err := errors.New("operations Hostinger sync service is not configured")
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionProbe,
			Resource:     adminAuditResourceOpsVPSBinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := parseUintParam(c, "id", "invalid VPS binding id")
	if err != nil {
		return
	}

	result, syncErr := h.hostingerSyncService.SyncVPS(c.Request.Context(), id)
	if result == nil {
		h.recordOpsVPSBindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionProbe,
			Resource:     adminAuditResourceOpsVPSBinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: errorMessage(syncErr),
		})
		if repository.IsRecordNotFound(syncErr) {
			apierror.RespondNotFound(c, "Operations VPS binding")
			return
		}
		apierror.RespondInternalError(c, syncErr)
		return
	}

	auditStatus := adminAuditStatusSuccess
	if result.ObservedError != "" || syncErr != nil && !errors.Is(syncErr, service.ErrOpsHostingerSync) {
		auditStatus = adminAuditStatusFailed
	}
	h.recordOpsVPSBindingAudit(c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionProbe,
		Resource:     adminAuditResourceOpsVPSBinding,
		ResourceID:   id,
		Status:       auditStatus,
		ErrorMessage: result.ObservedError,
		Changes: map[string]interface{}{
			"observed_status":    result.ObservedStatus,
			"remote_state":       result.RemoteState,
			"observed_source":    result.ObservedSource,
			"hostname_confirmed": result.Hostname != "",
			"ipv4_confirmed":     result.IPv4 != "",
			"operating_system":   result.OperatingSystem,
		},
		NewValue: result,
	})
	response.Success(c, result)
}

func respondOpsVPSBindingError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidOpsVPSBinding) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Operations VPS binding")
		return
	}
	apierror.RespondInternalError(c, err)
}
