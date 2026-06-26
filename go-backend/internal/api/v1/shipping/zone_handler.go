package shipping

import (
	"strconv"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Shipping Zone 相关接口 ============

// ListZones 获取配送区域列表
// @Summary 获取配送区域列表
// @Tags Shipping
// @Produce json
// @Success 200 {array} shipping.ShippingZone
// @Router /api/v1/shipping/zones [get]
func (h *Handler) ListZones(c *gin.Context) {
	zones, err := h.shippingRepo.FindAllZones()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": zones})
}

// GetZone 获取配送区域详情
// @Summary 获取配送区域详情
// @Tags Shipping
// @Produce json
// @Param id path int true "区域ID"
// @Success 200 {object} shipping.ShippingZone
// @Router /api/v1/shipping/zones/{id} [get]
func (h *Handler) GetZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid zone id")
		return
	}

	zone, err := h.shippingRepo.FindZoneByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Zone")
		return
	}

	response.Success(c, zone)
}

// CreateZone 创建配送区域（管理员）
// @Summary 创建配送区域
// @Tags Shipping
// @Accept json
// @Produce json
// @Param zone body shipping.ShippingZone true "区域信息"
// @Success 201 {object} shipping.ShippingZone
// @Router /api/v1/admin/shipping/zones [post]
func (h *Handler) CreateZone(c *gin.Context) {
	var zone shipping.ShippingZone
	if err := c.ShouldBindJSON(&zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingRepo.CreateZone(&zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, zone)
}

// UpdateZone 更新配送区域（管理员）
// @Summary 更新配送区域
// @Tags Shipping
// @Accept json
// @Produce json
// @Param id path int true "区域ID"
// @Param zone body shipping.ShippingZone true "区域信息"
// @Success 200 {object} shipping.ShippingZone
// @Router /api/v1/admin/shipping/zones/{id} [put]
func (h *Handler) UpdateZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid zone id")
		return
	}

	var zone shipping.ShippingZone
	if err := c.ShouldBindJSON(&zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	zone.ID = uint(id)
	if err := h.shippingRepo.UpdateZone(&zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, zone)
}

// DeleteZone 删除配送区域（管理员）
// @Summary 删除配送区域
// @Tags Shipping
// @Produce json
// @Param id path int true "区域ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/shipping/zones/{id} [delete]
func (h *Handler) DeleteZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid zone id")
		return
	}

	if err := h.shippingRepo.DeleteZone(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "zone deleted", nil)
}
