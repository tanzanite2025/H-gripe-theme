package payment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

func TestStripeVerifiedRefundInputUsesPaymentIntentAndMinorAmount(t *testing.T) {
	refund := &stripe.Refund{}
	refundJSON := []byte(`{
		"id": "re_stripe_1",
		"amount": 5000,
		"currency": "usd",
		"payment_intent": "pi_stripe_1",
		"balance_transaction": {
			"id": "txn_refund_1",
			"net": -5750,
			"currency": "usd"
		},
		"metadata": {"order_number": "ORD-STRIPE-1"}
	}`)
	require.NoError(t, json.Unmarshal(refundJSON, refund))

	input, err := stripeVerifiedRefundInput(refund, nil, refundJSON)

	require.NoError(t, err)
	require.Equal(t, string(pgateway.GatewayStripe), input.Provider)
	require.Equal(t, "ORD-STRIPE-1", input.OrderNumber)
	require.Equal(t, "pi_stripe_1", input.TransactionID)
	require.Equal(t, "re_stripe_1", input.RefundID)
	require.Equal(t, int64(5000), input.ProviderRefundAmount.AmountMinor())
	require.Equal(t, "USD", input.ProviderRefundAmount.Currency().String())
	require.Equal(t, int64(5750), input.SettlementAmountMinor)
	require.Equal(t, "USD", input.SettlementCurrency)
	require.Equal(t, "txn_refund_1", input.SettlementBalanceTransactionID)
}

func TestStripeChargeRefundedInputUsesExpandedRefundAndPaymentIntent(t *testing.T) {
	chargeJSON := []byte(`{
		"id": "ch_stripe_1",
		"currency": "usd",
		"payment_intent": "pi_stripe_charge_1",
		"refunds": {
			"data": [{
				"id": "re_stripe_charge_1",
				"amount": 8400,
				"currency": "usd",
				"status": "succeeded"
			}]
		}
	}`)
	var charge stripe.Charge
	require.NoError(t, json.Unmarshal(chargeJSON, &charge))
	require.NotNil(t, charge.Refunds)
	require.Len(t, charge.Refunds.Data, 1)

	input, err := stripeVerifiedRefundInput(charge.Refunds.Data[0], &charge, chargeJSON)

	require.NoError(t, err)
	require.Equal(t, "pi_stripe_charge_1", input.TransactionID)
	require.Equal(t, "re_stripe_charge_1", input.RefundID)
	require.Equal(t, int64(8400), input.ProviderRefundAmount.AmountMinor())
}

func TestPayPalVerifiedRefundFromEventUsesCaptureLink(t *testing.T) {
	resource := []byte(`{
		"id": "REFUND-PAYPAL-1",
		"status": "COMPLETED",
		"amount": {
			"currency_code": "USD",
			"value": "84.00"
		},
		"links": [{
			"rel": "up",
			"href": "https://api-m.paypal.com/v2/payments/captures/CAPTURE-PAYPAL-1"
		}]
	}`)

	input, handled, err := paypalVerifiedRefundFromEvent(pgateway.PayPalWebhookEvent{
		EventType: paypalPaymentCaptureRefunded,
		Resource:  resource,
	}, resource)

	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, string(pgateway.GatewayPayPal), input.Provider)
	require.Equal(t, "CAPTURE-PAYPAL-1", input.TransactionID)
	require.Equal(t, "REFUND-PAYPAL-1", input.RefundID)
	require.Equal(t, int64(8400), input.ProviderRefundAmount.AmountMinor())
	require.Zero(t, input.RequestedRefundAmount)
	require.Equal(t, "USD", input.ProviderRefundAmount.Currency().String())
}

func TestPayPalVerifiedRefundFromEventIgnoresFailedRefund(t *testing.T) {
	input, handled, err := paypalVerifiedRefundFromEvent(pgateway.PayPalWebhookEvent{
		EventType: paypalPaymentCaptureRefunded,
		Resource: []byte(`{
			"id": "REFUND-PAYPAL-FAILED",
			"status": "FAILED",
			"capture_id": "CAPTURE-PAYPAL-FAILED",
			"amount": {"currency_code": "USD", "value": "10.00"}
		}`),
	}, nil)

	require.NoError(t, err)
	require.False(t, handled)
	require.Empty(t, input.RefundID)
}

