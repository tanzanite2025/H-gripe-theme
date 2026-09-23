package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/plutov/paypal/v4"
)

const (
	paypalCheckoutOrderCompleted  = "CHECKOUT.ORDER.COMPLETED"
	paypalCheckoutOrderApproved   = "CHECKOUT.ORDER.APPROVED"
	paypalPaymentCaptureCompleted = "PAYMENT.CAPTURE.COMPLETED"
	paypalPaymentCaptureDenied    = "PAYMENT.CAPTURE.DENIED"
	paypalPaymentCaptureRefunded  = "PAYMENT.CAPTURE.REFUNDED"
	alipayTradeStatusSuccess      = "TRADE_SUCCESS"
	alipayTradeStatusFinished     = "TRADE_FINISHED"
	wechatTradeStateSuccess       = "SUCCESS"
)

func providerRefundAmountMajor(value domainmoney.Money) string {
	amount, _ := value.FormatMajor()
	return amount
}

type verifiedProviderPayment struct {
	Provider         pgateway.GatewayType
	OrderNumber      string
	TransactionID    string
	PaymentMethod    string
	Amount           domainmoney.Money
	GatewayResponse  string
	LiabilityShifted *bool
}

type paypalRefundWebhookResource struct {
	ID        string                     `json:"id"`
	Status    string                     `json:"status"`
	State     string                     `json:"state"`
	CaptureID string                     `json:"capture_id"`
	CustomID  string                     `json:"custom_id"`
	InvoiceID string                     `json:"invoice_id"`
	Amount    *paypalRefundWebhookAmount `json:"amount"`
	Links     []paypal.Link              `json:"links"`
}

type paypalRefundWebhookAmount struct {
	CurrencyCode string `json:"currency_code"`
	Currency     string `json:"currency"`
	Value        string `json:"value"`
	Total        string `json:"total"`
}

func (h *Handler) handlePayPalWebhook(c *gin.Context, payload []byte) {
	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayPayPal)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	event, err := pgateway.VerifyPayPalWebhook(c.Request.Context(), config, c.Request.Header, payload, nil)
	if err != nil {
		respondPayPalWebhookVerificationError(c, err)
		return
	}
	claimed, err := h.paymentService.ClaimPayPalWebhookEvent(event.ID, event.EventType, string(payload))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !claimed {
		response.SuccessWithMessage(c, "PayPal event already handled", gin.H{"event_id": event.ID, "event_type": event.EventType})
		return
	}
	defer func() {
		if c.Writer.Status() >= http.StatusBadRequest {
			_ = h.paymentService.MarkPayPalWebhookEventFailed(event.ID, fmt.Errorf("paypal event handler returned HTTP %d", c.Writer.Status()))
		}
	}()

	if riskHandled, riskErr := h.recordPayPalDisputeRiskEvent(c, event, payload); riskErr != nil {
		apierror.RespondInternalError(c, riskErr)
		return
	} else if riskHandled {
		details := gin.H{
			"event_id":   event.ID,
			"event_type": event.EventType,
		}
		if value, exists := c.Get("paypal_dispute_id"); exists {
			details["dispute_id"] = value
		}
		if value, exists := c.Get("paypal_dispute_external_id"); exists {
			details["paypal_dispute_id"] = value
		}
		if value, exists := c.Get("paypal_dispute_order_id"); exists {
			details["order_id"] = value
		}
		if value, exists := c.Get("paypal_dispute_evidence_submitted"); exists {
			details["evidence_submitted"] = value
		}
		if value, exists := c.Get("paypal_dispute_evidence_pending_review"); exists {
			details["evidence_pending_review"] = value
		}
		if value, exists := c.Get("paypal_dispute_evidence_tracking_number"); exists {
			details["tracking_number"] = value
		}
		if value, exists := c.Get("paypal_dispute_evidence_error"); exists {
			details["evidence_error"] = value
		}
		response.SuccessWithMessage(c, "PayPal dispute risk event recorded", details)
		return
	}

	if strings.EqualFold(strings.TrimSpace(event.EventType), paypalPaymentCaptureRefunded) {
		refund, handled, err := paypalVerifiedRefundFromEvent(event, payload)
		if err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		if !handled {
			response.SuccessWithMessage(c, "Ignored unsupported PayPal event", gin.H{
				"event_id":   event.ID,
				"event_type": event.EventType,
			})
			return
		}
		if !h.recordVerifiedGatewayRefund(c, refund) {
			return
		}
		response.SuccessWithMessage(c, "PayPal refund webhook processed successfully", gin.H{
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"order_number":    refund.OrderNumber,
			"transaction_id":  refund.TransactionID,
			"refund_id":       refund.RefundID,
			"refund_amount":   providerRefundAmountMajor(refund.ProviderRefundAmount),
			"refund_currency": refund.ProviderRefundAmount.Currency().String(),
		})
		return
	}

	if strings.EqualFold(strings.TrimSpace(event.EventType), paypalPaymentCaptureDenied) {
		response.SuccessWithMessage(c, "PayPal capture denial recorded", gin.H{
			"event_id": event.ID, "event_type": event.EventType,
		})
		return
	}
	if strings.EqualFold(strings.TrimSpace(event.EventType), paypalCheckoutOrderApproved) {
		payment, err := h.captureApprovedPayPalOrder(c, event, payload)
		if err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		processed, result := h.recordVerifiedProviderPayment(c, payment)
		if !processed {
			return
		}
		message := "PayPal approved order captured successfully"
		details := gin.H{"event_id": event.ID, "event_type": event.EventType, "order_number": payment.OrderNumber, "transaction_id": payment.TransactionID}
		if result.DuplicatePaid {
			message = "PayPal duplicate payment recorded; refund is pending"
			details["duplicate_paid"] = true
			details["refund_id"] = result.RefundID
		}
		response.SuccessWithMessage(c, message, details)
		return
	}

	payment, handled, err := paypalVerifiedPaymentFromEvent(event, payload)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !handled {
		response.SuccessWithMessage(c, "Ignored unsupported PayPal event", gin.H{
			"event_id":   event.ID,
			"event_type": event.EventType,
		})
		return
	}

	processed, result := h.recordVerifiedProviderPayment(c, payment)
	if !processed {
		return
	}
	message := "PayPal webhook processed successfully"
	details := gin.H{
		"event_id":       event.ID,
		"event_type":     event.EventType,
		"order_number":   payment.OrderNumber,
		"transaction_id": payment.TransactionID,
	}
	if result.DuplicatePaid {
		message = "PayPal duplicate payment recorded; refund is pending"
		details["duplicate_paid"] = true
		details["refund_id"] = result.RefundID
	}
	response.SuccessWithMessage(c, message, details)
}

