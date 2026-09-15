package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/plutov/paypal/v4"
)

const (
	paypalCheckoutOrderCompleted = "CHECKOUT.ORDER.COMPLETED"
	paypalPaymentCaptureRefunded = "PAYMENT.CAPTURE.REFUNDED"
	alipayTradeStatusSuccess     = "TRADE_SUCCESS"
	alipayTradeStatusFinished    = "TRADE_FINISHED"
	wechatTradeStateSuccess      = "SUCCESS"
)

func providerRefundAmountMajor(value domainmoney.Money) float64 {
	amount, _ := value.MajorFloat()
	return amount
}

type verifiedProviderPayment struct {
	Provider         pgateway.GatewayType
	OrderNumber      string
	TransactionID    string
	PaymentMethod    string
	Amount           float64
	Currency         string
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
		apierror.RespondUnauthorized(c)
		return
	}

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

	amount, err := pgateway.ParsePaymentAmount("alipay total_amount", notification.TotalAmount)
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
		Currency:        notification.Currency,
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
	amount, err := webhookMajorAmountFromMinor(transaction.Amount.Total, currency)
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
		Currency:        currency,
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
	amount, err := domainmoney.FromMajorFloat(payment.Amount, payment.Currency)
	if err != nil {
		return service.VerifiedGatewayPaymentResult{}, fmt.Errorf("invalid provider payment amount: %w", err)
	}
	return h.paymentService.RecordVerifiedGatewayPaymentResult(service.VerifiedGatewayPaymentInput{
		Provider:         string(payment.Provider),
		OrderNumber:      payment.OrderNumber,
		TransactionID:    payment.TransactionID,
		PaymentMethod:    payment.PaymentMethod,
		Amount:           amount,
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
	if !strings.EqualFold(strings.TrimSpace(event.EventType), paypalCheckoutOrderCompleted) {
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
				payment.Currency = capture.Amount.Currency
				amount, err := pgateway.ParsePaymentAmount("paypal capture amount", capture.Amount.Value)
				if err != nil {
					return verifiedProviderPayment{}, true, err
				}
				payment.Amount = amount
				break
			}
		}
		if payment.OrderNumber != "" && payment.TransactionID != "" && payment.Amount > 0 && payment.Currency != "" {
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
	if payment.Amount <= 0 {
		return verifiedProviderPayment{}, true, fmt.Errorf("paypal order resource does not contain a positive amount")
	}
	if strings.TrimSpace(payment.Currency) == "" {
		return verifiedProviderPayment{}, true, fmt.Errorf("paypal order resource does not contain currency")
	}
	return payment, true, nil
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
