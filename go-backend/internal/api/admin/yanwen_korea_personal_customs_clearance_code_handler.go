package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type YanwenKoreaPersonalCustomsClearanceCodeHandler struct {
	service *service.YanwenKoreaPersonalCustomsClearanceCodeService
}

func NewYanwenKoreaPersonalCustomsClearanceCodeHandler(
	customsService *service.YanwenKoreaPersonalCustomsClearanceCodeService,
) *YanwenKoreaPersonalCustomsClearanceCodeHandler {
	return &YanwenKoreaPersonalCustomsClearanceCodeHandler{service: customsService}
}

func (h *YanwenKoreaPersonalCustomsClearanceCodeHandler) VerifyKoreaPersonalCustomsClearanceCode(c *gin.Context) {
	var input service.YanwenKoreaPersonalCustomsClearanceCodeVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.VerifyKoreaPersonalCustomsClearanceCode(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func registerYanwenKoreaPersonalCustomsClearanceCodeRoutes(
	authenticated *gin.RouterGroup,
	handler *YanwenKoreaPersonalCustomsClearanceCodeHandler,
) {
	group := authenticated.Group("/logistics/yanwen/customs")
	group.Use(middleware.RequirePermission(auth.PermYanwenShip))
	group.POST("/korea/pccc", handler.VerifyKoreaPersonalCustomsClearanceCode)
}