func respondPayPalWebhookVerificationError(c *gin.Context, err error) {
	if errors.Is(err, pgateway.ErrPayPalWebhookVerificationUnavailable) {
		apierror.RespondError(c, http.StatusServiceUnavailable, "paypal_webhook_verification_unavailable", "PayPal webhook verification is temporarily unavailable")
		return
	}
	apierror.RespondUnauthorized(c)
}

// captureApprovedPayPalOrder is the browser-independent fallback for the
// approval redirect. PayPal webhooks do not carry the storefront session.
func (h *Handler) captureApprovedPayPalOrder(c *gin.Context, event pgateway.PayPalWebhookEvent, rawPayload []byte) (verifiedProviderPayment, error) {
	var order paypal.Order
	if err := json.Unmarshal(event.Resource, &order); err != nil {
		return verifiedProviderPayment{}, fmt.Errorf("invalid paypal approved order resource: %w", err)
	}
	paypalOrderID := strings.TrimSpace(order.ID)
	if paypalOrderID == "" {
		return verifiedProviderPayment{}, errors.New("paypal approved order id is required")
	}
	orderNumber := ""
	for _, unit := range order.PurchaseUnits {
		orderNumber = firstNonBlank(unit.CustomID, unit.InvoiceID, unit.ReferenceID)
		if orderNumber != "" {
			break
		}
	}
	if orderNumber == "" {
		return verifiedProviderPayment{}, errors.New("paypal approved order does not contain order metadata")
	}
	orderRecord, err := h.orderService.GetOrderByNumberForPayment(orderNumber)
	if err != nil {
		return verifiedProviderPayment{}, service.ErrOrderNotFound
	}
	if pgateway.ProviderForPaymentMethod(orderRecord.PaymentMethod) != string(pgateway.GatewayPayPal) {
		return verifiedProviderPayment{}, errors.New("paypal approved order payment method mismatch")
	}
	if orderRecord.PaymentStatus == "paid" {
		return verifiedProviderPayment{}, errors.New("paypal order is already paid")
	}
	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayPayPal)
	if err != nil {
		return verifiedProviderPayment{}, err
	}
	gateway, err := h.createPaymentGatewayFromConfiguration(config)
	if err != nil {
		return verifiedProviderPayment{}, err
	}
	middleware.MarkPaymentOperationExternalCallStarted(c)
	resp, err := capturePayPalPayment(c.Request.Context(), gateway, paypalOrderID, pgateway.PayPalCaptureRequestID(paypalOrderID))
	if err != nil {
		return verifiedProviderPayment{}, fmt.Errorf("paypal capture failed: %w", err)
	}
	if !paypalResponseMatchesOrder(resp, orderNumber) {
		return verifiedProviderPayment{}, errors.New("paypal order does not match local order")
	}
	if !strings.EqualFold(strings.TrimSpace(resp.Status), "COMPLETED") {
		return verifiedProviderPayment{}, fmt.Errorf("paypal payment is not completed: %s", resp.Status)
	}
	transactionID := paypalTransactionID(resp)
	if transactionID == "" {
		return verifiedProviderPayment{}, errors.New("paypal capture id is missing")
	}
	amount, err := strictProviderSettlement(orderRecord)
	if err != nil {
		return verifiedProviderPayment{}, err
	}
	providerAmount, err := providerPaymentResponseMoney(resp, amount)
	if err != nil {
		return verifiedProviderPayment{}, err
	}
	gatewayResponse, _ := json.Marshal(resp)
	return verifiedProviderPayment{Provider: pgateway.GatewayPayPal, OrderNumber: orderNumber, TransactionID: transactionID, PaymentMethod: "paypal", Amount: providerAmount, GatewayResponse: string(gatewayResponse), LiabilityShifted: paymentResponseLiabilityShifted(resp, gatewayResponse)}, nil
}

