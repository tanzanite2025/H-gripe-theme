package admin

import (
	"errors"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const adminAuditResourceOpsDeploymentWorkflow = "ops_deployment_workflow"

type OpsDeploymentWorkflowHandler struct {
	service      *service.OpsDeploymentWorkflowService
	auditService adminAuditRecorder
}

type opsDeploymentWorkflowCreateRequest struct {
	ProjectID    uint   `json:"project_id"`
	RequestedRef string `json:"requested_ref"`
	Mode         string `json:"mode"`
}

func NewOpsDeploymentWorkflowHandler(workflowService *service.OpsDeploymentWorkflowService) *OpsDeploymentWorkflowHandler {
	return &OpsDeploymentWorkflowHandler{service: workflowService}
}

func (h *OpsDeploymentWorkflowHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *OpsDeploymentWorkflowHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment workflow service is not configured"))
		return
	}
	projectID, err := parseOptionalWorkflowProjectID(c)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid project binding id")
		return
	}
	workflows, err := h.service.List(projectID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"workflows": workflows})
}

func (h *OpsDeploymentWorkflowHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment workflow service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid deployment workflow id")
	if err != nil {
		return
	}
	workflow, err := h.service.Get(id)
	if err != nil {
		respondOpsDeploymentWorkflowError(c, err)
		return
	}
	response.Success(c, workflow)
}

func (h *OpsDeploymentWorkflowHandler) Create(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req opsDeploymentWorkflowCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionSubmit,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if req.ProjectID == 0 {
		apierror.RespondBadRequest(c, "project_id is required")
		return
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = ops.DeploymentWorkflowModeDryRun
	}
	if mode != ops.DeploymentWorkflowModeDryRun && mode != ops.DeploymentWorkflowModeProduction {
		apierror.RespondBadRequest(c, "workflow mode must be dry_run or production")
		return
	}
	if mode == ops.DeploymentWorkflowModeProduction && !workflowActorHasPermission(c, auth.PermOpsDeployExecute) {
		apierror.RespondForbidden(c)
		return
	}
	if h == nil || h.service == nil {
		err := errors.New("operations deployment workflow service is not configured")
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionSubmit,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			ResourceID:   req.ProjectID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		apierror.RespondInternalError(c, err)
		return
	}
	input := service.OpsDeploymentWorkflowCreateInput{
		ProjectID:    req.ProjectID,
		RequestedRef: req.RequestedRef,
		CreatedByID:  c.GetUint("user_id"),
		CreatedBy:    workflowActorName(c),
	}
	var workflow *ops.DeploymentWorkflowRun
	var err error
	if mode == ops.DeploymentWorkflowModeProduction {
		workflow, err = h.service.CreateProduction(input)
	} else {
		workflow, err = h.service.CreateDryRun(input)
	}
	if err != nil {
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionSubmit,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			ResourceID:   req.ProjectID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      req,
		})
		respondOpsDeploymentWorkflowError(c, err)
		return
	}
	h.recordAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionSubmit,
		Resource:   adminAuditResourceOpsDeploymentWorkflow,
		ResourceID: workflow.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    req,
		NewValue:   workflowAuditValue(workflow),
	})
	response.Success(c, workflow)
}

func (h *OpsDeploymentWorkflowHandler) CreateDryRun(c *gin.Context) {
	h.Create(c)
}

func (h *OpsDeploymentWorkflowHandler) Validate(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, err := parseUintParam(c, "id", "invalid deployment workflow id")
	if err != nil {
		return
	}
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment workflow service is not configured"))
		return
	}
	if workflow, getErr := h.service.Get(id); getErr == nil &&
		workflow.Mode == ops.DeploymentWorkflowModeProduction &&
		!workflowActorHasPermission(c, auth.PermOpsDeployExecute) {
		apierror.RespondForbidden(c)
		return
	}
	workflow, err := h.service.Validate(id)
	if err != nil {
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionRecompute,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondOpsDeploymentWorkflowError(c, err)
		return
	}
	h.recordAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionRecompute,
		Resource:   adminAuditResourceOpsDeploymentWorkflow,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		NewValue:   workflowAuditValue(workflow),
	})
	response.Success(c, workflow)
}

