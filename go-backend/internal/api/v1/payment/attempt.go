package payment

import (
	"commerce-platform/internal/api/middleware"
	domainmoney "commerce-platform/internal/domain/money"
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
	amount domainmoney.Money,
) (*paymentdomain.Transaction, bool) {
	if h == nil || h.paymentService == nil || orderRecord == nil {
		return nil, false
	}
	attemptKey := service.NormalizePaymentAttemptKey(middleware.GetIdempotencyKey(c))
	if attemptKey == "" {
		apierror.RespondBadRequest(c, "Idempotency-Key header is required")
		return nil, false
	}
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
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return nil, false
	}
	return attempt, true
}
