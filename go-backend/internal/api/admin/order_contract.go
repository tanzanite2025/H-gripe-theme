package admin

import (
	"commerce-platform/internal/service"
	"errors"
	"net/http"

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

type orderFulfillmentRequest struct {
	TrackingNumber     string `json:"tracking_number" binding:"required"`
	TrackingProviderID uint   `json:"tracking_provider_id" binding:"required"`
	CarrierID          *uint  `json:"carrier_id"`
	CarrierServiceID   *uint  `json:"carrier_service_id"`
}

type adminNoteRequest struct {
	AdminNote string `json:"admin_note"`
}

type orderItemCustomsRequest struct {
	DeclaredValue          *float64 `json:"declared_value"`
	DeclaredValueConfirmed bool     `json:"declared_value_confirmed"`
}

type orderDisputeContactEmailRequest struct {
	Provider  string `json:"provider" binding:"required"`
	DisputeID uint   `json:"dispute_id" binding:"required"`
	Subject   string `json:"subject" binding:"required"`
	Body      string `json:"body" binding:"required"`
	Confirm   bool   `json:"confirm"`
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

func (r orderFulfillmentRequest) toServiceInput() service.OrderTrackingUpdateInput {
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
	case errors.Is(err, service.ErrOrderDisputePaymentNotConfigured):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order dispute analysis is not configured"})
	case errors.Is(err, service.ErrOrderDisputeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order dispute not found"})
	case errors.Is(err, service.ErrOrderDisputeEmailConfirmRequired),
		errors.Is(err, service.ErrOrderDisputeEmailRecipientMissing),
		errors.Is(err, service.ErrOrderDisputeEmailSubjectRequired),
		errors.Is(err, service.ErrOrderDisputeEmailBodyRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderDisputeEmailNotConfigured):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order dispute contact email is not configured"})
	case errors.Is(err, service.ErrSystemManagedOrderStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderFulfillmentNotAllowed),
		errors.Is(err, service.ErrOrderFulfillmentPaymentRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderDeleteNotAllowed):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
	case errors.Is(err, service.ErrDeclaredValueInvalid),
		errors.Is(err, service.ErrDeclaredValueConfirmationRequired):
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
