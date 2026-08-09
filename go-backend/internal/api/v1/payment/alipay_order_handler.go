package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"tanzanite/internal/pkg/apierror"
	pgateway "tanzanite/internal/pkg/payment"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type createAlipayOrderRequest struct {
	OrderNumber string `json:"order_number" binding:"required"`
	ReturnURL   string `json:"return_url"`
	CancelURL   string `json:"cancel_url"`
}

func (h *Handler) CreateAlipayOrder(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req createAlipayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	orderRecord, ok := h.loadPayableProviderOrder(
		c,
		req.OrderNumber,
		userIDValue.(uint),
		pgateway.GatewayAlipay,
		"Alipay",
	)
	if !ok {
		return
	}
	if !h.authorizePaymentStart(c, paymentStartProtectionInput{
		Provider:      string(pgateway.GatewayAlipay),
		PaymentMethod: "alipay",
		Order:         orderRecord,
	}) {
		return
	}

	config, err := h.loadGatewayConfig(pgateway.GatewayAlipay)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !alipayGatewayConfigured(config) {
		apierror.RespondError(c, http.StatusServiceUnavailable, "alipay_not_configured", "Alipay is not configured")
		return
	}

	orderCurrency, err := strictOrderCurrency(orderRecord)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !ensureGatewayCurrency(c, pgateway.GatewayAlipay, orderCurrency) {
		return
	}
	gateway, err := h.newPaymentGateway(config)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	paymentResponse, err := gateway.CreatePayment(c.Request.Context(), &pgateway.PaymentRequest{
		Amount:         orderRecord.TotalAmount,
		Currency:       orderCurrency,
		OrderID:        orderRecord.OrderNumber,
		Description:    fmt.Sprintf("Order %s", orderRecord.OrderNumber),
		ReturnURL:      sanitizedPaymentRedirectURL(req.ReturnURL),
		CancelURL:      sanitizedPaymentRedirectURL(req.CancelURL),
		NotifyURL:      h.providerWebhookURL(c, pgateway.GatewayAlipay),
		IdempotencyKey: fmt.Sprintf("order-%d-alipay-page-pay", orderRecord.ID),
		Customer:       paymentCustomerFromOrder(orderRecord),
		Metadata: map[string]string{
			"order_number": orderRecord.OrderNumber,
			"order_id":     orderRecord.OrderNumber,
			"out_trade_no": orderRecord.OrderNumber,
		},
	})
	if err != nil {
		apierror.RespondError(c, http.StatusBadGateway, "alipay_order_create_failed", err.Error())
		return
	}

	gatewayResponse, _ := json.Marshal(paymentResponse)
	if err := h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
		Provider:        string(pgateway.GatewayAlipay),
		OrderNumber:     orderRecord.OrderNumber,
		TransactionID:   gatewayTransactionID(paymentResponse, orderRecord.OrderNumber),
		PaymentMethod:   "alipay",
		Status:          alipayAttemptStatus(paymentResponse.Status),
		Amount:          paymentResponse.Amount,
		Currency:        paymentResponse.Currency,
		GatewayResponse: string(gatewayResponse),
	}); err != nil {
		respondVerifiedProviderPaymentError(c, err)
		return
	}

	response.Success(c, paymentResponse)
}

func (h *Handler) ConfirmAlipayOrder(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	orderNumber := strings.TrimSpace(c.Param("order_number"))
	if orderNumber == "" {
		apierror.RespondBadRequest(c, "order_number is required")
		return
	}

	orderRecord, ok := h.loadProviderOrderForConfirmation(
		c,
		orderNumber,
		userIDValue.(uint),
		pgateway.GatewayAlipay,
		"Alipay",
	)
	if !ok {
		return
	}
	if orderRecord.PaymentStatus == "paid" {
		response.Success(c, &pgateway.PaymentResponse{
			ID:            orderRecord.OrderNumber,
			Status:        alipayTradeStatusSuccess,
			Amount:        orderRecord.TotalAmount,
			Currency:      orderRecord.Currency,
			TransactionID: orderRecord.OrderNumber,
		})
		return
	}

	config, err := h.loadGatewayConfig(pgateway.GatewayAlipay)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !alipayGatewayConfigured(config) {
		apierror.RespondError(c, http.StatusServiceUnavailable, "alipay_not_configured", "Alipay is not configured")
		return
	}

	gateway, err := h.newPaymentGateway(config)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	paymentResponse, err := gateway.GetPayment(c.Request.Context(), orderRecord.OrderNumber)
	if err != nil {
		apierror.RespondError(c, http.StatusBadGateway, "alipay_order_query_failed", err.Error())
		return
	}
	if !providerPaymentResponseMatchesOrder(paymentResponse, orderRecord.OrderNumber) {
		apierror.RespondError(c, http.StatusBadRequest, "alipay_order_mismatch", "Alipay trade does not match this order")
		return
	}

	gatewayResponse, _ := json.Marshal(paymentResponse)
	if !isAlipayPaidStatus(paymentResponse.Status) {
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:        string(pgateway.GatewayAlipay),
			OrderNumber:     orderRecord.OrderNumber,
			TransactionID:   gatewayTransactionID(paymentResponse, orderRecord.OrderNumber),
			PaymentMethod:   "alipay",
			Status:          alipayAttemptStatus(paymentResponse.Status),
			Amount:          paymentResponse.Amount,
			Currency:        paymentResponse.Currency,
			GatewayResponse: string(gatewayResponse),
		})
		apierror.RespondErrorWithDetails(c, http.StatusConflict, "alipay_payment_incomplete", "Alipay payment is not completed", gin.H{
			"status": paymentResponse.Status,
		})
		return
	}
	transactionID := providerTransactionID(paymentResponse)
	if transactionID == "" {
		apierror.RespondError(c, http.StatusBadGateway, "alipay_transaction_id_missing", "Alipay trade query did not return a trade number")
		return
	}

	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:        string(pgateway.GatewayAlipay),
		OrderNumber:     orderRecord.OrderNumber,
		TransactionID:   transactionID,
		PaymentMethod:   "alipay",
		Amount:          paymentResponse.Amount,
		Currency:        paymentResponse.Currency,
		GatewayResponse: string(gatewayResponse),
	}); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			apierror.RespondNotFound(c, "Order")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, paymentResponse)
}

func alipayGatewayConfigured(config *pgateway.Config) bool {
	return config != nil &&
		strings.TrimSpace(config.APIKey) != "" &&
		strings.TrimSpace(config.SecretKey) != "" &&
		strings.TrimSpace(config.WebhookSecret) != ""
}

func isAlipayPaidStatus(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case alipayTradeStatusSuccess, alipayTradeStatusFinished:
		return true
	default:
		return false
	}
}

func alipayAttemptStatus(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case alipayTradeStatusSuccess, alipayTradeStatusFinished:
		return "processing"
	case "WAIT_BUYER_PAY":
		return "pending"
	case "TRADE_CLOSED":
		return "failed"
	default:
		return "pending"
	}
}
