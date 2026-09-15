package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment" // alias for gateway
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
)

const paymentWebhookMaxBodyBytes = 1 << 20

// ============ Webhook 相关接口 ============

// HandleWebhook 处理外部支付服务的回调通知
// @Summary 处理外部支付回调
// @Tags Payment
// @Accept json
// @Produce json
// @Param provider path string true "支付渠道 (如: stripe, alipay)"
// @Router /api/v1/payment/webhook/{provider} [post]
func (h *Handler) HandleWebhook(c *gin.Context) {
	provider := c.Param("provider")

	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, paymentWebhookMaxBodyBytes+1))
	if err != nil {
		apierror.RespondBadRequest(c, "Failed to read request body")
		return
	}
	if len(payload) > paymentWebhookMaxBodyBytes {
		apierror.RespondError(c, http.StatusRequestEntityTooLarge, "payment_webhook_payload_too_large", "Payment webhook payload is too large")
		return
	}

	switch provider {
	case "stripe":
		h.handleStripeWebhook(c, payload)
	case "paypal":
		h.handlePayPalWebhook(c, payload)
	case "alipay":
		h.handleAlipayWebhook(c, payload)
	case "wechat":
		h.handleWechatWebhook(c, payload)
	default:
		apierror.RespondBadRequest(c, "Unsupported payment provider")
	}
}

func (h *Handler) handleStripeWebhook(c *gin.Context, payload []byte) {
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		apierror.RespondUnauthorized(c)
		return
	}

	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayStripe)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if config.WebhookSecret == "" {
		apierror.RespondInternalError(c, fmt.Errorf("payment webhook is not configured"))
		return
	}

	event, err := pgateway.ParseStripeWebhookEvent(payload, signature, config.WebhookSecret)
	if err != nil {
		apierror.RespondUnauthorized(c)
		return
	}

	claimed, err := h.paymentService.ClaimStripeWebhookEvent(event.ID, string(event.Type), string(payload))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !claimed {
		response.SuccessWithMessage(c, "Stripe event already handled", gin.H{
			"event_id":   event.ID,
			"event_type": string(event.Type),
		})
		return
	}

	switch string(event.Type) {
	case "payment_intent.succeeded":
		h.handleStripePaymentIntentSucceeded(c, event, payload)
	case "payment_intent.payment_failed":
		h.handleStripePaymentIntentFailed(c, event)
	case "payment_intent.requires_action":
		h.handleStripePaymentIntentRequiresAction(c, event)
	case "payment_intent.processing":
		h.handleStripePaymentIntentProcessing(c, event)
	case "charge.refunded":
		h.handleStripeChargeRefunded(c, event, payload)
	case "refund.created":
		h.handleStripeRefundCreated(c, event, payload)
	case "review.opened":
		h.handleStripeReviewOpened(c, event)
	case "review.closed":
		h.handleStripeReviewClosed(c, event)
	case "radar.early_fraud_warning.created", "radar.early_fraud_warning.updated":
		h.handleStripeEarlyFraudWarning(c, event, payload)
	case "charge.dispute.created", "charge.dispute.updated",
		"charge.dispute.funds_withdrawn", "charge.dispute.funds_reinstated",
		"charge.dispute.closed":
		h.handleStripeDispute(c, event, payload)
	default:
		setStripeWebhookSuccess(c, "Ignored unsupported Stripe event", gin.H{
			"event_id":   event.ID,
			"event_type": string(event.Type),
		})
	}

	if c.Writer.Status() >= http.StatusBadRequest {
		_ = h.paymentService.MarkStripeWebhookEventFailed(event.ID, fmt.Errorf("stripe event handler returned HTTP %d", c.Writer.Status()))
		return
	}
	if err := h.paymentService.MarkStripeWebhookEventProcessed(event.ID); err != nil {
		// Stripe will retry if the inbox cannot be finalized.
		apierror.RespondInternalError(c, err)
		return
	}
	writeStripeWebhookSuccess(c)
}

type stripeWebhookSuccessResponse struct {
	Message string
	Data    gin.H
}

const stripeWebhookSuccessKey = "stripe_webhook_success_response"

func setStripeWebhookSuccess(c *gin.Context, message string, data gin.H) {
	c.Set(stripeWebhookSuccessKey, stripeWebhookSuccessResponse{
		Message: message,
		Data:    data,
	})
}

