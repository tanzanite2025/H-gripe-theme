package admin

import (
	"errors"
	"strings"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ShippingHandler) ListZones(c *gin.Context) {
	zones, err := h.shippingService.ListZones()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": zones})
}

func (h *ShippingHandler) GetZone(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid zone id")
	if err != nil {
		return
	}

	zone, err := h.shippingService.GetZone(id)
	if err != nil {
		apierror.RespondNotFound(c, "Shipping zone")
		return
	}

	response.Success(c, zone)
}

func (h *ShippingHandler) CreateZone(c *gin.Context) {
	var req shippingZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	zone := req.toDomain()
	if err := validateShippingZone(zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateZone(&zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, zone)
}

func (h *ShippingHandler) UpdateZone(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid zone id")
	if err != nil {
		return
	}

	var req shippingZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	zone := req.toDomain()
	zone.ID = id
	if err := validateShippingZone(zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateZone(&zone); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, zone)
}

func (h *ShippingHandler) DeleteZone(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid zone id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteZone(id); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "shipping zone deleted", nil)
}

func validateShippingZone(zone shippingdomain.ShippingZone) error {
	if strings.TrimSpace(zone.Name) == "" {
		return errors.New("zone name is required")
	}
	if strings.TrimSpace(zone.Countries) == "" {
		return errors.New("zone countries are required")
	}
	return nil
}
