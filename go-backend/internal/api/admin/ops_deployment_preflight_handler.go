package admin

import (
	"errors"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsDeploymentPreflightHandler struct {
	service *service.OpsDeploymentPreflightService
}

func NewOpsDeploymentPreflightHandler(preflightService *service.OpsDeploymentPreflightService) *OpsDeploymentPreflightHandler {
	return &OpsDeploymentPreflightHandler{service: preflightService}
}

func (h *OpsDeploymentPreflightHandler) GetProjectReport(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment preflight service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid project binding id")
	if err != nil {
		return
	}

	report, err := h.service.EvaluateProject(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations project binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, report)
}

func (h *OpsDeploymentPreflightHandler) GetProjectReportByQuery(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment preflight service is not configured"))
		return
	}
	id, err := parseUintQuery(c, "project_id", "invalid project binding id")
	if err != nil {
		return
	}

	report, err := h.service.EvaluateProject(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Operations project binding")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, report)
}

func (h *OpsDeploymentPreflightHandler) GetOverview(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("operations deployment preflight service is not configured"))
		return
	}

	overview, err := h.service.EvaluateOverviewForEnvironment(strings.TrimSpace(c.Query("environment")))
	if err != nil {
		if errors.Is(err, service.ErrInvalidOpsProjectEnvironment) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, overview)
}

func parseUintQuery(c *gin.Context, name string, message string) (uint, error) {
	id, err := strconv.ParseUint(c.Query(name), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, message)
		return 0, err
	}
	return uint(id), nil
}
