package admin

import (
	"time"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// YanwenTrackingAlertHandler exposes only the read-only Yanwen tracking alert
// projection. It does not create customer-service cases or update shipment
// state.
type YanwenTrackingAlertHandler struct {
	service *service.YanwenTrackingAlertService
}

func NewYanwenTrackingAlertHandler(alertService *service.YanwenTrackingAlertService) *YanwenTrackingAlertHandler {
	return &YanwenTrackingAlertHandler{service: alertService}
}

func (h *YanwenTrackingAlertHandler) ListProductionTrackingAlerts(c *gin.Context) {
	alerts, err := h.service.ListProductionYanwenTrackingAlerts(time.Now().UTC())
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": alerts})
}

func registerYanwenTrackingAlertRoutes(
	authenticated *gin.RouterGroup,
	handler *YanwenTrackingAlertHandler,
) {
	group := authenticated.Group("/logistics/yanwen/tracking/alerts")
	group.Use(middleware.RequirePermission(auth.PermYanwenView))
	group.GET("", handler.ListProductionTrackingAlerts)
}
