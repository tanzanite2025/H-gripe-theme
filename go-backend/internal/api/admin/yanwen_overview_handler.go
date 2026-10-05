package admin

import (
	"strings"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type YanwenOverviewHandler struct {
	service *service.YanwenOperationsOverviewService
}

func NewYanwenOverviewHandler(overviewService *service.YanwenOperationsOverviewService) *YanwenOverviewHandler {
	return &YanwenOverviewHandler{service: overviewService}
}

func (h *YanwenOverviewHandler) GetYanwenOperationsOverview(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "production")))
	if environment == "test" {
		environment = "fat"
	}
	if environment != "fat" && environment != "production" {
		apierror.RespondBadRequest(c, "Yanwen environment must be fat or production")
		return
	}
	overview, err := h.service.GetYanwenOperationsOverview(environment)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, overview)
}

func registerYanwenOverviewRoutes(authenticated *gin.RouterGroup, handler *YanwenOverviewHandler) {
	group := authenticated.Group("/logistics/yanwen/overview")
	group.Use(middleware.RequirePermission(auth.PermYanwenView))
	group.GET("", handler.GetYanwenOperationsOverview)
}