func (h *Handler) recordVerifiedGatewayRefund(c *gin.Context, refund service.VerifiedGatewayRefundInput) bool {
	if h == nil || h.paymentService == nil {
		apierror.RespondInternalError(c, errors.New("payment service is unavailable"))
		return false
	}
	if err := h.paymentService.RecordVerifiedGatewayRefund(refund); err != nil {
		respondVerifiedGatewayRefundError(c, err)
		return false
	}
	return true
}

func respondVerifiedGatewayRefundError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrOrderNotFound) {
		apierror.RespondNotFound(c, "Order")
		return
	}
	apierror.RespondBadRequest(c, err.Error())
}

func (h *Handler) handleAlipayWebhook(c *gin.Context, payload []byte) {
	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayAlipay)
	if err != nil {
		respondAlipayWebhookFailure(c, http.StatusInternalServerError)
		return
	}

	notification, err := pgateway.VerifyAlipayWebhook(c.Request.Context(), config, payload)
	if err != nil {
		respondAlipayWebhookFailure(c, http.StatusUnauthorized)
		return
	}
	if err := pgateway.ValidateAlipayWebhookMerchantIdentity(config, notification); err != nil {
		respondAlipayWebhookFailure(c, http.StatusUnauthorized)
		return
	}
	if alipayRefundNotificationPresent(notification) {
		refund, err := alipayVerifiedRefundFromNotification(notification, payload)
		if err != nil {
			respondAlipayWebhookFailure(c, http.StatusBadRequest)
			return
		}
		if strings.EqualFold(strings.TrimSpace(notification.RefundStatus), "REFUND_SUCCESS") {
			if err := h.paymentService.RecordVerifiedGatewayRefund(refund); err != nil {
				respondAlipayWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err))
				return
			}
		} else {
			refund.ErrorMessage = fmt.Sprintf("alipay refund status: %s", strings.TrimSpace(notification.RefundStatus))
			if err := h.paymentService.RecordGatewayRefundFailure(refund); err != nil {
				respondAlipayWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err))
				return
			}
		}
		respondAlipayWebhookSuccess(c)
		return
	}

	status := strings.ToUpper(strings.TrimSpace(notification.TradeStatus))
	if status != alipayTradeStatusSuccess && status != alipayTradeStatusFinished {
		respondAlipayWebhookSuccess(c)
		return
	}

	currencyCode := firstNonBlank(notification.Currency, "CNY")
	amount, err := domainmoney.ParseMajor(notification.TotalAmount, currencyCode)
	if err != nil {
		respondAlipayWebhookFailure(c, http.StatusBadRequest)
		return
	}
	transactionID := strings.TrimSpace(notification.TradeNo)
	if transactionID == "" {
		respondAlipayWebhookFailure(c, http.StatusBadRequest)
		return
	}

	payment := verifiedProviderPayment{
		Provider:        pgateway.GatewayAlipay,
		OrderNumber:     strings.TrimSpace(notification.OutTradeNo),
		TransactionID:   transactionID,
		PaymentMethod:   "alipay",
		Amount:          amount,
		GatewayResponse: string(payload),
	}
	result, err := h.recordVerifiedProviderPaymentResult(payment)
	if err != nil {
		respondAlipayWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err))
		return
	}
	_ = result
	respondAlipayWebhookSuccess(c)
}

