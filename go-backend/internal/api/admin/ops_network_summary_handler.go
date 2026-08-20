package admin

import (
	"errors"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OpsNetworkSummaryHandler struct {
	networkSummaryService *service.OpsNetworkSummaryService
}

func NewOpsNetworkSummaryHandler(networkSummaryService *service.OpsNetworkSummaryService) *OpsNetworkSummaryHandler {
	return &OpsNetworkSummaryHandler{networkSummaryService: networkSummaryService}
}

func (h *OpsNetworkSummaryHandler) Get(c *gin.Context) {
	if h == nil || h.networkSummaryService == nil {
		apierror.RespondInternalError(c, errors.New("operations network summary service is not configured"))
		return
	}
	summary, err := h.networkSummaryService.Get(strings.TrimSpace(c.Query("environment")))
	if err != nil {
		if errors.Is(err, service.ErrInvalidOpsNetworkEnvironment) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, summary)
}
