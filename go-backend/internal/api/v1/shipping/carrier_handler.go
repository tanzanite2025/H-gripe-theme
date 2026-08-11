package shipping

import (
	"strconv"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCarriers(c *gin.Context) {
	carriers, err := h.shippingService.ListPublicCarriers()
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

	carrier, err := h.shippingService.GetPublicCarrier(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Carrier")
		return
	}

	response.Success(c, carrier)
}
