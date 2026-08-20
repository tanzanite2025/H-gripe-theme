package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type createPayPalOrderRequest struct {
	OrderNumber string `json:"order_number" binding:"required"`
	ReturnURL   string `json:"return_url"`
	CancelURL   string `json:"cancel_url"`
}

type capturePayPalOrderRequest struct {
	OrderNumber string `json:"order_number" binding:"required"`
}

func (h *Handler) CreatePayPalOrder(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req createPayPalOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	orderRecord, ok := h.loadPayablePayPalOrder(c, strings.TrimSpace(req.OrderNumber), userIDValue.(uint))
	if !ok {
		return
	}
	if !h.authorizePaymentStart(c, paymentStartProtectionInput{
		Provider:      string(pgateway.GatewayPayPal),
		PaymentMethod: "paypal",
		Order:         orderRecord,
	}) {
		return
	}

	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayPayPal)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		apierror.RespondError(c, http.StatusServiceUnavailable, "paypal_not_configured", "PayPal is not configured")
		return
	}

	orderCurrency, err := strictOrderCurrency(orderRecord)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !ensureGatewayCurrency(c, pgateway.GatewayPayPal, orderCurrency) {
		return
	}
	if !h.allowPaymentGatewayAttemptOrRespondWithFallbackRecommendation(c, pgateway.GatewayPayPal) {
		return
	}
	defer h.releasePaymentGatewayAttemptIfUnrecorded(c, pgateway.GatewayPayPal)
	gateway, err := h.createPaymentGatewayFromConfiguration(config)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayPayPal,
			http.StatusInternalServerError,
			"paypal_gateway_initialization_failed",
			err,
		)
		return
	}

	attempt, ok := h.ensurePaymentAttempt(
		c,
		pgateway.GatewayPayPal,
		"paypal",
		orderRecord,
		orderRecord.TotalAmount,
		orderCurrency,
	)
	if !ok {
		return
	}
	paymentResponse, err := gateway.CreatePayment(c.Request.Context(), &pgateway.PaymentRequest{
		Amount:         orderRecord.TotalAmount,
		Currency:       orderCurrency,
		OrderID:        orderRecord.OrderNumber,
		Description:    fmt.Sprintf("Order %s", orderRecord.OrderNumber),
		ReturnURL:      sanitizedPayPalRedirectURL(req.ReturnURL),
		CancelURL:      sanitizedPayPalRedirectURL(req.CancelURL),
		IdempotencyKey: attempt.ProviderRequestKey,
		Customer: &pgateway.Customer{
			Name:  strings.TrimSpace(orderRecord.ShippingAddress.FirstName + " " + orderRecord.ShippingAddress.LastName),
			Email: orderRecord.ShippingAddress.Email,
			Phone: orderRecord.ShippingAddress.Phone,
		},
		Metadata: map[string]string{
			"order_number": orderRecord.OrderNumber,
		},
	})
	if err != nil {
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:           string(pgateway.GatewayPayPal),
			OrderNumber:        orderRecord.OrderNumber,
			TransactionID:      attempt.TransactionID,
			AttemptKey:         attempt.AttemptKey,
			ProviderRequestKey: attempt.ProviderRequestKey,
			PaymentMethod:      "paypal",
			Status:             "failed",
			Amount:             orderRecord.TotalAmount,
			Currency:           orderCurrency,
			ErrorMessage:       err.Error(),
		})
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayPayPal,
			http.StatusBadGateway,
			"paypal_order_create_failed",
			err,
		)
		return
	}
	h.recordSuccessfulPaymentGatewayAPIResponse(c, pgateway.GatewayPayPal)

	gatewayResponse, _ := json.Marshal(paymentResponse)
	if err := h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
		Provider:           string(pgateway.GatewayPayPal),
		OrderNumber:        orderRecord.OrderNumber,
		TransactionID:      paymentResponse.TransactionID,
		AttemptKey:         attempt.AttemptKey,
		ProviderRequestKey: attempt.ProviderRequestKey,
		PaymentMethod:      "paypal",
		Status:             paypalAttemptStatus(paymentResponse.Status),
		Amount:             paymentResponse.Amount,
		Currency:           paymentResponse.Currency,
		GatewayResponse:    string(gatewayResponse),
	}); err != nil {
		respondVerifiedProviderPaymentError(c, err)
		return
	}

	response.Success(c, paymentResponse)
}

