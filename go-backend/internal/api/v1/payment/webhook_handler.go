package payment

import (
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
		setStripeWebhookSuccess(c, "Ignored Stripe payment_intent.succeeded without order metadata", gin.H{
			"event_id":          event.ID,
			"payment_intent_id": intent.ID,
		})
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
	majorAmount, err := pgateway.MinorToMajorAmount(amount, string(intent.Currency))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:         string(pgateway.GatewayStripe),
		OrderNumber:      orderNumber,
		TransactionID:    intent.ID,
		PaymentMethod:    "stripe",
		Amount:           majorAmount,
		Currency:         string(intent.Currency),
		GatewayResponse:  string(payload),
		LiabilityShifted: stripeLiabilityShiftedFromPaymentIntentSucceeded(event.Data.Raw, payload),
	}); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			apierror.RespondNotFound(c, "Order")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
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

	setStripeWebhookSuccess(c, "Stripe webhook processed successfully", gin.H{
		"event_id":          event.ID,
		"event_type":        string(event.Type),
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
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
		majorAmount, err := pgateway.MinorToMajorAmount(intent.Amount, string(intent.Currency))
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
			Amount:        majorAmount,
			Currency:      string(intent.Currency),
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
		majorAmount, err := pgateway.MinorToMajorAmount(intent.Amount, string(intent.Currency))
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
			Amount:        majorAmount,
			Currency:      string(intent.Currency),
		})
	}
	setStripeWebhookSuccess(c, "Stripe payment processing recorded", gin.H{
		"event_id":          event.ID,
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
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

	majorAmount, err := pgateway.MinorToMajorAmount(dispute.Amount, string(dispute.Currency))
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

func stripeOrderNumber(metadata map[string]string) string {
	for _, key := range []string{"order_number", "order_id", "order"} {
		if value := metadata[key]; value != "" {
			return value
		}
	}
	return ""
}
