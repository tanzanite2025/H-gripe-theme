package shipping

import (
	"strconv"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Tracking 相关接口 ============

// TrackShipment 追踪物流
// @Summary 追踪物流
// @Tags Shipping
// @Produce json
// @Param tracking_number path string true "追踪号"
// @Success 200 {array} shipping.TrackingEvent
// @Router /api/v1/shipping/track/{tracking_number} [get]
func (h *Handler) TrackShipment(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		apierror.RespondBadRequest(c, "tracking number is required")
		return
	}

	events, err := h.shippingRepo.FindTrackingEventsByTrackingNumber(trackingNumber)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": events})
}

// GetOrderTracking 获取订单物流追踪
// @Summary 获取订单物流追踪
// @Tags Shipping
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {array} shipping.TrackingEvent
// @Router /api/v1/shipping/orders/{order_id}/tracking [get]
func (h *Handler) GetOrderTracking(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}

	events, err := h.shippingRepo.FindTrackingEventsByOrderID(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": events})
}

// CreateTrackingEvent 创建物流追踪事件（管理员）
// @Summary 创建物流追踪事件
// @Tags Shipping
// @Accept json
// @Produce json
// @Param event body shipping.TrackingEvent true "追踪事件"
// @Success 201 {object} shipping.TrackingEvent
// @Router /api/v1/admin/shipping/tracking [post]
func (h *Handler) CreateTrackingEvent(c *gin.Context) {
	var event shipping.TrackingEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingRepo.CreateTrackingEvent(&event); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, event)
}