func (h *Handler) CapturePayPalOrder(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	paypalOrderID := strings.TrimSpace(c.Param("paypal_order_id"))
	if paypalOrderID == "" {
		apierror.RespondBadRequest(c, "paypal_order_id is required")
		return
	}
	var req capturePayPalOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	orderRecord, ok := h.loadPayablePayPalOrder(c, strings.TrimSpace(req.OrderNumber), userIDValue.(uint))
	if !ok {
		return
	}

	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayPayPal)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		apierror.RespondError(c, http.StatusServiceUnavailable, "paypal_not_configured", "PayPal is not configured")
		return
	}
	if !h.allowPaymentGatewayAttemptOrRespondWithFallbackRecommendation(c, pgateway.GatewayPayPal) {
		return
	}
	defer h.releasePaymentGatewayAttemptIfUnrecorded(c, pgateway.GatewayPayPal)

	gateway, err := h.createPaymentGatewayFromConfiguration(config)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayPayPal,
			http.StatusInternalServerError,
			"paypal_gateway_initialization_failed",
			err,
		)
		return
	}
	paymentResponse, err := capturePayPalPayment(
		c.Request.Context(),
		gateway,
		paypalOrderID,
		pgateway.PayPalCaptureRequestID(paypalOrderID),
	)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayPayPal,
			http.StatusBadGateway,
			"paypal_capture_failed",
			err,
		)
		return
	}
	h.recordSuccessfulPaymentGatewayAPIResponse(c, pgateway.GatewayPayPal)
	if !paypalResponseMatchesOrder(paymentResponse, orderRecord.OrderNumber) {
		apierror.RespondError(c, http.StatusBadRequest, "paypal_order_mismatch", "PayPal order does not match this order")
		return
	}
	gatewayResponse, _ := json.Marshal(paymentResponse)
	if !strings.EqualFold(strings.TrimSpace(paymentResponse.Status), "COMPLETED") {
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:        string(pgateway.GatewayPayPal),
			OrderNumber:     orderRecord.OrderNumber,
			TransactionID:   gatewayTransactionID(paymentResponse, paypalOrderID),
			PaymentMethod:   "paypal",
			Status:          paypalAttemptStatus(paymentResponse.Status),
			Amount:          paymentResponse.Amount,
			Currency:        paymentResponse.Currency,
			GatewayResponse: string(gatewayResponse),
		})
		apierror.RespondErrorWithDetails(c, http.StatusConflict, "paypal_capture_incomplete", "PayPal payment is not completed", gin.H{
			"status": paymentResponse.Status,
		})
		return
	}
	transactionID := paypalTransactionID(paymentResponse)
	if transactionID == "" {
		apierror.RespondError(c, http.StatusBadGateway, "paypal_capture_id_missing", "PayPal capture response did not return a capture id")
		return
	}

	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:        string(pgateway.GatewayPayPal),
		OrderNumber:     orderRecord.OrderNumber,
		TransactionID:   transactionID,
		PaymentMethod:   "paypal",
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

func capturePayPalPayment(
	ctx context.Context,
	gateway pgateway.PaymentGateway,
	paypalOrderID string,
	idempotencyKey string,
) (*pgateway.PaymentResponse, error) {
	if gatewayWithOptions, ok := gateway.(pgateway.CapturePaymentWithOptions); ok {
		return gatewayWithOptions.CapturePaymentWithOptions(ctx, paypalOrderID, pgateway.CaptureOptions{
			IdempotencyKey: idempotencyKey,
		})
	}
	return gateway.CapturePayment(ctx, paypalOrderID)
}

func (h *Handler) loadPayablePayPalOrder(c *gin.Context, orderNumber string, userID uint) (*orderdomain.Order, bool) {
	orderRecord, err := h.orderService.GetOrderByNumber(orderNumber, userID)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return nil, false
	}
	if pgateway.ProviderForPaymentMethod(orderRecord.PaymentMethod) != string(pgateway.GatewayPayPal) {
		apierror.RespondBadRequest(c, "Order payment method is not PayPal")
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
	return orderRecord, true
}

func sanitizedPayPalRedirectURL(value string) string {
	return sanitizedPaymentRedirectURL(value)
}

func paypalResponseMatchesOrder(paymentResponse *pgateway.PaymentResponse, orderNumber string) bool {
	if paymentResponse == nil {
		return false
	}
	orderNumber = strings.TrimSpace(orderNumber)
	for _, key := range []string{"order_number", "order_id", "custom_id"} {
		if strings.TrimSpace(paymentResponse.Metadata[key]) == orderNumber {
			return true
		}
	}
	return false
}

func paypalTransactionID(paymentResponse *pgateway.PaymentResponse) string {
	if paymentResponse != nil {
		if value := strings.TrimSpace(paymentResponse.TransactionID); value != "" {
			return value
		}
	}
	return ""
}

func paypalAttemptStatus(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "COMPLETED":
		return "completed"
	case "PAYER_ACTION_REQUIRED", "APPROVED":
		return "requires_action"
	case "CREATED", "SAVED", "VOIDED":
		return "pending"
	case "FAILED", "DECLINED":
		return "failed"
	default:
		return "pending"
	}
}
