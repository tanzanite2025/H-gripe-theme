package admin

import (
	"strconv"
	"strings"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

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

	shipment, err := h.shippingService.GetTrackingShipmentByOrderID(uint(orderID))
	if err != nil {
		apierror.RespondNotFound(c, "Tracking shipment")
		return
	}

	events, err := h.shippingService.GetTrackingEventsByOrderID(shipment.OrderID)
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

	shipment, err := h.shippingService.GetTrackingShipmentByOrderID(uint(orderID))
	if err != nil {
		apierror.RespondNotFound(c, "Tracking shipment")
		return
	}

	if err := h.shippingService.RegisterTrackingShipment(c.Request.Context(), service.TrackingSyncInput{
		OrderID:                  shipment.OrderID,
		ProviderID:               shipment.TrackingProviderID,
		TrackingNumber:           shipment.TrackingNumber,
		ProviderCarrierCode:      shipment.ProviderCarrierCode,
		CarrierID:                shipment.CarrierID,
		CarrierServiceID:         shipment.CarrierServiceID,
		TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
	}); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	updatedShipment, err := h.shippingService.GetTrackingShipmentByOrderID(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"shipment": updatedShipment})
}

func (h *ShippingHandler) SyncTrackingShipment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderID"), 10, 32)
	if err != nil || orderID == 0 {
		apierror.RespondBadRequest(c, "invalid order ID")
		return
	}

	shipment, err := h.shippingService.GetTrackingShipmentByOrderID(uint(orderID))
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
