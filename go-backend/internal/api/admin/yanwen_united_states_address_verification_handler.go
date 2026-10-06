package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type YanwenUnitedStatesAddressVerificationHandler struct {
	service *service.YanwenUnitedStatesAddressVerificationService
}

func NewYanwenUnitedStatesAddressVerificationHandler(
	addressVerificationService *service.YanwenUnitedStatesAddressVerificationService,
) *YanwenUnitedStatesAddressVerificationHandler {
	return &YanwenUnitedStatesAddressVerificationHandler{service: addressVerificationService}
}

func (h *YanwenUnitedStatesAddressVerificationHandler) VerifyUnitedStatesAddress(c *gin.Context) {
	var input service.YanwenUnitedStatesAddressVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.VerifyUnitedStatesAddress(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func registerYanwenUnitedStatesAddressVerificationRoutes(
	authenticated *gin.RouterGroup,
	handler *YanwenUnitedStatesAddressVerificationHandler,
) {
	group := authenticated.Group("/logistics/yanwen/customs")
	group.Use(middleware.RequirePermission(auth.PermYanwenShip))
	group.POST("/united-states/address", handler.VerifyUnitedStatesAddress)
}