func writeStripeWebhookSuccess(c *gin.Context) {
	value, exists := c.Get(stripeWebhookSuccessKey)
	if !exists {
		response.SuccessWithMessage(c, "Stripe event processed", nil)
		return
	}
	payload, ok := value.(stripeWebhookSuccessResponse)
	if !ok {
		response.SuccessWithMessage(c, "Stripe event processed", nil)
		return
	}
	response.SuccessWithMessage(c, payload.Message, payload.Data)
}

func (h *Handler) handleStripePaymentIntentFailed(c *gin.Context, event stripe.Event) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	orderNumber := stripeOrderNumber(intent.Metadata)
	if orderNumber != "" {
		_ = h.paymentService.RecordGatewayPaymentFailure(
			c.Request.Context(),
			string(pgateway.GatewayStripe),
			orderNumber,
			intent.ID,
		)
	}
	if h.antiFraud != nil {
		_ = h.antiFraud.RecordPaymentIntentFailure(c.Request.Context(), intent.ID)
	}
	if h.cardBINLimiter != nil {
		if _, err := h.cardBINLimiter.RecordPaymentIntentFailure(c.Request.Context(), intent.ID); err != nil {
			apierror.RespondError(c, http.StatusServiceUnavailable, "payment_bin_risk_unavailable", "Payment card risk service is temporarily unavailable")
			return
		}
	}

	setStripeWebhookSuccess(c, "Stripe payment failure recorded", gin.H{
		"event_id":          event.ID,
		"event_type":        string(event.Type),
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
}

func (h *Handler) handleStripePaymentIntentSucceeded(c *gin.Context, event stripe.Event, payload []byte) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	orderNumber := stripeOrderNumber(intent.Metadata)
	if orderNumber == "" {
		apierror.RespondBadRequest(c, "Stripe payment_intent.succeeded is missing order_number metadata")
		return
	}

	amount := intent.AmountReceived
	if amount <= 0 {
		amount = intent.Amount
	}
	if amount <= 0 {
		setStripeWebhookSuccess(c, "Ignored Stripe payment_intent.succeeded without amount", gin.H{
			"event_id":          event.ID,
			"payment_intent_id": intent.ID,
		})
		return
	}
	verifiedAmount, err := domainmoney.New(amount, string(intent.Currency))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	result, err := h.paymentService.RecordVerifiedGatewayPaymentResult(service.VerifiedGatewayPaymentInput{
		Provider:         string(pgateway.GatewayStripe),
		OrderNumber:      orderNumber,
		TransactionID:    intent.ID,
		PaymentMethod:    "stripe",
		Amount:           verifiedAmount,
		GatewayResponse:  string(payload),
		LiabilityShifted: stripeLiabilityShiftedFromPaymentIntentSucceeded(event.Data.Raw, payload),
	})
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			apierror.RespondNotFound(c, "Order")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := h.paymentService.ResolveStripeRequiresActionReview(intent.ID); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if h.antiFraud != nil {
		_ = h.antiFraud.RecordPaymentIntentSuccess(c.Request.Context(), intent.ID)
	}
	if h.cardBINLimiter != nil {
		if err := h.cardBINLimiter.RecordPaymentIntentSuccess(c.Request.Context(), intent.ID); err != nil {
			apierror.RespondError(c, http.StatusServiceUnavailable, "payment_bin_risk_unavailable", "Payment card risk service is temporarily unavailable")
			return
		}
	}

	message := "Stripe webhook processed successfully"
	data := gin.H{
		"event_id":          event.ID,
		"event_type":        string(event.Type),
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	}
	if result.DuplicatePaid {
		message = "Stripe duplicate payment recorded; refund is pending"
		data["duplicate_paid"] = true
		data["refund_id"] = result.RefundID
	}
	setStripeWebhookSuccess(c, message, data)
}

