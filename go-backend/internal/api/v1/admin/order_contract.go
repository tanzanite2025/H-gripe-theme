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
	TrackingNumber string `json:"tracking_number" binding:"required"`
	CarrierCode    string `json:"carrier_code"`
}

type adminNoteRequest struct {
	AdminNote string `json:"admin_note"`
}

type orderBatchStatusRequest struct {
	OrderIDs []uint `json:"order_ids" binding:"required,min=1"`
	Status   string `json:"status" binding:"required,oneof=pending processing shipped completed cancelled"`
}

func respondOrderServiceError(c *gin.Context, err error, fallbackMessage string, defaultStatus int) {
	switch {
	case errors.Is(err, service.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
	case errors.Is(err, service.ErrOrderDeleteNotAllowed):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(defaultStatus, gin.H{"error": fallbackMessage})
	}
}
