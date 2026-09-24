package payment

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76/webhook"
)

func TestStripeLiabilityShiftedFromPaymentIntentSucceeded(t *testing.T) {
	shifted := stripeLiabilityShiftedFromPaymentIntentSucceeded([]byte(`{
		"id": "pi_1",
		"latest_charge": {
			"payment_method_details": {
				"card": {
					"three_d_secure": {
						"liability_shifted": false
					}
				}
			}
		}
	}`), nil)

	require.NotNil(t, shifted)
	require.False(t, *shifted)

	shifted = stripeLiabilityShiftedFromPaymentIntentSucceeded(nil, []byte(`{
		"id": "evt_1",
		"data": {
			"object": {
				"metadata": {
					"liability_shifted": "true"
				}
			}
		}
	}`))

	require.NotNil(t, shifted)
	require.True(t, *shifted)
}

func TestStripeLiabilityShiftedDoesNotInferFromThreeDSecureResult(t *testing.T) {
	shifted := stripeLiabilityShiftedFromPaymentIntentSucceeded([]byte(`{
		"id": "pi_1",
		"latest_charge": {
			"payment_method_details": {
				"card": {
					"three_d_secure": {
						"result": "authenticated"
					}
				}
			}
		}
	}`), nil)

	require.Nil(t, shifted)
}

func TestStripePaymentIntentSucceededRefetchesExpandedChargeWhenWebhookOnlyHasChargeID(t *testing.T) {
	const webhookSecret = "whsec_liability_refetch"
	t.Setenv("STRIPE_WEBHOOK_SECRET", webhookSecret)
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_liability_refetch")

	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	require.NoError(t, db.AutoMigrate(&paymentdomain.StripeWebhookEvent{}))
	orderRecord := seedPayPalOrder(t, db, "ORD-STRIPE-LIABILITY-REFETCH", 7, 120000, "pending", "unpaid")
	shifted := true
	gateway := &fakePaymentGateway{
		getResponse: &pgateway.PaymentResponse{LiabilityShifted: &shifted},
	}
	handler.gatewayFactory = func(*pgateway.Config) (pgateway.PaymentGateway, error) {
		return gateway, nil
	}

	payload := []byte(`{
		"id": "evt_stripe_liability_refetch",
		"object": "event",
		"api_version": "2023-10-16",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_stripe_liability_refetch",
				"object": "payment_intent",
				"amount": 120000,
				"amount_received": 120000,
				"currency": "usd",
				"status": "succeeded",
				"metadata": {"order_number": "ORD-STRIPE-LIABILITY-REFETCH"},
				"latest_charge": "ch_stripe_liability_refetch"
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

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, 1, gateway.getCalls)
	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "processing", savedOrder.Status)
	require.False(t, savedOrder.FulfillmentHold)

	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "pi_stripe_liability_refetch").First(&transaction).Error)
	require.NotNil(t, transaction.LiabilityShifted)
	require.True(t, *transaction.LiabilityShifted)
}

func TestStripePaymentIntentSucceededRetriesWhenLiabilityLookupFails(t *testing.T) {
	const webhookSecret = "whsec_liability_refetch_error"
	t.Setenv("STRIPE_WEBHOOK_SECRET", webhookSecret)
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_liability_refetch_error")

	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	require.NoError(t, db.AutoMigrate(&paymentdomain.StripeWebhookEvent{}))
	seedPayPalOrder(t, db, "ORD-STRIPE-LIABILITY-REFETCH-ERROR", 7, 120000, "pending", "unpaid")
	gateway := &fakePaymentGateway{getErr: errors.New("stripe timeout")}
	handler.gatewayFactory = func(*pgateway.Config) (pgateway.PaymentGateway, error) {
		return gateway, nil
	}

	payload := []byte(`{
		"id": "evt_stripe_liability_refetch_error",
		"object": "event",
		"api_version": "2023-10-16",
		"type": "payment_intent.succeeded",
		"data": {"object": {
			"id": "pi_stripe_liability_refetch_error",
			"amount": 120000,
			"amount_received": 120000,
			"currency": "usd",
			"status": "succeeded",
			"metadata": {"order_number": "ORD-STRIPE-LIABILITY-REFETCH-ERROR"},
			"latest_charge": "ch_stripe_liability_refetch_error"
		}}
	}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    webhookSecret,
		Timestamp: time.Now(),
	})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "provider", Value: "stripe"}}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/stripe/webhook", bytes.NewReader(payload))
	ctx.Request.Header.Set("Stripe-Signature", signed.Header)

	handler.HandleWebhook(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Equal(t, 1, gateway.getCalls)
	var savedEvent paymentdomain.StripeWebhookEvent
	require.NoError(t, db.Where("event_id = ?", "evt_stripe_liability_refetch_error").First(&savedEvent).Error)
	require.Equal(t, "failed", savedEvent.Status)
	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("transaction_id = ?", "pi_stripe_liability_refetch_error").Count(&transactionCount).Error)
	require.Zero(t, transactionCount)
}

func TestPayPalVerifiedPaymentFromEventExtractsLiabilityShifted(t *testing.T) {
	resource := []byte(`{
		"id": "PAYPAL-ORDER-1",
		"status": "COMPLETED",
		"purchase_units": [{
			"custom_id": "ORD-PAYPAL-LIABILITY",
			"payments": {
				"captures": [{
					"id": "PAYPAL-CAPTURE-1",
					"status": "COMPLETED",
					"amount": {
						"currency_code": "USD",
						"value": "84.00"
					},
					"liability_shifted": false
				}]
			}
		}]
	}`)

	payment, handled, err := paypalVerifiedPaymentFromEvent(pgateway.PayPalWebhookEvent{
		EventType: paypalCheckoutOrderCompleted,
		Resource:  resource,
	}, nil)

	require.NoError(t, err)
	require.True(t, handled)
	require.NotNil(t, payment.LiabilityShifted)
	require.False(t, *payment.LiabilityShifted)
}

func TestPaymentResponseLiabilityShiftedPrefersExplicitField(t *testing.T) {
	notShifted := false
	shifted := paymentResponseLiabilityShifted(&pgateway.PaymentResponse{
		LiabilityShifted: &notShifted,
		Metadata: map[string]string{
			"liability_shifted": "true",
		},
	}, nil)

	require.NotNil(t, shifted)
	require.False(t, *shifted)
}
