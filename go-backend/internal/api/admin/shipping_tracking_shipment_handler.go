package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *ShippingHandler) ListTrackingShipments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	shipments, err := h.shippingService.ListTrackingShipments(service.TrackingShipmentListFilter{
		SyncStatus:          c.Query("sync_status"),
		RegistrationStatus:  c.Query("registration_status"),
		TrackingNumber:      c.Query("tracking_number"),
		ProviderCarrierCode: c.Query("provider_carrier_code"),
		Keyword:             c.Query("keyword"),
		OrderID:             queryUint(c, "order_id"),
		ProviderID:          queryUint(c, "provider_id"),
		CarrierID:           queryUint(c, "carrier_id"),
		CarrierServiceID:    queryUint(c, "carrier_service_id"),
		Enabled:             queryBoolPtr(c, "enabled"),
		DueOnly:             strings.EqualFold(c.Query("due_only"), "true"),
		Limit:               limit,
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": shipments})
}

func (h *ShippingHandler) GetTrackingPollingState(c *gin.Context) {
	response.Success(c, h.shippingService.TrackingPollingState())
}

func (h *ShippingHandler) GetTrackingWebhookState(c *gin.Context) {
	response.Success(c, h.shippingService.TrackingWebhookState())
}

func (h *ShippingHandler) ListTrackingEvents(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderID"), 10, 32)
	if err != nil || orderID == 0 {
		apierror.RespondBadRequest(c, "invalid order ID")
		return
	}
	trackingNumber := strings.TrimSpace(c.Query("tracking_number"))
	if trackingNumber == "" {
		apierror.RespondBadRequest(c, "tracking_number is required")
		return
	}

	shipment, err := h.shippingService.GetTrackingShipmentByOrderIDAndTrackingNumber(uint(orderID), trackingNumber)
	if err != nil {
		apierror.RespondNotFound(c, "Tracking shipment")
		return
	}

	events, err := h.shippingService.GetTrackingEventsByOrderIDAndTrackingNumber(shipment.OrderID, shipment.TrackingNumber)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": events})
}

func queryUint(c *gin.Context, key string) uint {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Query(key)), 10, 32)
	if err != nil {
		return 0
	}
	return uint(value)
}

func queryBoolPtr(c *gin.Context, key string) *bool {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" || strings.EqualFold(raw, "all") {
		return nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &value
}

func (h *ShippingHandler) SyncDueTrackingShipments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.shippingService.SyncDueTrackingShipments(c.Request.Context(), limit)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ShippingHandler) RegisterTrackingShipment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderID"), 10, 32)
	if err != nil || orderID == 0 {
		apierror.RespondBadRequest(c, "invalid order ID")
		return
	}

	trackingNumber := strings.TrimSpace(c.Query("tracking_number"))
	if trackingNumber == "" {
		apierror.RespondBadRequest(c, "tracking_number is required")
		return
	}
	shipment, err := h.shippingService.RequestTrackingShipmentRegistration(c.Request.Context(), uint(orderID), trackingNumber)
	if err != nil {
		if service.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Tracking shipment")
			return
		}
		if errors.Is(err, resilience.ErrExternalOutcomeUnknown) {
			apierror.RespondConflict(c, "Tracking registration requires reconciliation before another attempt")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Accepted(c, gin.H{
		"message":  "Tracking registration has been queued",
		"shipment": shipment,
	})
}

func (h *ShippingHandler) SyncTrackingShipment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderID"), 10, 32)
	if err != nil || orderID == 0 {
		apierror.RespondBadRequest(c, "invalid order ID")
		return
	}

	trackingNumber := strings.TrimSpace(c.Query("tracking_number"))
	if trackingNumber == "" {
		apierror.RespondBadRequest(c, "tracking_number is required")
		return
	}
	shipment, err := h.shippingService.GetTrackingShipmentByOrderIDAndTrackingNumber(uint(orderID), trackingNumber)
	if err != nil {
		apierror.RespondNotFound(c, "Tracking shipment")
		return
	}

	result, err := h.shippingService.SyncTracking(c.Request.Context(), service.TrackingSyncInput{
		OrderID:                  shipment.OrderID,
		ProviderID:               shipment.TrackingProviderID,
		TrackingNumber:           shipment.TrackingNumber,
		ProviderCarrierCode:      shipment.ProviderCarrierCode,
		CarrierID:                shipment.CarrierID,
		CarrierServiceID:         shipment.CarrierServiceID,
		TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
	})
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"tracking": result})
}
