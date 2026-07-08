package shipping

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Shipping Zone 閻╃鍙ч幒銉ュ經 ============

// ListZones 閼惧嘲褰囬柊宥夆偓浣稿隘閸╃喎鍨悰?// @Summary 閼惧嘲褰囬柊宥夆偓浣稿隘閸╃喎鍨悰?// @Tags Shipping
// @Produce json
// @Success 200 {array} shipping.ShippingZone
// @Router /api/v1/shipping/zones [get]
func (h *Handler) ListZones(c *gin.Context) {
	zones, err := h.shippingService.ListZones()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": zones})
}

// GetZone 閼惧嘲褰囬柊宥夆偓浣稿隘閸╃喕顕涢幆?// @Summary 閼惧嘲褰囬柊宥夆偓浣稿隘閸╃喕顕涢幆?// @Tags Shipping
// @Produce json
// @Param id path int true "閸栧搫鐓橧D"
// @Success 200 {object} shipping.ShippingZone
// @Router /api/v1/shipping/zones/{id} [get]
func (h *Handler) GetZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid zone id")
		return
	}

	zone, err := h.shippingService.GetZone(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Zone")
		return
	}

	response.Success(c, zone)
}
