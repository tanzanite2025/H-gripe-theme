package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/pkg/apierror"
	pgateway "tanzanite/internal/pkg/payment"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

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

	config, err := h.loadGatewayConfig(pgateway.GatewayPayPal)
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
		ReturnURL:      sanitizedPayPalRedirectURL(req.ReturnURL),
		CancelURL:      sanitizedPayPalRedirectURL(req.CancelURL),
		IdempotencyKey: fmt.Sprintf("order-%d-paypal-order", orderRecord.ID),
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
		apierror.RespondError(c, http.StatusBadGateway, "paypal_order_create_failed", err.Error())
		return
	}

	gatewayResponse, _ := json.Marshal(paymentResponse)
	if err := h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
		Provider:        string(pgateway.GatewayPayPal),
		OrderNumber:     orderRecord.OrderNumber,
		TransactionID:   paymentResponse.TransactionID,
		PaymentMethod:   "paypal",
		Status:          paypalAttemptStatus(paymentResponse.Status),
		Amount:          paymentResponse.Amount,
		Currency:        paymentResponse.Currency,
		GatewayResponse: string(gatewayResponse),
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

	config, err := h.loadGatewayConfig(pgateway.GatewayPayPal)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		apierror.RespondError(c, http.StatusServiceUnavailable, "paypal_not_configured", "PayPal is not configured")
		return
	}

	gateway, err := h.newPaymentGateway(config)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	paymentResponse, err := gateway.CapturePayment(c.Request.Context(), paypalOrderID)
	if err != nil {
		apierror.RespondError(c, http.StatusBadGateway, "paypal_capture_failed", err.Error())
		return
	}
	if !paypalResponseMatchesOrder(paymentResponse, orderRecord.OrderNumber) {
		apierror.RespondError(c, http.StatusBadRequest, "paypal_order_mismatch", "PayPal order does not match this order")
		return
	}
	gatewayResponse, _ := json.Marshal(paymentResponse)
	if !strings.EqualFold(strings.TrimSpace(paymentResponse.Status), "COMPLETED") {
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:        string(pgateway.GatewayPayPal),
			OrderNumber:     orderRecord.OrderNumber,
			TransactionID:   paypalTransactionID(paymentResponse, paypalOrderID),
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

	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:        string(pgateway.GatewayPayPal),
		OrderNumber:     orderRecord.OrderNumber,
		TransactionID:   paypalTransactionID(paymentResponse, paypalOrderID),
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

func paypalTransactionID(paymentResponse *pgateway.PaymentResponse, fallback string) string {
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