func (h *Handler) handleStripePaymentIntentRequiresAction(c *gin.Context, event stripe.Event) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}
	orderNumber := stripeOrderNumber(intent.Metadata)
	var orderID *uint
	if orderNumber != "" {
		amount, err := domainmoney.New(intent.Amount, string(intent.Currency))
		if err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:      string(pgateway.GatewayStripe),
			OrderNumber:   orderNumber,
			TransactionID: intent.ID,
			PaymentMethod: "stripe",
			Status:        "requires_action",
			Amount:        amount,
		})
		if record, err := h.orderService.GetOrderByNumberForPayment(orderNumber); err == nil {
			orderID = &record.ID
		}
	}
	_, err := h.paymentService.CreatePaymentReview(service.CreatePaymentReviewInput{
		OrderID:         orderID,
		PaymentIntentID: intent.ID,
		Status:          "pending",
		Reason:          "stripe_requires_action",
		Source:          "radar",
		Notes:           "Stripe requested customer authentication or additional payment action.",
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	setStripeWebhookSuccess(c, "Stripe payment action recorded", gin.H{
		"event_id":          event.ID,
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
}

func (h *Handler) handleStripePaymentIntentProcessing(c *gin.Context, event stripe.Event) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}
	orderNumber := stripeOrderNumber(intent.Metadata)
	if orderNumber != "" {
		amount, err := domainmoney.New(intent.Amount, string(intent.Currency))
		if err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:      string(pgateway.GatewayStripe),
			OrderNumber:   orderNumber,
			TransactionID: intent.ID,
			PaymentMethod: "stripe",
			Status:        "processing",
			Amount:        amount,
		})
	}
	setStripeWebhookSuccess(c, "Stripe payment processing recorded", gin.H{
		"event_id":          event.ID,
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
}

func (h *Handler) handleStripeRefundCreated(c *gin.Context, event stripe.Event, payload []byte) {
	var refund stripe.Refund
	if err := json.Unmarshal(event.Data.Raw, &refund); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}
	if !stripeRefundStatusRecordable(refund.Status) {
		setStripeWebhookSuccess(c, "Ignored Stripe refund event with terminal failure status", gin.H{
			"event_id":  event.ID,
			"refund_id": refund.ID,
			"status":    string(refund.Status),
		})
		return
	}

	input, err := stripeVerifiedRefundInput(&refund, nil, payload)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !h.recordVerifiedGatewayRefund(c, input) {
		return
	}

	setStripeWebhookSuccess(c, "Stripe refund webhook processed successfully", gin.H{
		"event_id":        event.ID,
		"refund_id":       input.RefundID,
		"transaction_id":  input.TransactionID,
		"refund_amount":   providerRefundAmountMajor(input.ProviderRefundAmount),
		"refund_currency": input.ProviderRefundAmount.Currency().String(),
	})
}

func (h *Handler) handleStripeChargeRefunded(c *gin.Context, event stripe.Event, payload []byte) {
	var charge stripe.Charge
	if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	inputs := make([]service.VerifiedGatewayRefundInput, 0)
	if charge.Refunds != nil {
		for _, refund := range charge.Refunds.Data {
			if refund == nil || !stripeRefundStatusRecordable(refund.Status) {
				continue
			}
			input, err := stripeVerifiedRefundInput(refund, &charge, payload)
			if err != nil {
				apierror.RespondBadRequest(c, err.Error())
				return
			}
			inputs = append(inputs, input)
		}
	}

	if len(inputs) == 0 {
		if charge.AmountRefunded <= 0 {
			setStripeWebhookSuccess(c, "Ignored Stripe charge.refunded without refund details", gin.H{
				"event_id":  event.ID,
				"charge_id": charge.ID,
			})
			return
		}

		transactionID := stripeChargePaymentTransactionID(&charge)
		if transactionID == "" {
			apierror.RespondBadRequest(c, "Stripe refunded charge does not contain a payment intent")
			return
		}
		amount, err := domainmoney.New(charge.AmountRefunded, string(charge.Currency))
		if err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		inputs = append(inputs, service.VerifiedGatewayRefundInput{
			Provider:             string(pgateway.GatewayStripe),
			OrderNumber:          stripeOrderNumber(charge.Metadata),
			TransactionID:        transactionID,
			RefundID:             event.ID,
			ProviderStatus:       "succeeded",
			ProviderRefundAmount: amount,
			GatewayResponse:      string(payload),
		})
	}

	refundIDs := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if !h.recordVerifiedGatewayRefund(c, input) {
			return
		}
		refundIDs = append(refundIDs, input.RefundID)
	}

	setStripeWebhookSuccess(c, "Stripe charge refund webhook processed successfully", gin.H{
		"event_id":   event.ID,
		"charge_id":  charge.ID,
		"refund_ids": refundIDs,
	})
}