func (h *OpsDeploymentWorkflowHandler) Approve(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, err := parseUintParam(c, "id", "invalid deployment workflow id")
	if err != nil {
		return
	}
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment workflow service is not configured"))
		return
	}
	workflow, err := h.service.Approve(id, service.OpsDeploymentWorkflowActor{
		UserID:   c.GetUint("user_id"),
		Username: workflowActorName(c),
	})
	if err != nil {
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionApprove,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondOpsDeploymentWorkflowError(c, err)
		return
	}
	h.recordAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionApprove,
		Resource:   adminAuditResourceOpsDeploymentWorkflow,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		NewValue:   workflowAuditValue(workflow),
	})
	response.Success(c, workflow)
}

func (h *OpsDeploymentWorkflowHandler) Execute(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, err := parseUintParam(c, "id", "invalid deployment workflow id")
	if err != nil {
		return
	}
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment workflow service is not configured"))
		return
	}
	if workflow, getErr := h.service.Get(id); getErr == nil &&
		workflow.Mode == ops.DeploymentWorkflowModeProduction &&
		!workflowActorHasPermission(c, auth.PermOpsDeployExecute) {
		apierror.RespondForbidden(c)
		return
	}
	workflow, err := h.service.Execute(c.Request.Context(), id)
	if err != nil {
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondOpsDeploymentWorkflowError(c, err)
		return
	}
	h.recordAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionExecute,
		Resource:   adminAuditResourceOpsDeploymentWorkflow,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		NewValue:   workflowAuditValue(workflow),
	})
	response.Success(c, workflow)
}

func (h *OpsDeploymentWorkflowHandler) ExecuteDryRun(c *gin.Context) {
	h.Execute(c)
}

func (h *OpsDeploymentWorkflowHandler) Cancel(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, err := parseUintParam(c, "id", "invalid deployment workflow id")
	if err != nil {
		return
	}
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment workflow service is not configured"))
		return
	}
	workflow, err := h.service.Cancel(id)
	if err != nil {
		h.recordAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionRevoke,
			Resource:     adminAuditResourceOpsDeploymentWorkflow,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		respondOpsDeploymentWorkflowError(c, err)
		return
	}
	h.recordAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionRevoke,
		Resource:   adminAuditResourceOpsDeploymentWorkflow,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		NewValue:   workflowAuditValue(workflow),
	})
	response.Success(c, workflow)
}

func (h *OpsDeploymentWorkflowHandler) recordAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}

func respondOpsDeploymentWorkflowError(c *gin.Context, err error) {
	switch {
	case repository.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Operations deployment workflow")
	case errors.Is(err, service.ErrOpsDeploymentWorkflowInvalidTransition),
		errors.Is(err, service.ErrOpsDeploymentWorkflowPreflightBlocked),
		errors.Is(err, service.ErrOpsDeploymentWorkflowUnsupportedMode),
		errors.Is(err, service.ErrOpsDeploymentWorkflowUnsupportedRef),
		errors.Is(err, repository.ErrOpsDeploymentWorkflowLockHeld):
		apierror.RespondConflict(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}

func parseOptionalWorkflowProjectID(c *gin.Context) (uint, error) {
	raw := strings.TrimSpace(c.Query("project_id"))
	if raw == "" {
		return 0, nil
	}
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func workflowActorName(c *gin.Context) string {
	if value := strings.TrimSpace(c.GetString("username")); value != "" {
		return value
	}
	if value := strings.TrimSpace(c.GetString("email")); value != "" {
		return value
	}
	return "admin"
}

func workflowActorHasPermission(c *gin.Context, permission auth.Permission) bool {
	if c == nil {
		return false
	}
	roleValue := strings.TrimSpace(c.GetString("user_role"))
	return roleValue != "" && auth.Role(roleValue).HasPermission(permission)
}

func workflowAuditValue(workflow *ops.DeploymentWorkflowRun) map[string]interface{} {
	if workflow == nil {
		return nil
	}
	return map[string]interface{}{
		"id":               workflow.ID,
		"project_id":       workflow.ProjectID,
		"project":          workflow.ProjectName,
		"mode":             workflow.Mode,
		"status":           workflow.Status,
		"preflight_status": workflow.PreflightStatus,
		"requested_ref":    workflow.RequestedRef,
	}
}