func TestAlipayVerifiedRefundNotificationDefaultsCurrencyAndParsesAmount(t *testing.T) {
	input, err := alipayVerifiedRefundFromNotification(pgateway.AlipayWebhookNotification{
		AppID:        "app-id",
		OutTradeNo:   "ORD-ALIPAY-REFUND",
		TradeNo:      "TRADE-ALIPAY-1",
		OutRequestNo: "REFUND-ALIPAY-1",
		RefundStatus: "REFUND_SUCCESS",
		RefundAmount: "12.34",
	}, []byte(`{"refund_status":"REFUND_SUCCESS"}`))

	require.NoError(t, err)
	require.Equal(t, string(pgateway.GatewayAlipay), input.Provider)
	require.Equal(t, "TRADE-ALIPAY-1", input.TransactionID)
	require.Equal(t, "REFUND-ALIPAY-1", input.RefundID)
	require.Equal(t, "CNY", input.ProviderRefundAmount.Currency().String())
	require.Equal(t, int64(1234), input.ProviderRefundAmount.AmountMinor())
}

func TestAlipayVerifiedRefundNotificationCarriesFailureStatus(t *testing.T) {
	input, err := alipayVerifiedRefundFromNotification(pgateway.AlipayWebhookNotification{
		OutTradeNo:   "ORD-ALIPAY-REFUND-FAILED",
		TradeNo:      "TRADE-ALIPAY-FAILED",
		OutRequestNo: "REFUND-ALIPAY-FAILED",
		RefundStatus: "REFUND_CLOSED",
		RefundAmount: "20.00",
		Currency:     "CNY",
	}, nil)

	require.NoError(t, err)
	require.Equal(t, "REFUND_CLOSED", input.ProviderStatus)
	require.Equal(t, "CNY", input.ProviderRefundAmount.Currency().String())
	require.Equal(t, int64(2000), input.ProviderRefundAmount.AmountMinor())
}