func (h *Handler) handleWechatWebhook(c *gin.Context, payload []byte) {
	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayWechat)
	if err != nil {
		respondWechatWebhookFailure(c, http.StatusInternalServerError, "wechat config error")
		return
	}

	var envelope struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		respondWechatWebhookFailure(c, http.StatusBadRequest, "invalid wechat webhook envelope")
		return
	}
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(envelope.EventType)), "REFUND.") {
		h.handleWechatRefundWebhook(c, config, payload)
		return
	}

	verified, err := pgateway.VerifyWechatWebhook(c.Request.Context(), config, c.Request.Header, payload)
	if err != nil {
		respondWechatWebhookFailure(c, http.StatusUnauthorized, "wechat signature verification failed")
		return
	}
	if err := pgateway.ValidateWechatWebhookMerchantIdentity(config, verified.Transaction); err != nil {
		respondWechatWebhookFailure(c, http.StatusUnauthorized, "wechat merchant identity verification failed")
		return
	}

	transaction := verified.Transaction
	if !strings.EqualFold(strings.TrimSpace(transaction.TradeState), wechatTradeStateSuccess) {
		respondWechatWebhookSuccess(c)
		return
	}

	currency := strings.TrimSpace(transaction.Amount.Currency)
	if currency == "" {
		currency = "CNY"
	}
	amount, err := domainmoney.New(transaction.Amount.Total, currency)
	if err != nil {
		respondWechatWebhookFailure(c, http.StatusBadRequest, "invalid payment amount")
		return
	}
	transactionID := strings.TrimSpace(transaction.TransactionID)
	if transactionID == "" {
		respondWechatWebhookFailure(c, http.StatusBadRequest, "wechat transaction_id is required")
		return
	}

	payment := verifiedProviderPayment{
		Provider:        pgateway.GatewayWechat,
		OrderNumber:     strings.TrimSpace(transaction.OutTradeNo),
		TransactionID:   transactionID,
		PaymentMethod:   "wechat",
		Amount:          amount,
		GatewayResponse: string(payload),
	}
	if verified.Plaintext != "" {
		payment.GatewayResponse = verified.Plaintext
	}
	result, err := h.recordVerifiedProviderPaymentResult(payment)
	if err != nil {
		respondWechatWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err), err.Error())
		return
	}
	_ = result
	respondWechatWebhookSuccess(c)
}

func (h *Handler) handleWechatRefundWebhook(c *gin.Context, config *pgateway.Config, payload []byte) {
	verified, err := pgateway.VerifyWechatRefundWebhook(c.Request.Context(), config, c.Request.Header, payload)
	if err != nil {
		respondWechatWebhookFailure(c, http.StatusUnauthorized, "wechat refund webhook verification failed")
		return
	}
	refund := verified.Refund
	providerRefundID := firstNonBlank(refund.RefundID, refund.OutRefundNo, verified.ID)
	transactionID := strings.TrimSpace(refund.TransactionID)
	if transactionID == "" || providerRefundID == "" {
		respondWechatWebhookFailure(c, http.StatusBadRequest, "wechat refund notification is missing identifiers")
		return
	}
	currencyCode := strings.TrimSpace(refund.Amount.Currency)
	if currencyCode == "" {
		currencyCode = "CNY"
	}
	minorAmount := refund.Amount.Refund
	if minorAmount <= 0 {
		minorAmount = refund.Amount.PayerRefund
	}
	amount, amountErr := domainmoney.New(0, currencyCode)
	if amountErr != nil {
		respondWechatWebhookFailure(c, http.StatusBadRequest, "invalid wechat refund currency")
		return
	}
	if minorAmount > 0 {
		amount, err = domainmoney.New(minorAmount, currencyCode)
		if err != nil {
			respondWechatWebhookFailure(c, http.StatusBadRequest, "invalid wechat refund amount")
			return
		}
	}
	input := service.VerifiedGatewayRefundInput{
		Provider:             string(pgateway.GatewayWechat),
		OrderNumber:          strings.TrimSpace(refund.OutTradeNo),
		TransactionID:        transactionID,
		RefundID:             providerRefundID,
		ProviderStatus:       strings.TrimSpace(refund.RefundStatus),
		ProviderRefundAmount: amount,
		GatewayResponse:      verified.Plaintext,
	}
	if strings.EqualFold(strings.TrimSpace(refund.RefundStatus), "SUCCESS") {
		if err := h.paymentService.RecordVerifiedGatewayRefund(input); err != nil {
			respondWechatWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err), err.Error())
			return
		}
	} else {
		input.ErrorMessage = fmt.Sprintf("wechat refund status: %s", strings.TrimSpace(refund.RefundStatus))
		if err := h.paymentService.RecordGatewayRefundFailure(input); err != nil {
			respondWechatWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err), err.Error())
			return
		}
	}
	respondWechatWebhookSuccess(c)
}