func stripeVerifiedRefundInput(refund *stripe.Refund, fallbackCharge *stripe.Charge, payload []byte) (service.VerifiedGatewayRefundInput, error) {
	if refund == nil {
		return service.VerifiedGatewayRefundInput{}, errors.New("Stripe refund resource is required")
	}
	refundID := strings.TrimSpace(refund.ID)
	if refundID == "" {
		return service.VerifiedGatewayRefundInput{}, errors.New("Stripe refund resource does not contain a refund id")
	}
	if refund.Amount <= 0 {
		return service.VerifiedGatewayRefundInput{}, errors.New("Stripe refund resource does not contain a positive amount")
	}
	currency := strings.TrimSpace(string(refund.Currency))
	if currency == "" && refund.Charge != nil {
		currency = strings.TrimSpace(string(refund.Charge.Currency))
	}
	if currency == "" && fallbackCharge != nil {
		currency = strings.TrimSpace(string(fallbackCharge.Currency))
	}
	if currency == "" {
		return service.VerifiedGatewayRefundInput{}, errors.New("Stripe refund resource does not contain currency")
	}
	amount, err := domainmoney.New(refund.Amount, currency)
	if err != nil {
		return service.VerifiedGatewayRefundInput{}, err
	}
	transactionID := stripeRefundTransactionID(refund, fallbackCharge)
	if transactionID == "" {
		return service.VerifiedGatewayRefundInput{}, errors.New("Stripe refund resource does not contain a payment intent")
	}

	orderNumber := stripeOrderNumber(refund.Metadata)
	if orderNumber == "" && refund.PaymentIntent != nil {
		orderNumber = stripeOrderNumber(refund.PaymentIntent.Metadata)
	}
	if orderNumber == "" && refund.Charge != nil {
		orderNumber = stripeOrderNumber(refund.Charge.Metadata)
	}
	if orderNumber == "" && fallbackCharge != nil {
		orderNumber = stripeOrderNumber(fallbackCharge.Metadata)
	}

	return service.VerifiedGatewayRefundInput{
		Provider:             string(pgateway.GatewayStripe),
		OrderNumber:          orderNumber,
		TransactionID:        transactionID,
		RefundID:             refundID,
		ProviderStatus:       strings.TrimSpace(string(refund.Status)),
		ProviderRefundAmount: amount,
		GatewayResponse:      string(payload),
	}, nil
}

func stripeRefundTransactionID(refund *stripe.Refund, fallbackCharge *stripe.Charge) string {
	if refund != nil {
		if refund.PaymentIntent != nil {
			if paymentIntentID := strings.TrimSpace(refund.PaymentIntent.ID); paymentIntentID != "" {
				return paymentIntentID
			}
		}
		if refund.Charge != nil {
			if refund.Charge.PaymentIntent != nil {
				if paymentIntentID := strings.TrimSpace(refund.Charge.PaymentIntent.ID); paymentIntentID != "" {
					return paymentIntentID
				}
			}
		}
	}
	if fallbackCharge != nil {
		if fallbackCharge.PaymentIntent != nil {
			if paymentIntentID := strings.TrimSpace(fallbackCharge.PaymentIntent.ID); paymentIntentID != "" {
				return paymentIntentID
			}
		}
	}
	return ""
}

func stripeChargePaymentTransactionID(charge *stripe.Charge) string {
	if charge == nil {
		return ""
	}
	if charge.PaymentIntent != nil {
		if paymentIntentID := strings.TrimSpace(charge.PaymentIntent.ID); paymentIntentID != "" {
			return paymentIntentID
		}
	}
	for _, key := range []string{"payment_intent_id", "payment_intent"} {
		if value := strings.TrimSpace(charge.Metadata[key]); value != "" {
			return value
		}
	}
	return ""
}

func stripeRefundStatusRecordable(status stripe.RefundStatus) bool {
	switch strings.ToLower(strings.TrimSpace(string(status))) {
	case "failed", "canceled":
		return false
	default:
		return true
	}
}

