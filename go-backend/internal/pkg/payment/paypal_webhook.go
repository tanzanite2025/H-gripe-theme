package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/plutov/paypal/v4"
)

// Keep the verifier deadline bounded for request handling. PayPal's
// transmission-time freshness is checked by PayPal itself; this timeout is
// only the outbound verification call and must not be used as a timestamp
// replay window.
const paypalWebhookVerificationTimeout = 5 * time.Second

// ErrPayPalWebhookVerificationUnavailable marks a transient failure while
// asking PayPal to verify a webhook. Callers must return a 5xx response so
// PayPal retries the event; it must not be treated as an invalid signature.
var ErrPayPalWebhookVerificationUnavailable = errors.New("paypal webhook verification temporarily unavailable")

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
	if ctx == nil {
		ctx = context.Background()
	}
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

	verificationCtx, cancel := context.WithTimeout(ctx, paypalWebhookVerificationTimeout)
	defer cancel()

	verification, err := verifier.VerifyWebhookSignature(verificationCtx, req, config.WebhookSecret)
	if err != nil {
		if isTransientPayPalWebhookVerificationError(err) {
			return PayPalWebhookEvent{}, fmt.Errorf("%w: %v", ErrPayPalWebhookVerificationUnavailable, err)
		}
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

func isTransientPayPalWebhookVerificationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var paypalErr *paypal.ErrorResponse
	return errors.As(err, &paypalErr) && paypalErr != nil && paypalErr.Response != nil && paypalErr.Response.StatusCode >= http.StatusInternalServerError
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
	client.SetHTTPClient(&http.Client{
		Timeout: paypalWebhookVerificationTimeout,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
		},
	})
	return client, nil
}
