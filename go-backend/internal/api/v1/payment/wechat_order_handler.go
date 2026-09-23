package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type createWechatOrderRequest struct {
	OrderNumber string `json:"order_number" binding:"required"`
}

func (h *Handler) CreateWechatOrder(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req createWechatOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	orderRecord, ok := h.loadPayableProviderOrder(
		c,
		req.OrderNumber,
		userIDValue.(uint),
		pgateway.GatewayWechat,
		"WeChat Pay",
	)
	if !ok {
		return
	}
	if !h.authorizePaymentStart(c, paymentStartProtectionInput{
		Provider:      string(pgateway.GatewayWechat),
		PaymentMethod: "wechat",
		Order:         orderRecord,
	}) {
		return
	}

	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayWechat)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !wechatGatewayConfigured(config) {
		apierror.RespondError(c, http.StatusServiceUnavailable, "wechat_not_configured", "WeChat Pay is not configured")
		return
	}

	settlement, err := strictProviderSettlement(orderRecord)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	orderCurrency := settlement.Currency().String()
	if !ensureGatewayCurrency(c, pgateway.GatewayWechat, orderCurrency) {
		return
	}
	if !h.allowPaymentGatewayAttemptOrRespondWithFallbackRecommendation(c, pgateway.GatewayWechat) {
		return
	}
	defer h.releasePaymentGatewayAttemptIfUnrecorded(c, pgateway.GatewayWechat)
	gateway, err := h.createPaymentGatewayFromConfiguration(config)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayWechat,
			http.StatusInternalServerError,
			"wechat_gateway_initialization_failed",
			err,
		)
		return
	}

	attempt, ok := h.ensurePaymentAttempt(
		c,
		pgateway.GatewayWechat,
		"wechat",
		orderRecord,
		settlement,
	)
	if !ok {
		return
	}
	paymentResponse, err := gateway.CreatePayment(c.Request.Context(), &pgateway.PaymentRequest{
		AmountMinor:    settlement.AmountMinor(),
		Currency:       orderCurrency,
		OrderID:        orderRecord.OrderNumber,
		Description:    fmt.Sprintf("Order %s", orderRecord.OrderNumber),
		NotifyURL:      h.providerWebhookURL(c, pgateway.GatewayWechat),
		IdempotencyKey: attempt.ProviderRequestKey,
		Customer:       paymentCustomerFromOrder(orderRecord),
		Metadata: map[string]string{
			"order_number": orderRecord.OrderNumber,
			"order_id":     orderRecord.OrderNumber,
			"out_trade_no": orderRecord.OrderNumber,
		},
	})
	if err != nil {
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:           string(pgateway.GatewayWechat),
			OrderNumber:        orderRecord.OrderNumber,
			TransactionID:      attempt.TransactionID,
			AttemptKey:         attempt.AttemptKey,
			ProviderRequestKey: attempt.ProviderRequestKey,
			PaymentMethod:      "wechat",
			Status:             "failed",
			Amount:             settlement,
			ErrorMessage:       err.Error(),
		})
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayWechat,
			http.StatusBadGateway,
			"wechat_order_create_failed",
			err,
		)
		return
	}
	h.recordSuccessfulPaymentGatewayAPIResponse(c, pgateway.GatewayWechat)

	gatewayResponse, _ := json.Marshal(paymentResponse)
	providerAmount, amountErr := providerPaymentResponseMoney(paymentResponse, settlement)
	if amountErr != nil {
		apierror.RespondInternalError(c, amountErr)
		return
	}
	if err := h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
		Provider:           string(pgateway.GatewayWechat),
		OrderNumber:        orderRecord.OrderNumber,
		TransactionID:      gatewayTransactionID(paymentResponse, orderRecord.OrderNumber),
		AttemptKey:         attempt.AttemptKey,
		ProviderRequestKey: attempt.ProviderRequestKey,
		PaymentMethod:      "wechat",
		Status:             wechatAttemptStatus(paymentResponse.Status),
		Amount:             providerAmount,
		GatewayResponse:    string(gatewayResponse),
	}); err != nil {
		respondVerifiedProviderPaymentError(c, err)
		return
	}

	response.Success(c, paymentResponse)
}

