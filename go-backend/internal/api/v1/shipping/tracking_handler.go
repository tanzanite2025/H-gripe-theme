package shipping

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Tracking 閻╃鍙ч幒銉ュ經 ============

// TrackShipment 鏉╁€熼嚋閻椻晜绁?// @Summary 鏉╁€熼嚋閻椻晜绁?// @Tags Shipping
// @Produce json
// @Param tracking_number path string true "鏉╁€熼嚋閸?
// @Success 200 {array} shipping.TrackingEvent
// @Router /api/v1/shipping/track/{tracking_number} [get]
func (h *Handler) TrackShipment(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		apierror.RespondBadRequest(c, "tracking number is required")
		return
	}

	events, err := h.shippingService.GetTrackingEventsByTrackingNumber(trackingNumber)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": events})
}

// GetOrderTracking 閼惧嘲褰囩拋銏犲礋閻椻晜绁︽潻鍊熼嚋
// @Summary 閼惧嘲褰囩拋銏犲礋閻椻晜绁︽潻鍊熼嚋
// @Tags Shipping
// @Produce json
// @Param order_id path int true "鐠併垹宕烮D"
// @Success 200 {array} shipping.TrackingEvent
// @Router /api/v1/shipping/orders/{order_id}/tracking [get]
func (h *Handler) GetOrderTracking(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}

	events, err := h.shippingService.GetTrackingEventsByOrderID(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": events})
}
