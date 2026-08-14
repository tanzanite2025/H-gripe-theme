package admin

import (
	"errors"
	"sort"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsConnectorHandler struct {
	connectorService *service.OpsConnectorService
	auditService     adminAuditRecorder
}

func NewOpsConnectorHandler(connectorService *service.OpsConnectorService) *OpsConnectorHandler {
	return &OpsConnectorHandler{connectorService: connectorService}
}

func (h *OpsConnectorHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *OpsConnectorHandler) List(c *gin.Context) {
	if h == nil || h.connectorService == nil {
		apierror.RespondInternalError(c, errors.New("operations connector service is not configured"))
		return
	}
	connectors, err := h.connectorService.List()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"connectors": connectors})
}

func (h *OpsConnectorHandler) Get(c *gin.Context) {
	if h == nil || h.connectorService == nil {
		apierror.RespondInternalError(c, errors.New("operations connector service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid connector id")
	if err != nil {
		return
	}
	connector, err := h.connectorService.Get(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations connector")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, connector)
}

func (h *OpsConnectorHandler) Create(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req opsConnectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsConnector,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      connectorAuditDetails(req),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.connectorService == nil {
		err := errors.New("operations connector service is not configured")
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsConnector,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      connectorAuditDetails(req),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	connector, err := h.connectorService.Create(req.toServiceInput())
	if err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOpsConnector,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      connectorAuditDetails(req),
		})
		respondOpsConnectorError(c, err)
		return
	}
	h.recordOpsConnectorAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceOpsConnector,
		ResourceID: connector.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    connectorAuditDetails(req),
		NewValue:   connector,
	})
	response.Created(c, connector)
}

func (h *OpsConnectorHandler) Update(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.connectorService == nil {
		apierror.RespondInternalError(c, errors.New("operations connector service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid connector id")
	if err != nil {
		return
	}
	oldConnector, _ := h.connectorService.Get(id)
	var req opsConnectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsConnector,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      connectorAuditDetails(req),
			OldValue:     oldConnector,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	connector, err := h.connectorService.Update(id, req.toServiceInput())
	if err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsConnector,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      connectorAuditDetails(req),
			OldValue:     oldConnector,
		})
		respondOpsConnectorError(c, err)
		return
	}
	h.recordOpsConnectorAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsConnector,
		ResourceID: connector.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    connectorAuditDetails(req),
		OldValue:   oldConnector,
		NewValue:   connector,
	})
	response.Success(c, connector)
}

func (h *OpsConnectorHandler) UpdateStatus(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.connectorService == nil {
		apierror.RespondInternalError(c, errors.New("operations connector service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid connector id")
	if err != nil {
		return
	}
	oldConnector, _ := h.connectorService.Get(id)
	var req opsConnectorStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsConnector,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldConnector,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	connector, err := h.connectorService.SetEnabled(id, req.Enabled)
	if err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOpsConnector,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
			OldValue:     oldConnector,
		})
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations connector")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	h.recordOpsConnectorAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOpsConnector,
		ResourceID: connector.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		OldValue:   oldConnector,
		NewValue:   connector,
	})
	response.Success(c, connector)
}

func (h *OpsConnectorHandler) Test(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.connectorService == nil {
		apierror.RespondInternalError(c, errors.New("operations connector service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid connector id")
	if err != nil {
		return
	}
	result, err := h.connectorService.Test(c.Request.Context(), id)
	if err != nil {
		h.recordOpsConnectorAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionProbe,
			Resource:     adminAuditResourceOpsConnector,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations connector")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	status := adminAuditStatusFailed
	if result.Success {
		status = adminAuditStatusSuccess
	}
	h.recordOpsConnectorAudit(c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionProbe,
		Resource:     adminAuditResourceOpsConnector,
		ResourceID:   id,
		Status:       status,
		ErrorMessage: result.Message,
		Changes: map[string]interface{}{
			"success":               result.Success,
			"status_code":           result.StatusCode,
			"credential_configured": result.CredentialConfigured,
		},
	})
	response.Success(c, result)
}

func respondOpsConnectorError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidOpsConnector) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Operations connector")
		return
	}
	apierror.RespondInternalError(c, err)
}

func connectorAuditDetails(req opsConnectorRequest) map[string]interface{} {
	fields := make([]string, 0, len(req.Credentials))
	for key, value := range req.Credentials {
		if key != "" && value != "" {
			fields = append(fields, key)
		}
	}
	sort.Strings(fields)
	return map[string]interface{}{
		"name":              req.Name,
		"provider":          req.Provider,
		"environment":       req.Environment,
		"endpoint":          req.Endpoint,
		"auth_type":         req.AuthType,
		"credential_ref":    req.CredentialRef,
		"credential_fields": fields,
		"scopes":            req.Scopes,
		"status":            req.Status,
		"enabled":           req.Enabled,
		"notes":             req.Notes,
	}
}
