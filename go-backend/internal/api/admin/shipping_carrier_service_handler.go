package admin

import (
	"errors"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	carrierServiceBillingActualWeight = "actual_weight"
	carrierServiceBillingVolumetric   = "volumetric_weight"
	carrierServiceBillingGreaterOf    = "greater_of_actual_and_volumetric"
)

func (h *ShippingHandler) ListCarrierServices(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	services, err := h.shippingService.ListCarrierServices(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": services})
}

func (h *ShippingHandler) GetCarrierService(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid carrier service id")
	if err != nil {
		return
	}

	service, err := h.shippingService.GetCarrierService(id)
	if err != nil {
		apierror.RespondNotFound(c, "Carrier service")
		return
	}

	response.Success(c, service)
}

func (h *ShippingHandler) CreateCarrierService(c *gin.Context) {
	var req shippingCarrierServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	service := req.toDomain()
	if err := validateCarrierService(service); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateCarrierService(&service); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, service)
}

func (h *ShippingHandler) UpdateCarrierService(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid carrier service id")
	if err != nil {
		return
	}

	var req shippingCarrierServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	service := req.toDomain()
	service.ID = id
	if err := validateCarrierService(service); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateCarrierService(&service); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, service)
}

func (h *ShippingHandler) DeleteCarrierService(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid carrier service id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteCarrierService(id); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "carrier service deleted", nil)
}

func validateCarrierService(service shippingdomain.CarrierService) error {
	if service.CarrierID == 0 {
		return errors.New("carrier id is required")
	}
	if service.ServiceCode == "" {
		return errors.New("service code is required")
	}
	if service.ServiceName == "" {
		return errors.New("service name is required")
	}

	switch service.BillingMode {
	case carrierServiceBillingActualWeight, carrierServiceBillingVolumetric, carrierServiceBillingGreaterOf:
	default:
		return errors.New("billing mode must be actual_weight, volumetric_weight or greater_of_actual_and_volumetric")
	}

	if service.FirstWeightGrams < 0 ||
		service.AdditionalWeightGrams < 0 ||
		service.MinChargeWeightGrams < 0 ||
		service.VolumetricDivisor < 0 ||
		service.FuelSurchargePercent < 0 ||
		service.RemoteSurcharge < 0 ||
		service.EtaMinDays < 0 ||
		service.EtaMaxDays < 0 {
		return errors.New("carrier service numeric fields cannot be negative")
	}
	if service.VolumetricDivisor == 0 && service.BillingMode != carrierServiceBillingActualWeight {
		return errors.New("volumetric divisor is required for volumetric billing")
	}
	if service.EtaMaxDays > 0 && service.EtaMinDays > 0 && service.EtaMaxDays < service.EtaMinDays {
		return errors.New("eta max days cannot be less than eta min days")
	}

	return nil
}
