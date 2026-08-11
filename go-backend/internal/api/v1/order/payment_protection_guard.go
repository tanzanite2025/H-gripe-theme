package order

import (
	"net/http"
	"strings"

	"commerce-platform/internal/api/middleware"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
)

func (h *Handler) authorizeOrderPaymentStart(c *gin.Context, req CreateOrderRequest) bool {
	if h == nil || h.paymentProtection == nil {
		return true
	}

	paymentMethod := strings.ToLower(strings.TrimSpace(req.PaymentMethod))
	decision, err := h.paymentProtection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      pgateway.ProviderForPaymentMethod(paymentMethod),
		Country:       orderPaymentProtectionCountry(c, req),
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		apierror.RespondError(c, http.StatusServiceUnavailable, "payment_protection_unavailable", "Payment protection is temporarily unavailable")
		return false
	}
	if !decision.PausePayment {
		return true
	}

	apierror.RespondErrorWithDetails(c, http.StatusForbidden, "payment_paused", "Payment is temporarily unavailable for this order", gin.H{
		"action": "pause_payment",
	})
	return false
}

func orderPaymentProtectionCountry(c *gin.Context, req CreateOrderRequest) string {
	if country := strings.TrimSpace(req.BillingAddress.Country); country != "" {
		return country
	}
	if country := strings.TrimSpace(req.ShippingAddress.Country); country != "" {
		return country
	}
	return middleware.TrustedEdgeCountry(c)
}
