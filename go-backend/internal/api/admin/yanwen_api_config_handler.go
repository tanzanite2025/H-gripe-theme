package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// Yanwen gateway configuration and self-check admin endpoints.

type YanwenAPIConfigHandler struct {
	service *service.YanwenGatewayConfigurationService
}

func NewYanwenAPIConfigHandler(configurationService *service.YanwenGatewayConfigurationService) *YanwenAPIConfigHandler {
	return &YanwenAPIConfigHandler{service: configurationService}
}

func (h *YanwenAPIConfigHandler) GetYanwenAPIConfiguration(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "fat")))
	if environment == "test" {
		environment = "fat"
	}
	if environment != "fat" && environment != "production" {
		apierror.RespondBadRequest(c, "Yanwen environment must be fat or production")
		return
	}
	view, err := h.service.GetYanwenAPIConfigurationView(environment)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, view)
}

func (h *YanwenAPIConfigHandler) SaveYanwenAPIConfiguration(c *gin.Context) {
	var input service.YanwenAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := h.service.SaveYanwenAPIConfiguration(input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "Yanwen API configuration saved", map[string]bool{"saved": true})
}

func (h *YanwenAPIConfigHandler) PingYanwenGateway(c *gin.Context) {
	var input service.YanwenAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.PingYanwenGateway(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func registerYanwenAPIConfigRoutes(authenticated *gin.RouterGroup, handler *YanwenAPIConfigHandler) {
	group := authenticated.Group("/logistics/yanwen/api")
	group.Use(middleware.RequirePermission(auth.PermYanwenView))
	group.GET("/config", handler.GetYanwenAPIConfiguration)
	group.PUT("/config", middleware.RequirePermission(auth.PermYanwenManage), handler.SaveYanwenAPIConfiguration)
	group.POST("/ping", middleware.RequirePermission(auth.PermYanwenManage), handler.PingYanwenGateway)
}
