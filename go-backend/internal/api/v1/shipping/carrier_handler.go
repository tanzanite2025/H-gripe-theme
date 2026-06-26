package shipping

import (
	"strconv"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Carrier 相关接口 ============

// ListCarriers 获取物流公司列表
// @Summary 获取物流公司列表
// @Tags Shipping
// @Produce json
// @Param enabled query bool false "只显示启用的"
// @Success 200 {array} shipping.Carrier
// @Router /api/v1/shipping/carriers [get]
func (h *Handler) ListCarriers(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	carriers, err := h.shippingRepo.FindAllCarriers(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": carriers})
}

// GetCarrier 获取物流公司详情
// @Summary 获取物流公司详情
// @Tags Shipping
// @Produce json
// @Param id path int true "物流公司ID"
// @Success 200 {object} shipping.Carrier
// @Router /api/v1/shipping/carriers/{id} [get]
func (h *Handler) GetCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	carrier, err := h.shippingRepo.FindCarrierByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Carrier")
		return
	}

	response.Success(c, carrier)
}

// CreateCarrier 创建物流公司（管理员）
// @Summary 创建物流公司
// @Tags Shipping
// @Accept json
// @Produce json
// @Param carrier body shipping.Carrier true "物流公司信息"
// @Success 201 {object} shipping.Carrier
// @Router /api/v1/admin/shipping/carriers [post]
func (h *Handler) CreateCarrier(c *gin.Context) {
	var carrier shipping.Carrier
	if err := c.ShouldBindJSON(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingRepo.CreateCarrier(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, carrier)
}

// UpdateCarrier 更新物流公司（管理员）
// @Summary 更新物流公司
// @Tags Shipping
// @Accept json
// @Produce json
// @Param id path int true "物流公司ID"
// @Param carrier body shipping.Carrier true "物流公司信息"
// @Success 200 {object} shipping.Carrier
// @Router /api/v1/admin/shipping/carriers/{id} [put]
func (h *Handler) UpdateCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	var carrier shipping.Carrier
	if err := c.ShouldBindJSON(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	carrier.ID = uint(id)
	if err := h.shippingRepo.UpdateCarrier(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, carrier)
}

// DeleteCarrier 删除物流公司（管理员）
// @Summary 删除物流公司
// @Tags Shipping
// @Produce json
// @Param id path int true "物流公司ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/shipping/carriers/{id} [delete]
func (h *Handler) DeleteCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	if err := h.shippingRepo.DeleteCarrier(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "carrier deleted", nil)
}
