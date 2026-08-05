package payment

import (
	"net/http"
	"strings"

	"tanzanite/internal/api/middleware"
	orderdomain "tanzanite/internal/domain/order"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type paymentStartProtectionInput struct {
	Provider      string
	PaymentMethod string
	Order         *orderdomain.Order
}

// authorizePaymentStart is the common guard for any payment endpoint that
// creates a provider payment intent or hosted checkout session.
func (h *Handler) authorizePaymentStart(c *gin.Context, input paymentStartProtectionInput) bool {
	if h == nil || h.protection == nil || input.Order == nil {
		return true
	}

	decision, err := h.protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      input.provider(),
		Country:       paymentProtectionOrderCountry(c, input.Order),
		PaymentMethod: input.paymentMethod(),
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

func (input paymentStartProtectionInput) provider() string {
	return strings.ToLower(strings.TrimSpace(input.Provider))
}

func (input paymentStartProtectionInput) paymentMethod() string {
	if method := strings.ToLower(strings.TrimSpace(input.PaymentMethod)); method != "" {
		return method
	}
	if input.Order != nil {
		if method := strings.ToLower(strings.TrimSpace(input.Order.PaymentMethod)); method != "" {
			return method
		}
	}
	return input.provider()
}

func paymentProtectionOrderCountry(c *gin.Context, orderRecord *orderdomain.Order) string {
	if orderRecord == nil {
		return paymentRequestCountry(c)
	}
	if country := strings.TrimSpace(orderRecord.BillingAddress.Country); country != "" {
		return country
	}
	if country := strings.TrimSpace(orderRecord.ShippingAddress.Country); country != "" {
		return country
	}
	return paymentRequestCountry(c)
}

func paymentRequestCountry(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return middleware.TrustedEdgeCountry(c)
}