func (h *Handler) recordVerifiedProviderPayment(c *gin.Context, payment verifiedProviderPayment) (bool, service.VerifiedGatewayPaymentResult) {
	result, err := h.recordVerifiedProviderPaymentResult(payment)
	if err != nil {
		respondVerifiedProviderPaymentError(c, err)
		return false, service.VerifiedGatewayPaymentResult{}
	}
	return true, result
}

func (h *Handler) recordVerifiedProviderPaymentResult(payment verifiedProviderPayment) (service.VerifiedGatewayPaymentResult, error) {
	if strings.TrimSpace(payment.OrderNumber) == "" {
		return service.VerifiedGatewayPaymentResult{}, errors.New("order_number is required")
	}
	if strings.TrimSpace(payment.TransactionID) == "" {
		return service.VerifiedGatewayPaymentResult{}, errors.New("transaction_id is required")
	}
	if err := payment.Amount.Validate(); err != nil {
		return service.VerifiedGatewayPaymentResult{}, fmt.Errorf("invalid provider payment amount: %w", err)
	}
	if payment.Amount.AmountMinor() <= 0 {
		return service.VerifiedGatewayPaymentResult{}, errors.New("invalid provider payment amount: must be greater than zero")
	}
	return h.paymentService.RecordVerifiedGatewayPaymentResult(service.VerifiedGatewayPaymentInput{
		Provider:         string(payment.Provider),
		OrderNumber:      payment.OrderNumber,
		TransactionID:    payment.TransactionID,
		PaymentMethod:    payment.PaymentMethod,
		Amount:           payment.Amount,
		GatewayResponse:  payment.GatewayResponse,
		LiabilityShifted: payment.LiabilityShifted,
	})
}

func respondVerifiedProviderPaymentError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrOrderNotFound) {
		apierror.RespondNotFound(c, "Order")
		return
	}
	apierror.RespondBadRequest(c, err.Error())
}

func verifiedProviderPaymentWebhookStatus(err error) int {
	if errors.Is(err, service.ErrOrderNotFound) {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func respondAlipayWebhookSuccess(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, "success")
}

func respondAlipayWebhookFailure(c *gin.Context, status int) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(status, "fail")
}

func respondWechatWebhookSuccess(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

func respondWechatWebhookFailure(c *gin.Context, status int, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "wechat webhook processing failed"
	}
	c.JSON(status, gin.H{
		"code":    "FAIL",
		"message": message,
	})
}

