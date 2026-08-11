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
	"github.com/plutov/paypal/v4"
)

const (
	paypalCheckoutOrderCompleted = "CHECKOUT.ORDER.COMPLETED"
	alipayTradeStatusSuccess     = "TRADE_SUCCESS"
	alipayTradeStatusFinished    = "TRADE_FINISHED"
	wechatTradeStateSuccess      = "SUCCESS"
)

type verifiedProviderPayment struct {
	Provider        pgateway.GatewayType
	OrderNumber     string
	TransactionID   string
	PaymentMethod   string
	Amount          float64
	Currency        string
	GatewayResponse string
}

func (h *Handler) handlePayPalWebhook(c *gin.Context, payload []byte) {
	config, err := h.loadGatewayConfig(pgateway.GatewayPayPal)
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
		if value, exists := c.Get("paypal_dispute_evidence_submitted"); exists {
			details["evidence_submitted"] = value
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

	if !h.recordVerifiedProviderPayment(c, payment) {
		return
	}
	response.SuccessWithMessage(c, "PayPal webhook processed successfully", gin.H{
		"event_id":       event.ID,
		"event_type":     event.EventType,
		"order_number":   payment.OrderNumber,
		"transaction_id": payment.TransactionID,
	})
}

func (h *Handler) handleAlipayWebhook(c *gin.Context, payload []byte) {
	config, err := h.loadGatewayConfig(pgateway.GatewayAlipay)
	if err != nil {
		respondAlipayWebhookFailure(c, http.StatusInternalServerError)
		return
	}

	notification, err := pgateway.VerifyAlipayWebhook(c.Request.Context(), config, payload)
	if err != nil {
		respondAlipayWebhookFailure(c, http.StatusUnauthorized)
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
	if err := h.recordVerifiedProviderPaymentResult(payment); err != nil {
		respondAlipayWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err))
		return
	}
	respondAlipayWebhookSuccess(c)
}

func (h *Handler) handleWechatWebhook(c *gin.Context, payload []byte) {
	config, err := h.loadGatewayConfig(pgateway.GatewayWechat)
	if err != nil {
		respondWechatWebhookFailure(c, http.StatusInternalServerError, "wechat config error")
		return
	}

	verified, err := pgateway.VerifyWechatWebhook(c.Request.Context(), config, c.Request.Header, payload)
	if err != nil {
		respondWechatWebhookFailure(c, http.StatusUnauthorized, "wechat signature verification failed")
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
	amount, err := pgateway.MinorToMajorAmount(transaction.Amount.Total, currency)
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
	if err := h.recordVerifiedProviderPaymentResult(payment); err != nil {
		respondWechatWebhookFailure(c, verifiedProviderPaymentWebhookStatus(err), err.Error())
		return
	}
	respondWechatWebhookSuccess(c)
}

func (h *Handler) recordVerifiedProviderPayment(c *gin.Context, payment verifiedProviderPayment) bool {
	if err := h.recordVerifiedProviderPaymentResult(payment); err != nil {
		respondVerifiedProviderPaymentError(c, err)
		return false
	}
	return true
}

func (h *Handler) recordVerifiedProviderPaymentResult(payment verifiedProviderPayment) error {
	if strings.TrimSpace(payment.OrderNumber) == "" {
		return errors.New("order_number is required")
	}
	if strings.TrimSpace(payment.TransactionID) == "" {
		return errors.New("transaction_id is required")
	}
	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:        string(payment.Provider),
		OrderNumber:     payment.OrderNumber,
		TransactionID:   payment.TransactionID,
		PaymentMethod:   payment.PaymentMethod,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		GatewayResponse: payment.GatewayResponse,
	}); err != nil {
		return err
	}
	return nil
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
		Provider:        pgateway.GatewayPayPal,
		PaymentMethod:   "paypal",
		GatewayResponse: string(rawPayload),
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
