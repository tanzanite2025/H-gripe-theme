package shipping

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCarrierServices(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	services, err := h.shippingService.ListCarrierServices(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": services})
}

func (h *Handler) GetCarrierService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier service id")
		return
	}

	service, err := h.shippingService.GetCarrierService(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Carrier service")
		return
	}

	response.Success(c, service)
}
