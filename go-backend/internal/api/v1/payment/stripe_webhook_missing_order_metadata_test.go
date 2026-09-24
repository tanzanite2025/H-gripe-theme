package payment

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76/webhook"
)

func TestStripePaymentIntentSucceededWithoutOrderMetadataFailsInboxAndDoesNotAcknowledge(t *testing.T) {
	const webhookSecret = "whsec_missing_order_metadata"
	t.Setenv("STRIPE_WEBHOOK_SECRET", webhookSecret)

	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	require.NoError(t, db.AutoMigrate(&paymentdomain.StripeWebhookEvent{}))

	payload := []byte(`{
		"id": "evt_stripe_missing_order_metadata",
		"object": "event",
		"api_version": "2023-10-16",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_stripe_missing_order_metadata",
				"object": "payment_intent",
				"amount": 15000,
				"amount_received": 15000,
				"currency": "usd",
				"status": "succeeded",
				"metadata": {}
			}
		}
	}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    webhookSecret,
		Timestamp: time.Now(),
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "stripe"}}
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payments/stripe/webhook",
		bytes.NewReader(payload),
	)
	context.Request.Header.Set("Stripe-Signature", signed.Header)

	handler.HandleWebhook(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "missing order_number metadata")
	require.NotContains(t, recorder.Body.String(), "Ignored Stripe payment_intent.succeeded without order metadata")

	var savedEvent paymentdomain.StripeWebhookEvent
	require.NoError(t, db.Where("event_id = ?", "evt_stripe_missing_order_metadata").First(&savedEvent).Error)
	require.Equal(t, "failed", savedEvent.Status)
	require.Contains(t, savedEvent.ErrorMessage, "HTTP 400")

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).
		Where("transaction_id = ?", "pi_stripe_missing_order_metadata").
		Count(&transactionCount).Error)
	require.Zero(t, transactionCount)
}
