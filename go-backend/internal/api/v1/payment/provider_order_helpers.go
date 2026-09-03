package payment

import (
	"net/http"
	"net/url"
	"strings"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
)

func (h *Handler) loadPayableProviderOrder(
	c *gin.Context,
	orderNumber string,
	userID uint,
	provider pgateway.GatewayType,
	paymentMethodLabel string,
) (*orderdomain.Order, bool) {
	orderRecord, err := h.orderService.GetOrderByNumber(strings.TrimSpace(orderNumber), userID)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return nil, false
	}
	if pgateway.ProviderForPaymentMethod(orderRecord.PaymentMethod) != string(provider) {
		apierror.RespondBadRequest(c, "Order payment method is not "+paymentMethodLabel)
		return nil, false
	}
	if orderRecord.PaymentStatus == "paid" {
		apierror.RespondBadRequest(c, "Order is already paid")
		return nil, false
	}
	if orderRecord.Status == "cancelled" || orderRecord.Status == "refunded" || orderRecord.Status == "payment_expired" {
		apierror.RespondBadRequest(c, "Order is not payable")
		return nil, false
	}
	if !ensureOrderHasPayableAmount(c, orderRecord) {
		return nil, false
	}
	return orderRecord, true
}

func (h *Handler) loadProviderOrderForConfirmation(
	c *gin.Context,
	orderNumber string,
	userID uint,
	provider pgateway.GatewayType,
	paymentMethodLabel string,
) (*orderdomain.Order, bool) {
	orderRecord, err := h.orderService.GetOrderByNumber(strings.TrimSpace(orderNumber), userID)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return nil, false
	}
	if pgateway.ProviderForPaymentMethod(orderRecord.PaymentMethod) != string(provider) {
		apierror.RespondBadRequest(c, "Order payment method is not "+paymentMethodLabel)
		return nil, false
	}
	if orderRecord.Status == "cancelled" || orderRecord.Status == "refunded" {
		apierror.RespondBadRequest(c, "Order is not payable")
		return nil, false
	}
	if !ensureOrderHasPayableAmount(c, orderRecord) {
		return nil, false
	}
	return orderRecord, true
}

func ensureOrderHasPayableAmount(c *gin.Context, orderRecord *orderdomain.Order) bool {
	if orderRecord == nil || orderRecord.TotalAmount > 0 {
		return true
	}
	apierror.RespondError(c, http.StatusBadRequest, "order_has_no_payable_amount", "Order has no payable amount")
	return false
}

func paymentCustomerFromOrder(orderRecord *orderdomain.Order) *pgateway.Customer {
	if orderRecord == nil {
		return nil
	}
	return &pgateway.Customer{
		Name:  strings.TrimSpace(orderRecord.ShippingAddress.FirstName + " " + orderRecord.ShippingAddress.LastName),
		Email: strings.TrimSpace(orderRecord.ShippingAddress.Email),
		Phone: strings.TrimSpace(orderRecord.ShippingAddress.Phone),
	}
}

func sanitizedPaymentRedirectURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 2048 {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	return parsed.String()
}

func (h *Handler) providerWebhookURL(c *gin.Context, provider pgateway.GatewayType) string {
	baseURL := ""
	if h != nil {
		baseURL = h.publicBaseURL
	}
	if baseURL == "" {
		baseURL = requestBaseURL(c)
	}
	if baseURL == "" {
		return ""
	}
	return pgateway.GatewayWebhookURL(baseURL, provider)
}

func normalizePaymentBaseURL(value string) string {
	return pgateway.NormalizePublicBaseURL(value)
}

func requestBaseURL(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	scheme := firstForwardedHeaderValue(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := firstForwardedHeaderValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

func firstForwardedHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, ","); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(value)
}

func providerPaymentResponseMatchesOrder(paymentResponse *pgateway.PaymentResponse, orderNumber string) bool {
	if paymentResponse == nil {
		return false
	}
	orderNumber = strings.TrimSpace(orderNumber)
	if strings.TrimSpace(paymentResponse.ID) == orderNumber {
		return true
	}
	for _, key := range []string{"order_number", "order_id", "out_trade_no"} {
		if strings.TrimSpace(paymentResponse.Metadata[key]) == orderNumber {
			return true
		}
	}
	return false
}

func gatewayTransactionID(paymentResponse *pgateway.PaymentResponse, fallback string) string {
	if paymentResponse != nil {
		if value := strings.TrimSpace(paymentResponse.TransactionID); value != "" {
			return value
		}
		if value := strings.TrimSpace(paymentResponse.ID); value != "" {
			return value
		}
	}
	return strings.TrimSpace(fallback)
}

func providerTransactionID(paymentResponse *pgateway.PaymentResponse) string {
	if paymentResponse == nil {
		return ""
	}
	return strings.TrimSpace(paymentResponse.TransactionID)
}

func ensureGatewayCurrency(c *gin.Context, provider pgateway.GatewayType, currency string) bool {
	if err := pgateway.ValidateGatewayCurrency(provider, currency); err != nil {
		apierror.RespondError(c, http.StatusBadRequest, "provider_order_currency_not_supported", err.Error())
		return false
	}
	return true
}
