package shipping

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCarrierServices(c *gin.Context) {
	services, err := h.shippingService.ListPublicCarrierServices()
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

	service, err := h.shippingService.GetPublicCarrierService(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Carrier service")
		return
	}

	response.Success(c, service)
}
