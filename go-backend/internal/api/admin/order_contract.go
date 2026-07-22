package admin

import (
	"errors"
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type orderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing shipped completed cancelled"`
}

type shippingStatusRequest struct {
	ShippingStatus string `json:"shipping_status" binding:"required,oneof=pending processing shipped delivered"`
}

type trackingInfoRequest struct {
	TrackingNumber     string `json:"tracking_number" binding:"required"`
	TrackingProviderID uint   `json:"tracking_provider_id" binding:"required"`
	CarrierID          *uint  `json:"carrier_id"`
	CarrierServiceID   *uint  `json:"carrier_service_id"`
}

type adminNoteRequest struct {
	AdminNote string `json:"admin_note"`
}

type orderBatchStatusRequest struct {
	OrderIDs []uint `json:"order_ids" binding:"required,min=1"`
	Status   string `json:"status" binding:"required,oneof=pending processing shipped completed cancelled"`
}

func (r trackingInfoRequest) toServiceInput() service.OrderTrackingUpdateInput {
	return service.OrderTrackingUpdateInput{
		TrackingNumber:     r.TrackingNumber,
		TrackingProviderID: r.TrackingProviderID,
		CarrierID:          r.CarrierID,
		CarrierServiceID:   r.CarrierServiceID,
	}
}

func respondOrderServiceError(c *gin.Context, err error, fallbackMessage string, defaultStatus int) {
	switch {
	case errors.Is(err, service.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
	case errors.Is(err, service.ErrOrderDeleteNotAllowed):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrTrackingNumberRequired),
		errors.Is(err, service.ErrTrackingProviderRequired),
		errors.Is(err, service.ErrTrackingLocalTargetRequired),
		errors.Is(err, service.ErrTrackingProviderDisabled),
		errors.Is(err, service.ErrTrackingCarrierDisabled),
		errors.Is(err, service.ErrTrackingCarrierServiceDisabled),
		errors.Is(err, service.ErrTrackingCarrierMappingMissing):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrTrackingOrderRequired),
		errors.Is(err, service.ErrTrackingCarrierCodeRequired),
		errors.Is(err, service.ErrTrackingProviderAPIKeyMissing),
		errors.Is(err, service.ErrTrackingProviderBaseURLMissing),
		errors.Is(err, service.ErrTrackingProviderUnsupported):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(defaultStatus, gin.H{"error": fallbackMessage})
	}
}