func TestStripeRefundCreatedHandlerPersistsExternalRefund(t *testing.T) {
	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	orderRecord := seedPayPalDisputeWebhookOrder(t, db, "ORD-STRIPE-REFUND-WEBHOOK", 8400, "DHL-STRIPE")
	require.NoError(t, db.Model(&orderdomain.Order{}).Where("id = ?", orderRecord.ID).Update("user_id", 0).Error)
	seedPayPalDisputeWebhookTransaction(t, db, orderRecord.ID, "pi_stripe_refund_webhook", 8400)

	refundJSON := []byte(`{
		"id": "re_stripe_webhook",
		"amount": 8400,
		"currency": "usd",
		"payment_intent": "pi_stripe_refund_webhook",
		"status": "succeeded"
	}`)
	event := stripe.Event{
		ID:   "evt_stripe_refund_webhook",
		Data: &stripe.EventData{Raw: refundJSON},
	}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handler.handleStripeRefundCreated(context, event, refundJSON)
	writeStripeWebhookSuccess(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var refund paymentdomain.Refund
	require.NoError(t, db.Where("refund_id = ?", "re_stripe_webhook").First(&refund).Error)
	require.Equal(t, "completed", refund.Status)
	require.Equal(t, int64(8400), refund.AmountMinor)
}

func TestStripePaymentIntentSucceededResolvesRequiresActionReview(t *testing.T) {
	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	orderRecord := seedPayPalDisputeWebhookOrder(t, db, "ORD-STRIPE-3DS-SUCCESS", 8400, "DHL-STRIPE-3DS")
	require.NoError(t, db.Model(&orderdomain.Order{}).
		Where("id = ?", orderRecord.ID).
		Updates(map[string]interface{}{
			"status":          "pending",
			"payment_method":  "stripe",
			"payment_status":  "unpaid",
			"shipping_status": "pending",
		}).Error)
	_, err := handler.paymentService.CreatePaymentReview(service.CreatePaymentReviewInput{
		OrderID:         &orderRecord.ID,
		PaymentIntentID: "pi_stripe_3ds_success",
		Status:          "pending",
		Reason:          "stripe_requires_action",
		Source:          "radar",
		Notes:           "Stripe requested customer authentication or additional payment action.",
	})
	require.NoError(t, err)

	payload := []byte(`{
		"id": "pi_stripe_3ds_success",
		"amount": 8400,
		"amount_received": 8400,
		"currency": "usd",
		"metadata": {"order_number": "ORD-STRIPE-3DS-SUCCESS"}
	}`)
	event := stripe.Event{
		ID:   "evt_stripe_3ds_success",
		Type: "payment_intent.succeeded",
		Data: &stripe.EventData{Raw: payload},
	}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handler.handleStripePaymentIntentSucceeded(context, event, payload)
	writeStripeWebhookSuccess(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var savedReview paymentdomain.PaymentReview
	require.NoError(t, db.Where("payment_intent_id = ?", "pi_stripe_3ds_success").First(&savedReview).Error)
	assert.Equal(t, "approved", savedReview.Status)
	require.NotNil(t, savedReview.ReviewedAt)

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)
}

func TestStripeRefundCreatedWebhookRouteClaimsAndFinalizesInbox(t *testing.T) {
	const webhookSecret = "whsec_refund_route_test"
	t.Setenv("STRIPE_WEBHOOK_SECRET", webhookSecret)

	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	require.NoError(t, db.AutoMigrate(&paymentdomain.StripeWebhookEvent{}))
	orderRecord := seedPayPalDisputeWebhookOrder(t, db, "ORD-STRIPE-ROUTE-REFUND", 8400, "DHL-STRIPE-ROUTE")
	require.NoError(t, db.Model(&orderdomain.Order{}).Where("id = ?", orderRecord.ID).Update("user_id", 0).Error)
	seedPayPalDisputeWebhookTransaction(t, db, orderRecord.ID, "pi_stripe_route_refund", 8400)

	payload := []byte(`{
		"id": "evt_stripe_route_refund",
		"object": "event",
		"api_version": "2023-10-16",
		"type": "refund.created",
		"data": {
			"object": {
				"id": "re_stripe_route_refund",
				"amount": 8400,
				"currency": "usd",
				"payment_intent": "pi_stripe_route_refund",
				"status": "succeeded"
			}
		}
	}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    webhookSecret,
		Timestamp: time.Now(),
	})
	_, parseErr := pgateway.ParseStripeWebhookEvent(payload, signed.Header, webhookSecret)
	require.NoError(t, parseErr)
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

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Stripe refund webhook processed successfully")

	var savedEvent paymentdomain.StripeWebhookEvent
	require.NoError(t, db.Where("event_id = ?", "evt_stripe_route_refund").First(&savedEvent).Error)
	require.Equal(t, "processed", savedEvent.Status)
}

func TestStripeChargeRefundedHandlerPersistsExpandedRefund(t *testing.T) {
	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	orderRecord := seedPayPalDisputeWebhookOrder(t, db, "ORD-STRIPE-CHARGE-REFUND", 8400, "DHL-STRIPE-CHARGE")
	require.NoError(t, db.Model(&orderdomain.Order{}).Where("id = ?", orderRecord.ID).Update("user_id", 0).Error)
	seedPayPalDisputeWebhookTransaction(t, db, orderRecord.ID, "pi_stripe_charge_refund", 8400)

	chargeJSON := []byte(`{
		"id": "ch_stripe_refund_webhook",
		"amount_refunded": 8400,
		"currency": "usd",
		"payment_intent": "pi_stripe_charge_refund",
		"refunds": {
			"data": [{
				"id": "re_stripe_charge_webhook",
				"amount": 8400,
				"currency": "usd",
				"status": "succeeded"
			}]
		}
	}`)
	event := stripe.Event{
		ID:   "evt_stripe_charge_refund_webhook",
		Data: &stripe.EventData{Raw: chargeJSON},
	}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handler.handleStripeChargeRefunded(context, event, chargeJSON)
	writeStripeWebhookSuccess(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var refund paymentdomain.Refund
	require.NoError(t, db.Where("refund_id = ?", "re_stripe_charge_webhook").First(&refund).Error)
	require.Equal(t, "completed", refund.Status)
	require.Equal(t, int64(8400), refund.AmountMinor)
}

func TestPayPalRefundWebhookPersistsExternalRefund(t *testing.T) {
	db, handler, _ := newPayPalDisputeWebhookHarness(t)
	orderRecord := seedPayPalDisputeWebhookOrder(t, db, "ORD-PAYPAL-REFUND-WEBHOOK", 8400, "DHL-PAYPAL")
	require.NoError(t, db.Model(&orderdomain.Order{}).Where("id = ?", orderRecord.ID).Update("user_id", 0).Error)
	seedPayPalDisputeWebhookTransaction(t, db, orderRecord.ID, "CAPTURE-PAYPAL-REFUND", 8400)

	input, handled, err := paypalVerifiedRefundFromEvent(pgateway.PayPalWebhookEvent{
		ID:        "WH-PAYPAL-REFUND-1",
		EventType: paypalPaymentCaptureRefunded,
		Resource: []byte(`{
			"id": "REFUND-PAYPAL-WEBHOOK",
			"status": "COMPLETED",
			"capture_id": "CAPTURE-PAYPAL-REFUND",
			"amount": {"currency_code": "USD", "value": "84.00"}
		}`),
	}, []byte(`{"id":"WH-PAYPAL-REFUND-1"}`))
	require.NoError(t, err)
	require.True(t, handled)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	require.True(t, handler.recordVerifiedGatewayRefund(context, input))

	var refund paymentdomain.Refund
	require.NoError(t, db.Where("refund_id = ?", "REFUND-PAYPAL-WEBHOOK").First(&refund).Error)
	require.Equal(t, "completed", refund.Status)
	require.Equal(t, int64(8400), refund.AmountMinor)
}
