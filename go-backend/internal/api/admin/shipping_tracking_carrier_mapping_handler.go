package admin

import (
	"errors"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	trackingMappingScopeCarrier        = "carrier"
	trackingMappingScopeCarrierService = "carrier_service"
)

func (h *ShippingHandler) ListTrackingCarrierMappings(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	mappings, err := h.shippingService.ListTrackingCarrierMappings(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": mappings})
}

func (h *ShippingHandler) GetTrackingCarrierMapping(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid tracking carrier mapping id")
	if err != nil {
		return
	}

	mapping, err := h.shippingService.GetTrackingCarrierMapping(id)
	if err != nil {
		apierror.RespondNotFound(c, "Tracking carrier mapping")
		return
	}

	response.Success(c, mapping)
}

func (h *ShippingHandler) CreateTrackingCarrierMapping(c *gin.Context) {
	var req shippingTrackingCarrierMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	mapping := req.toDomain()
	if err := validateTrackingCarrierMapping(&mapping); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateTrackingCarrierMapping(&mapping); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, mapping)
}

func (h *ShippingHandler) UpdateTrackingCarrierMapping(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid tracking carrier mapping id")
	if err != nil {
		return
	}

	var req shippingTrackingCarrierMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	mapping := req.toDomain()
	mapping.ID = id
	if err := validateTrackingCarrierMapping(&mapping); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateTrackingCarrierMapping(&mapping); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, mapping)
}

func (h *ShippingHandler) DeleteTrackingCarrierMapping(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid tracking carrier mapping id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteTrackingCarrierMapping(id); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "tracking carrier mapping deleted", nil)
}

func validateTrackingCarrierMapping(mapping *shippingdomain.TrackingCarrierMapping) error {
	if mapping.ProviderID == 0 {
		return errors.New("tracking provider id is required")
	}
	if mapping.ProviderCarrierCode == "" {
		return errors.New("provider carrier code is required")
	}
	if mapping.Priority < 0 {
		return errors.New("tracking carrier mapping priority cannot be negative")
	}

	switch mapping.Scope {
	case trackingMappingScopeCarrier:
		if mapping.CarrierID == nil || *mapping.CarrierID == 0 {
			return errors.New("carrier id is required for carrier scope mapping")
		}
		mapping.CarrierServiceID = nil
	case trackingMappingScopeCarrierService:
		if mapping.CarrierServiceID == nil || *mapping.CarrierServiceID == 0 {
			return errors.New("carrier service id is required for carrier service scope mapping")
		}
		mapping.CarrierID = nil
	default:
		return errors.New("tracking carrier mapping scope must be carrier or carrier_service")
	}

	return nil
}
