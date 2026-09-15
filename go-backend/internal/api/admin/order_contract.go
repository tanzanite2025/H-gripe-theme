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
	SignatureConfirmed bool   `json:"signature_confirmed"`
}

type orderProductionRequest struct {
	Confirm bool `json:"confirm"`
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
		SignatureConfirmed: r.SignatureConfirmed,
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
	case errors.Is(err, service.ErrPaidOrderCancellationNotAllowed),
		errors.Is(err, service.ErrProductionStartedCancellationNotAllowed),
		errors.Is(err, service.ErrOrderCancellationConflict),
		errors.Is(err, service.ErrOrderStatusConflict),
		errors.Is(err, service.ErrOrderFulfillmentOnHold):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSystemManagedOrderStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderFulfillmentStatusManaged):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderCustomsUpdateLocked):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderCustomsUpdateTransactionNeeded):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order customs update transaction is not configured"})
	case errors.Is(err, service.ErrOrderFulfillmentNotAllowed),
		errors.Is(err, service.ErrOrderFulfillmentPaymentRequired),
		errors.Is(err, service.ErrOrderFulfillmentSignatureConfirmationRequired),
		errors.Is(err, service.ErrOrderFulfillmentEvidenceNotConfigured),
		errors.Is(err, service.ErrOrderFulfillmentEvidencePackageMissing),
		errors.Is(err, service.ErrOrderProductionNotCompleted),
		errors.Is(err, service.ErrOrderProductionNotRequired),
		errors.Is(err, service.ErrOrderProductionPaymentRequired),
		errors.Is(err, service.ErrOrderProductionNotAllowed),
		errors.Is(err, service.ErrOrderProductionNotStarted):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderFulfillmentTrackingConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderFulfillmentEvidenceIncomplete):
		var evidenceErr *service.OrderFulfillmentEvidenceError
		if errors.As(err, &evidenceErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    evidenceErr.Error(),
				"code":     "order_fulfillment_evidence_incomplete",
				"evidence": evidenceErr.Check,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "order_fulfillment_evidence_incomplete"})
	case errors.Is(err, service.ErrOrderProductionAlreadyStarted):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderHideTransactionRequired):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order hiding transaction is not configured"})
	case errors.Is(err, service.ErrOrderHideNotAllowed):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "order_hide_not_allowed",
		})
	case errors.Is(err, service.ErrOrderItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
	case errors.Is(err, service.ErrDeclaredValueInvalid),
		errors.Is(err, service.ErrDeclaredValueConfirmationRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrOrderCustomsDeclarationIncomplete):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "order_customs_declared_value_incomplete",
		})
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