func paypalVerifiedRefundFromEvent(event pgateway.PayPalWebhookEvent, rawPayload []byte) (service.VerifiedGatewayRefundInput, bool, error) {
	if !strings.EqualFold(strings.TrimSpace(event.EventType), paypalPaymentCaptureRefunded) {
		return service.VerifiedGatewayRefundInput{}, false, nil
	}

	var refund paypalRefundWebhookResource
	if err := json.Unmarshal(event.Resource, &refund); err != nil {
		return service.VerifiedGatewayRefundInput{}, true, fmt.Errorf("invalid paypal refund resource: %w", err)
	}
	status := firstNonBlank(refund.Status, refund.State)
	if !paypalRefundStatusRecordable(status) {
		return service.VerifiedGatewayRefundInput{}, false, nil
	}

	refundID := strings.TrimSpace(refund.ID)
	if refundID == "" {
		return service.VerifiedGatewayRefundInput{}, true, errors.New("paypal refund resource does not contain a refund id")
	}
	captureID := paypalRefundCaptureID(refund)
	if captureID == "" {
		return service.VerifiedGatewayRefundInput{}, true, errors.New("paypal refund resource does not contain a capture id")
	}
	if refund.Amount == nil {
		return service.VerifiedGatewayRefundInput{}, true, errors.New("paypal refund resource does not contain amount")
	}

	currency := firstNonBlank(refund.Amount.CurrencyCode, refund.Amount.Currency)
	value := firstNonBlank(refund.Amount.Value, refund.Amount.Total)
	if strings.TrimSpace(currency) == "" {
		return service.VerifiedGatewayRefundInput{}, true, errors.New("paypal refund resource does not contain currency")
	}
	amount, err := domainmoney.ParseMajor(value, currency)
	if err != nil {
		return service.VerifiedGatewayRefundInput{}, true, err
	}
	if amount.AmountMinor() <= 0 {
		return service.VerifiedGatewayRefundInput{}, true, errors.New("paypal refund resource does not contain a positive amount")
	}

	return service.VerifiedGatewayRefundInput{
		Provider:             string(pgateway.GatewayPayPal),
		OrderNumber:          firstNonBlank(refund.CustomID, refund.InvoiceID),
		TransactionID:        captureID,
		RefundID:             refundID,
		ProviderStatus:       status,
		ProviderRefundAmount: amount,
		GatewayResponse:      string(rawPayload),
	}, true, nil
}

func alipayRefundNotificationPresent(notification pgateway.AlipayWebhookNotification) bool {
	return strings.TrimSpace(notification.RefundStatus) != "" || strings.TrimSpace(notification.OutRequestNo) != ""
}

func alipayVerifiedRefundFromNotification(notification pgateway.AlipayWebhookNotification, rawPayload []byte) (service.VerifiedGatewayRefundInput, error) {
	transactionID := strings.TrimSpace(notification.TradeNo)
	if transactionID == "" {
		return service.VerifiedGatewayRefundInput{}, errors.New("alipay refund notification does not contain trade_no")
	}
	refundID := firstNonBlank(notification.OutRequestNo, notification.NotifyID, notification.TradeNo)
	currencyCode := strings.TrimSpace(notification.Currency)
	if currencyCode == "" {
		currencyCode = "CNY"
	}
	amount, err := domainmoney.New(0, currencyCode)
	if err != nil {
		return service.VerifiedGatewayRefundInput{}, err
	}
	value := firstNonBlank(notification.RefundAmount, notification.RefundFee)
	if strings.TrimSpace(value) != "" {
		amount, err = domainmoney.ParseMajor(value, currencyCode)
		if err != nil {
			return service.VerifiedGatewayRefundInput{}, err
		}
	}
	return service.VerifiedGatewayRefundInput{
		Provider:             string(pgateway.GatewayAlipay),
		OrderNumber:          strings.TrimSpace(notification.OutTradeNo),
		TransactionID:        transactionID,
		RefundID:             refundID,
		ProviderStatus:       strings.TrimSpace(notification.RefundStatus),
		ProviderRefundAmount: amount,
		GatewayResponse:      string(rawPayload),
	}, nil
}

func paypalRefundStatusRecordable(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "", "COMPLETED", "REFUNDED", "SUCCEEDED":
		return true
	default:
		return false
	}
}

func paypalRefundCaptureID(refund paypalRefundWebhookResource) string {
	if captureID := strings.TrimSpace(refund.CaptureID); captureID != "" {
		return captureID
	}
	for _, link := range refund.Links {
		captureID := paypalCaptureIDFromURL(link.Href)
		if captureID != "" {
			return captureID
		}
	}
	return ""
}

func paypalCaptureIDFromURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for index := 0; index+1 < len(segments); index++ {
		if strings.EqualFold(segments[index], "captures") {
			return strings.TrimSpace(segments[index+1])
		}
	}
	return ""
}

