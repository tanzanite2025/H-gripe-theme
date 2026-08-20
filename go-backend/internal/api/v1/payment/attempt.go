package payment

import (
	"commerce-platform/internal/api/middleware"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ensurePaymentAttempt(
	c *gin.Context,
	provider pgateway.GatewayType,
	paymentMethod string,
	orderRecord *orderdomain.Order,
	amount float64,
	currency string,
) (*paymentdomain.Transaction, bool) {
	if h == nil || h.paymentService == nil || orderRecord == nil {
		return nil, false
	}
	attemptKey := service.NormalizePaymentAttemptKey(
		string(provider),
		orderRecord.ID,
		middleware.GetIdempotencyKey(c),
	)
	providerRequestKey := service.PaymentProviderRequestKey(
		string(provider),
		orderRecord.ID,
		attemptKey,
	)
	attempt, err := h.paymentService.EnsureGatewayPaymentAttempt(service.EnsureGatewayPaymentAttemptInput{
		Provider:           string(provider),
		OrderNumber:        orderRecord.OrderNumber,
		AttemptKey:         attemptKey,
		ProviderRequestKey: providerRequestKey,
		PaymentMethod:      paymentMethod,
		Amount:             amount,
		Currency:           currency,
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return nil, false
	}
	return attempt, true
}
