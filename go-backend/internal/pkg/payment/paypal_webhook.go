package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/plutov/paypal/v4"
)

// VerifyWebhook 验证PayPal Webhook签名
func (g *paypalGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
	return false, fmt.Errorf("paypal webhook verification requires full PayPal transmission headers")
}

const paypalWebhookVerificationStatusSuccess = "SUCCESS"

type PayPalWebhookEvent struct {
	ID           string          `json:"id"`
	EventType    string          `json:"event_type"`
	ResourceType string          `json:"resource_type"`
	Resource     json.RawMessage `json:"resource"`
}

type PayPalWebhookVerifier interface {
	VerifyWebhookSignature(ctx context.Context, httpReq *http.Request, webhookID string) (*paypal.VerifyWebhookResponse, error)
}

func VerifyPayPalWebhook(ctx context.Context, config *Config, headers http.Header, payload []byte, verifier PayPalWebhookVerifier) (PayPalWebhookEvent, error) {
	if config == nil {
		return PayPalWebhookEvent{}, fmt.Errorf("paypal config is required")
	}
	if strings.TrimSpace(config.WebhookSecret) == "" {
		return PayPalWebhookEvent{}, fmt.Errorf("paypal webhook_id is not configured")
	}
	for _, key := range []string{"PAYPAL-AUTH-ALGO", "PAYPAL-CERT-URL", "PAYPAL-TRANSMISSION-ID", "PAYPAL-TRANSMISSION-SIG", "PAYPAL-TRANSMISSION-TIME"} {
		if strings.TrimSpace(headers.Get(key)) == "" {
			return PayPalWebhookEvent{}, fmt.Errorf("missing paypal webhook header %s", key)
		}
	}
	if verifier == nil {
		client, err := newPayPalVerificationClient(config)
		if err != nil {
			return PayPalWebhookEvent{}, err
		}
		verifier = client
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://paypal.local/webhook", bytes.NewReader(payload))
	if err != nil {
		return PayPalWebhookEvent{}, err
	}
	req.Header = headers.Clone()

	verification, err := verifier.VerifyWebhookSignature(ctx, req, config.WebhookSecret)
	if err != nil {
		return PayPalWebhookEvent{}, fmt.Errorf("paypal webhook signature verification failed: %w", err)
	}
	if verification == nil || !strings.EqualFold(verification.VerificationStatus, paypalWebhookVerificationStatusSuccess) {
		status := ""
		if verification != nil {
			status = verification.VerificationStatus
		}
		return PayPalWebhookEvent{}, fmt.Errorf("paypal webhook signature verification rejected: %s", status)
	}

	var event PayPalWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return PayPalWebhookEvent{}, fmt.Errorf("invalid paypal webhook payload: %w", err)
	}
	if strings.TrimSpace(event.ID) == "" {
		return PayPalWebhookEvent{}, fmt.Errorf("paypal webhook event id is required")
	}
	if strings.TrimSpace(event.EventType) == "" {
		return PayPalWebhookEvent{}, fmt.Errorf("paypal webhook event_type is required")
	}
	return event, nil
}

func newPayPalVerificationClient(config *Config) (*paypal.Client, error) {
	if strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		return nil, fmt.Errorf("paypal client_id and secret are required for webhook verification")
	}
	apiBase := paypal.APIBaseSandBox
	if strings.EqualFold(config.Environment, "production") {
		apiBase = paypal.APIBaseLive
	}
	client, err := paypal.NewClient(config.APIKey, config.SecretKey, apiBase)
	if err != nil {
		return nil, err
	}
	client.SetHTTPClient(&http.Client{Timeout: 10 * time.Second})
	return client, nil
}
