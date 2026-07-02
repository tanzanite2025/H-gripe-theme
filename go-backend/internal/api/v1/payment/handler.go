package payment

import (
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	paymentService *service.PaymentService
	orderService   *service.OrderService
}

func NewHandler(paymentService *service.PaymentService, orderService *service.OrderService) *Handler {
	return &Handler{
		paymentService: paymentService,
		orderService:   orderService,
	}
}

func (h *Handler) authorizeOrderPaymentRead(c *gin.Context, orderID uint) bool {
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

	if _, err := h.orderService.GetOrder(orderID, userIDValue.(uint)); err != nil {
		apierror.RespondForbidden(c)
		return false
	}

	return true
}
