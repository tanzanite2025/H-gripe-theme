package payment

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
)

func (h *Handler) handleStripeEarlyFraudWarning(c *gin.Context, event stripe.Event, payload []byte) {
	var warning stripe.RadarEarlyFraudWarning
	if err := json.Unmarshal(event.Data.Raw, &warning); err != nil {
		apierror.RespondBadRequest(c, "Invalid Stripe early fraud warning payload")
		return
	}
	if strings.TrimSpace(warning.ID) == "" {
		apierror.RespondBadRequest(c, "Stripe early fraud warning id is required")
		return
	}

	paymentIntentID := ""
	if warning.PaymentIntent != nil {
		paymentIntentID = strings.TrimSpace(warning.PaymentIntent.ID)
	}
	chargeID := ""
	if warning.Charge != nil {
		chargeID = strings.TrimSpace(warning.Charge.ID)
	}
	amount, currency := stripeEarlyFraudWarningAmount(warning)

	riskInput := service.PaymentRiskEventInput{
		Provider:          string(pgateway.GatewayStripe),
		Kind:              paymentdomain.PaymentRiskEventEarlyFraudWarning,
		ExternalReference: warning.ID,
		WebhookEventID:    event.ID,
		ProviderPaymentID: paymentIntentID,
		PaymentIntentID:   paymentIntentID,
		ChargeID:          chargeID,
		Amount:            amount,
		Currency:          currency,
		OccurredAt:        stripeRiskOccurredAt(warning.Created),
		Payload:           string(payload),
		Metadata: map[string]string{
			"actionable": fmt.Sprintf("%t", warning.Actionable),
			"fraud_type": string(warning.FraudType),
			"livemode":   fmt.Sprintf("%t", warning.Livemode),
			"event_type": string(event.Type),
		},
	}
	if err := h.recordAndRefreshPaymentRiskEvent(c, riskInput); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if err := h.enqueueRefundRecommendation(riskInput); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if paymentIntentID != "" {
		_, err := h.paymentService.CreatePaymentReview(service.CreatePaymentReviewInput{
			PaymentIntentID: paymentIntentID,
			Status:          "pending",
			Reason:          "stripe_early_fraud_warning",
			Source:          "radar",
			Notes: fmt.Sprintf(
				"Stripe early fraud warning %s received (actionable=%t, fraud_type=%s). No automatic refund was issued.",
				warning.ID,
				warning.Actionable,
				warning.FraudType,
			),
		})
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
	}

	setStripeWebhookSuccess(c, "Stripe early fraud warning recorded for manual review", gin.H{
		"event_id":               event.ID,
		"early_fraud_warning_id": warning.ID,
		"payment_intent_id":      paymentIntentID,
		"actionable":             warning.Actionable,
	})
}

func (h *Handler) recordAndRefreshPaymentRiskEvent(c *gin.Context, input service.PaymentRiskEventInput) error {
	if h == nil || h.riskMonitoring == nil || !h.riskMonitoring.Enabled() {
		return nil
	}
	if err := h.riskMonitoring.RecordEvent(input); err != nil {
		return err
	}
	_, err := h.riskMonitoring.RecomputeProvider(c.Request.Context(), input.Provider, time.Now().UTC())
	return err
}

func (h *Handler) enqueueRefundRecommendation(input service.PaymentRiskEventInput) error {
	if h == nil || h.refundReview == nil || !h.refundReview.Enabled() {
		return nil
	}
	_, err := h.refundReview.EnqueueFromRiskEvent(input)
	return err
}

func stripeEarlyFraudWarningAmount(warning stripe.RadarEarlyFraudWarning) (float64, string) {
	minorAmount := int64(0)
	currency := ""
	if warning.Charge != nil {
		minorAmount = warning.Charge.Amount
		currency = string(warning.Charge.Currency)
	}
	if (minorAmount <= 0 || currency == "") && warning.PaymentIntent != nil {
		minorAmount = warning.PaymentIntent.AmountReceived
		if minorAmount <= 0 {
			minorAmount = warning.PaymentIntent.Amount
		}
		currency = string(warning.PaymentIntent.Currency)
	}
	if minorAmount <= 0 || strings.TrimSpace(currency) == "" {
		return 0, strings.ToUpper(strings.TrimSpace(currency))
	}
	amount, err := pgateway.MinorToMajorAmount(minorAmount, currency)
	if err != nil {
		return 0, strings.ToUpper(strings.TrimSpace(currency))
	}
	return amount, strings.ToUpper(strings.TrimSpace(currency))
}

func stripeRiskOccurredAt(unixSeconds int64) time.Time {
	if unixSeconds <= 0 {
		return time.Now().UTC()
	}
	return time.Unix(unixSeconds, 0).UTC()
}
