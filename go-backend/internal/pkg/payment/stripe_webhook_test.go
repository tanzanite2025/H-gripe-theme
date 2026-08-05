package payment

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

const stripeWebhookTestSecret = "whsec_test_secret"

func TestParseStripeWebhookEventEnforcesReplayTolerance(t *testing.T) {
	payload := signedStripeWebhookPayload(t)

	tests := []struct {
		name      string
		timestamp time.Time
		wantErr   error
	}{
		{
			name:      "fresh signature is accepted",
			timestamp: time.Now().Add(-stripeWebhookReplayTolerance / 2),
		},
		{
			name:      "stale signature is rejected",
			timestamp: time.Now().Add(-stripeWebhookReplayTolerance - time.Minute),
			wantErr:   webhook.ErrTooOld,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
				Payload:   payload,
				Secret:    stripeWebhookTestSecret,
				Timestamp: tt.timestamp,
			})

			event, err := ParseStripeWebhookEvent(payload, signedPayload.Header, stripeWebhookTestSecret)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected valid webhook, got error: %v", err)
			}
			if event.ID != "evt_test_replay_guard" {
				t.Fatalf("expected parsed event id, got %q", event.ID)
			}
		})
	}
}

func TestStripeGatewayVerifyWebhookEnforcesReplayTolerance(t *testing.T) {
	payload := signedStripeWebhookPayload(t)
	gateway := &stripeGatewayImpl{config: &Config{WebhookSecret: stripeWebhookTestSecret}}
	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    stripeWebhookTestSecret,
		Timestamp: time.Now().Add(-stripeWebhookReplayTolerance - time.Minute),
	})

	ok, err := gateway.VerifyWebhook(payload, signedPayload.Header)
	if ok {
		t.Fatal("expected stale webhook verification to fail")
	}
	if !errors.Is(err, webhook.ErrTooOld) {
		t.Fatalf("expected %v, got %v", webhook.ErrTooOld, err)
	}
}

func signedStripeWebhookPayload(t *testing.T) []byte {
	t.Helper()

	return []byte(fmt.Sprintf(`{
		"id": "evt_test_replay_guard",
		"object": "event",
		"api_version": %q,
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_test_replay_guard",
				"object": "payment_intent"
			}
		}
	}`, stripe.APIVersion))
}
