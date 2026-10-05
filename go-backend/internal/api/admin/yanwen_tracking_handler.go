package admin

import (
	"strings"
	"unicode"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type YanwenTrackingHandler struct {
	service *service.YanwenTrackingOperationsService
}

func NewYanwenTrackingHandler(trackingService *service.YanwenTrackingOperationsService) *YanwenTrackingHandler {
	return &YanwenTrackingHandler{service: trackingService}
}

// QueryTracking proxies the official production-only Yanwen tracking endpoint
// through the backend so merchant credentials never reach the browser.
func (h *YanwenTrackingHandler) QueryYanwenTracking(c *gin.Context) {
	trackingNumbers := strings.FieldsFunc(c.Query("nums"), func(r rune) bool {
		return r == ',' || r == '，' || unicode.IsSpace(r)
	})
	results, err := h.service.QueryYanwenTracking(c.Request.Context(), trackingNumbers)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"data": results})
}

func registerYanwenTrackingRoutes(authenticated *gin.RouterGroup, handler *YanwenTrackingHandler) {
	group := authenticated.Group("/logistics/yanwen/tracking")
	group.Use(middleware.RequirePermission(auth.PermYanwenView))
	group.GET("", handler.QueryYanwenTracking)
}
