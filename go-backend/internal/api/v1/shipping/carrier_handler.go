package shipping

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCarriers(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	carriers, err := h.shippingService.ListCarriers(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": carriers})
}

func (h *Handler) GetCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	carrier, err := h.shippingService.GetCarrier(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Carrier")
		return
	}

	response.Success(c, carrier)
}
