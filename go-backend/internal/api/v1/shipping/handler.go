package shipping

import (
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"
	"errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	shippingService *service.ShippingService
	orderService    *service.OrderService
}

func NewHandler(shippingService *service.ShippingService, orderServices ...*service.OrderService) *Handler {
	var orderService *service.OrderService
	if len(orderServices) > 0 {
		orderService = orderServices[0]
	}
	return &Handler{
		shippingService: shippingService,
		orderService:    orderService,
	}
}

func (h *Handler) authorizeOrderTrackingRead(c *gin.Context, orderID uint) bool {
	if roleValue, exists := c.Get("role"); exists {
		if role, ok := roleValue.(string); ok && auth.NormalizeRole(role).HasPermission(auth.PermOrderView) {
			return true
		}
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return false
	}

	if h.orderService == nil {
		apierror.RespondInternalError(c, errors.New("order authorization service is unavailable"))
		return false
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		apierror.RespondUnauthorized(c)
		return false
	}

	if _, err := h.orderService.GetOrder(orderID, userID); err != nil {
		apierror.RespondForbidden(c)
		return false
	}

	return true
}

func (h *Handler) authorizeTrackingNumberRead(c *gin.Context, trackingNumber string) bool {
	if roleValue, exists := c.Get("role"); exists {
		if role, ok := roleValue.(string); ok && auth.NormalizeRole(role).HasPermission(auth.PermOrderView) {
			return true
		}
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return false
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		apierror.RespondUnauthorized(c)
		return false
	}

	if h.shippingService == nil || h.orderService == nil {
		apierror.RespondInternalError(c, errors.New("tracking authorization service is unavailable"))
		return false
	}

	shipment, err := h.shippingService.GetTrackingShipmentByTrackingNumber(trackingNumber)
	if err != nil {
		apierror.RespondNotFound(c, "Shipment")
		return false
	}

	if _, err := h.orderService.GetOrder(shipment.OrderID, userID); err != nil {
		apierror.RespondNotFound(c, "Shipment")
		return false
	}

	return true
}
