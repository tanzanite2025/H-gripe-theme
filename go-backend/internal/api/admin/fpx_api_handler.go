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

type FpxAPIHandler struct{ service *service.FpxAPIService }

func NewFpxAPIHandler(s *service.FpxAPIService) *FpxAPIHandler { return &FpxAPIHandler{service: s} }
func (h *FpxAPIHandler) Get(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "production")))
	if environment != "test" && environment != "production" {
		apierror.RespondBadRequest(c, "4PX environment must be test or production")
		return
	}
	view, err := h.service.View(environment)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, view)
}
func (h *FpxAPIHandler) Put(c *gin.Context) {
	var input service.FpxAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := h.service.Save(input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "4PX API configuration saved", nil)
}
func (h *FpxAPIHandler) Ping(c *gin.Context) {
	var input service.FpxAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.Ping(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}
func (h *FpxAPIHandler) SyncChannels(c *gin.Context) {
	var input service.FpxAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	summary, err := h.service.SyncChannels(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, summary)
}
func registerFpxAPIRoutes(authenticated *gin.RouterGroup, h *FpxAPIHandler) {
	group := authenticated.Group("/logistics/4px/api")
	group.Use(middleware.RequirePermission(auth.PermFPXView))
	group.GET("/config", h.Get)
	group.PUT("/config", middleware.RequirePermission(auth.PermFPXManage), h.Put)
	group.POST("/ping", middleware.RequirePermission(auth.PermFPXManage), h.Ping)
	group.POST("/sync-channels", middleware.RequirePermission(auth.PermFPXManage), h.SyncChannels)
}