func (h *Handler) handleStripeReviewOpened(c *gin.Context, event stripe.Event) {
	var payload struct {
		ID            string `json:"id"`
		Reason        string `json:"reason"`
		PaymentIntent string `json:"payment_intent"`
		Charge        string `json:"charge"`
	}
	if err := json.Unmarshal(event.Data.Raw, &payload); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}
	_, err := h.paymentService.CreatePaymentReview(service.CreatePaymentReviewInput{
		PaymentIntentID: payload.PaymentIntent,
		StripeReviewID:  payload.ID,
		Status:          "pending",
		Reason:          "stripe_review_opened",
		Source:          "radar",
		Notes:           fmt.Sprintf("Stripe Radar review %s opened (%s), charge %s.", payload.ID, payload.Reason, payload.Charge),
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	setStripeWebhookSuccess(c, "Stripe Radar review recorded", gin.H{
		"event_id":          event.ID,
		"review_id":         payload.ID,
		"payment_intent_id": payload.PaymentIntent,
	})
}

func (h *Handler) handleStripeReviewClosed(c *gin.Context, event stripe.Event) {
	var payload struct {
		ID            string `json:"id"`
		PaymentIntent string `json:"payment_intent"`
		ClosedReason  string `json:"closed_reason"`
	}
	if err := json.Unmarshal(event.Data.Raw, &payload); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}
	if err := h.paymentService.ResolveStripeReview(payload.ID, payload.PaymentIntent, payload.ClosedReason); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	setStripeWebhookSuccess(c, "Stripe Radar review closure recorded", gin.H{
		"event_id":          event.ID,
		"review_id":         payload.ID,
		"payment_intent_id": payload.PaymentIntent,
		"closed_reason":     payload.ClosedReason,
	})
}

func (h *Handler) handleStripeDispute(c *gin.Context, event stripe.Event, payload []byte) {
	var dispute stripe.Dispute
	if err := json.Unmarshal(event.Data.Raw, &dispute); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	majorAmount, err := webhookMajorAmountFromMinor(dispute.Amount, string(dispute.Currency))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	input := service.StripeDisputeInput{
		StripeDisputeID: dispute.ID,
		Amount:          majorAmount,
		Currency:        string(dispute.Currency),
		Reason:          string(dispute.Reason),
		Status:          string(dispute.Status),
		RawPayload:      string(payload),
	}
	if dispute.Charge != nil {
		input.StripeChargeID = dispute.Charge.ID
	}
	if dispute.PaymentIntent != nil {
		input.PaymentIntentID = dispute.PaymentIntent.ID
	}
	if dispute.EvidenceDetails != nil && dispute.EvidenceDetails.DueBy > 0 {
		due := time.Unix(dispute.EvidenceDetails.DueBy, 0)
		input.EvidenceDueAt = &due
	}
	record, err := h.paymentService.RecordStripeDispute(input)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	riskInput := service.PaymentRiskEventInput{
		Provider:          string(pgateway.GatewayStripe),
		Kind:              paymentdomain.PaymentRiskEventDispute,
		ExternalReference: dispute.ID,
		WebhookEventID:    event.ID,
		ProviderPaymentID: input.PaymentIntentID,
		PaymentIntentID:   input.PaymentIntentID,
		ChargeID:          input.StripeChargeID,
		OrderID:           record.OrderID,
		TransactionID:     record.TransactionID,
		Amount:            majorAmount,
		Currency:          string(dispute.Currency),
		OccurredAt:        stripeRiskOccurredAt(dispute.Created),
		Payload:           string(payload),
		Metadata: map[string]string{
			"reason": string(dispute.Reason),
			"status": string(dispute.Status),
		},
	}
	if err := h.recordAndRefreshPaymentRiskEvent(riskInput); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if err := h.enqueueRefundRecommendation(riskInput); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	setStripeWebhookSuccess(c, "Stripe dispute recorded", gin.H{
		"event_id":          event.ID,
		"dispute_id":        record.ID,
		"stripe_dispute_id": record.StripeDisputeID,
		"status":            record.Status,
	})
}

func webhookMajorAmountFromMinor(amount int64, code string) (float64, error) {
	money, err := domainmoney.New(amount, code)
	if err != nil {
		return 0, err
	}
	return money.MajorFloat()
}

func stripeOrderNumber(metadata map[string]string) string {
	for _, key := range []string{"order_number", "order_id", "order"} {
		if value := metadata[key]; value != "" {
			return value
		}
	}
	return ""
}