func paypalVerifiedPaymentFromEvent(event pgateway.PayPalWebhookEvent, rawPayload []byte) (verifiedProviderPayment, bool, error) {
	eventType := strings.TrimSpace(event.EventType)
	if strings.EqualFold(eventType, paypalPaymentCaptureCompleted) {
		return paypalVerifiedCapturePaymentFromEvent(event, rawPayload)
	}
	if !strings.EqualFold(eventType, paypalCheckoutOrderCompleted) {
		return verifiedProviderPayment{}, false, nil
	}

	var order paypal.Order
	if err := json.Unmarshal(event.Resource, &order); err != nil {
		return verifiedProviderPayment{}, true, fmt.Errorf("invalid paypal order resource: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(order.Status), "COMPLETED") {
		return verifiedProviderPayment{}, false, nil
	}

	payment := verifiedProviderPayment{
		Provider:         pgateway.GatewayPayPal,
		PaymentMethod:    "paypal",
		GatewayResponse:  string(rawPayload),
		LiabilityShifted: paypalLiabilityShiftedFromWebhookPayload(event.Resource, rawPayload),
	}
	foundCompletedCapture := false
	for _, unit := range order.PurchaseUnits {
		if payment.OrderNumber == "" {
			payment.OrderNumber = firstNonBlank(unit.CustomID, unit.InvoiceID, unit.ReferenceID)
		}
		if unit.Payments != nil {
			for _, capture := range unit.Payments.Captures {
				if !strings.EqualFold(strings.TrimSpace(capture.Status), "COMPLETED") {
					continue
				}
				foundCompletedCapture = true
				payment.TransactionID = strings.TrimSpace(capture.ID)
				if payment.TransactionID == "" {
					return verifiedProviderPayment{}, true, fmt.Errorf("paypal completed capture does not contain a capture id")
				}
				if payment.OrderNumber == "" {
					payment.OrderNumber = strings.TrimSpace(capture.CustomID)
				}
				if capture.Amount == nil {
					return verifiedProviderPayment{}, true, fmt.Errorf("paypal completed capture does not contain amount")
				}
				amount, err := domainmoney.ParseMajor(capture.Amount.Value, capture.Amount.Currency)
				if err != nil {
					return verifiedProviderPayment{}, true, err
				}
				payment.Amount = amount
				break
			}
		}
		if payment.OrderNumber != "" && payment.TransactionID != "" && payment.Amount.AmountMinor() > 0 {
			break
		}
	}

	if !foundCompletedCapture {
		return verifiedProviderPayment{}, true, fmt.Errorf("paypal order resource does not contain a completed capture")
	}
	if payment.OrderNumber == "" {
		return verifiedProviderPayment{}, true, fmt.Errorf("paypal order resource does not contain order metadata")
	}
	if payment.TransactionID == "" {
		return verifiedProviderPayment{}, true, fmt.Errorf("paypal order resource does not contain a transaction id")
	}
	if payment.Amount.AmountMinor() <= 0 {
		return verifiedProviderPayment{}, true, fmt.Errorf("paypal order resource does not contain a positive amount")
	}
	return payment, true, nil
}

func paypalVerifiedCapturePaymentFromEvent(event pgateway.PayPalWebhookEvent, rawPayload []byte) (verifiedProviderPayment, bool, error) {
	var resource struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount *struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"amount"`
		CustomID          string `json:"custom_id"`
		InvoiceID         string `json:"invoice_id"`
		SupplementaryData struct {
			RelatedIDs struct {
				OrderID string `json:"order_id"`
			} `json:"related_ids"`
		} `json:"supplementary_data"`
	}
	if err := json.Unmarshal(event.Resource, &resource); err != nil {
		return verifiedProviderPayment{}, true, fmt.Errorf("invalid paypal capture resource: %w", err)
	}
	if !strings.EqualFold(resource.Status, "COMPLETED") {
		return verifiedProviderPayment{}, false, nil
	}
	if resource.ID == "" || resource.Amount == nil {
		return verifiedProviderPayment{}, true, errors.New("paypal completed capture resource is incomplete")
	}
	amount, err := domainmoney.ParseMajor(resource.Amount.Value, resource.Amount.CurrencyCode)
	if err != nil {
		return verifiedProviderPayment{}, true, err
	}
	return verifiedProviderPayment{
		Provider: pgateway.GatewayPayPal, TransactionID: resource.ID,
		OrderNumber:   firstNonBlank(resource.CustomID, resource.InvoiceID, resource.SupplementaryData.RelatedIDs.OrderID),
		PaymentMethod: "paypal", Amount: amount, GatewayResponse: string(rawPayload),
		LiabilityShifted: paypalLiabilityShiftedFromWebhookPayload(event.Resource, rawPayload),
	}, true, nil
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
