package admin

import (
	"errors"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsOverviewHandler struct {
	overviewService *service.OpsOverviewService
}

func NewOpsOverviewHandler(overviewService *service.OpsOverviewService) *OpsOverviewHandler {
	return &OpsOverviewHandler{overviewService: overviewService}
}

func (h *OpsOverviewHandler) Get(c *gin.Context) {
	if h == nil || h.overviewService == nil {
		apierror.RespondInternalError(c, errors.New("operations overview service is not configured"))
		return
	}
	overview, err := h.overviewService.GetForEnvironment(strings.TrimSpace(c.Query("environment")))
	if err != nil {
		if errors.Is(err, service.ErrInvalidOpsOverviewEnvironment) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, overview)
}