func (h *Handler) ConfirmWechatOrder(c *gin.Context) {
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
		pgateway.GatewayWechat,
		"WeChat Pay",
	)
	if !ok {
		return
	}
	if orderRecord.PaymentStatus == "paid" {
		settlement, err := strictProviderSettlement(orderRecord)
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
		orderAmount, err := settlement.FormatMajor()
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
		response.Success(c, &pgateway.PaymentResponse{
			ID:            orderRecord.OrderNumber,
			Status:        wechatTradeStateSuccess,
			Amount:        orderAmount,
			Currency:      settlement.Currency().String(),
			TransactionID: orderRecord.OrderNumber,
		})
		return
	}

	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayWechat)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !wechatGatewayConfigured(config) {
		apierror.RespondError(c, http.StatusServiceUnavailable, "wechat_not_configured", "WeChat Pay is not configured")
		return
	}
	if !h.allowPaymentGatewayAttemptOrRespondWithFallbackRecommendation(c, pgateway.GatewayWechat) {
		return
	}
	defer h.releasePaymentGatewayAttemptIfUnrecorded(c, pgateway.GatewayWechat)

	gateway, err := h.createPaymentGatewayFromConfiguration(config)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayWechat,
			http.StatusInternalServerError,
			"wechat_gateway_initialization_failed",
			err,
		)
		return
	}
	paymentResponse, err := gateway.GetPayment(c.Request.Context(), orderRecord.OrderNumber)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayWechat,
			http.StatusBadGateway,
			"wechat_order_query_failed",
			err,
		)
		return
	}
	h.recordSuccessfulPaymentGatewayAPIResponse(c, pgateway.GatewayWechat)
	if !providerPaymentResponseMatchesOrder(paymentResponse, orderRecord.OrderNumber) {
		apierror.RespondError(c, http.StatusBadRequest, "wechat_order_mismatch", "WeChat Pay trade does not match this order")
		return
	}

	gatewayResponse, _ := json.Marshal(paymentResponse)
	settlement, err := strictProviderSettlement(orderRecord)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !isWechatPaidStatus(paymentResponse.Status) {
		providerAmount, amountErr := providerPaymentResponseMoney(paymentResponse, settlement)
		if amountErr != nil {
			apierror.RespondInternalError(c, amountErr)
			return
		}
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:        string(pgateway.GatewayWechat),
			OrderNumber:     orderRecord.OrderNumber,
			TransactionID:   gatewayTransactionID(paymentResponse, orderRecord.OrderNumber),
			PaymentMethod:   "wechat",
			Status:          wechatAttemptStatus(paymentResponse.Status),
			Amount:          providerAmount,
			GatewayResponse: string(gatewayResponse),
		})
		apierror.RespondErrorWithDetails(c, http.StatusConflict, "wechat_payment_incomplete", "WeChat Pay payment is not completed", gin.H{
			"status": paymentResponse.Status,
		})
		return
	}
	transactionID := providerTransactionID(paymentResponse)
	if transactionID == "" {
		apierror.RespondError(c, http.StatusBadGateway, "wechat_transaction_id_missing", "WeChat Pay trade query did not return a transaction id")
		return
	}
	providerAmount, amountErr := providerPaymentResponseMoney(paymentResponse, settlement)
	if amountErr != nil {
		apierror.RespondInternalError(c, amountErr)
		return
	}

	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:        string(pgateway.GatewayWechat),
		OrderNumber:     orderRecord.OrderNumber,
		TransactionID:   transactionID,
		PaymentMethod:   "wechat",
		Amount:          providerAmount,
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

func wechatGatewayConfigured(config *pgateway.Config) bool {
	if config == nil {
		return false
	}
	hasPlatformVerifier := strings.TrimSpace(config.WechatPayPlatformCertificate) != "" ||
		(strings.TrimSpace(config.WechatPayPlatformPublicKey) != "" && strings.TrimSpace(config.WechatPayPlatformPublicKeyID) != "")
	return strings.TrimSpace(config.APIKey) != "" &&
		strings.TrimSpace(config.WechatAppID) != "" &&
		strings.TrimSpace(config.SecretKey) != "" &&
		strings.TrimSpace(config.WebhookSecret) != "" &&
		strings.TrimSpace(config.WechatAPIv3Key) != "" &&
		hasPlatformVerifier
}

func isWechatPaidStatus(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), wechatTradeStateSuccess)
}

func wechatAttemptStatus(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case wechatTradeStateSuccess:
		return "processing"
	case "USERPAYING":
		return "processing"
	case "PAYERROR", "CLOSED", "REVOKED":
		return "failed"
	default:
		return "pending"
	}
}
