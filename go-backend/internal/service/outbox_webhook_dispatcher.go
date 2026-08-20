package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
)

var ErrOutboxWebhookNotConfigured = errors.New("outbox webhook is not configured")

type OutboxWebhookEnvelope struct {
	ID            uint            `json:"id"`
	EventKey      string          `json:"event_key"`
	EventType     string          `json:"event_type"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
	Attempts      int             `json:"attempts"`
	CreatedAt     time.Time       `json:"created_at"`
}

// OutboxWebhookDispatcher is the shared transport boundary for external
// Outbox consumers. Receivers should treat EventKey as their idempotency key.
type OutboxWebhookDispatcher struct {
	webhookURL string
	token      string
	client     *resilience.HTTPClient
}

func NewOutboxWebhookDispatcher(webhookURL, token string, client *http.Client) *OutboxWebhookDispatcher {
	webhookURL = strings.TrimSpace(webhookURL)
	breakerKey := "outbox-webhook:" + webhookURL
	return NewOutboxWebhookDispatcherWithResilience(
		webhookURL,
		token,
		client,
		resilience.HTTPRetryPolicy{
			MaxAttempts: 3,
			Backoff: resilience.BackoffPolicy{
				BaseDelay: 250 * time.Millisecond,
				MaxDelay:  3 * time.Second,
				Jitter:    250 * time.Millisecond,
			},
			RetryUnsafeMethods: true,
		},
		resilience.SharedCircuitBreaker(breakerKey, resilience.CircuitBreakerConfig{
			FailureThreshold: 3,
			FailureWindow:    60 * time.Second,
			OpenDuration:     30 * time.Second,
		}),
	)
}

// NewOutboxWebhookDispatcherWithResilience accepts a shared circuit
// controller so every dispatcher replica observes the same third-party
// outage. EventKey remains the downstream idempotency key.
func NewOutboxWebhookDispatcherWithResilience(
	webhookURL string,
	token string,
	client *http.Client,
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *OutboxWebhookDispatcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	webhookURL = strings.TrimSpace(webhookURL)
	breakerKey := "outbox-webhook:" + webhookURL
	return &OutboxWebhookDispatcher{
		webhookURL: webhookURL,
		token:      strings.TrimSpace(token),
		client:     resilience.NewHTTPClient(client, retry, breaker, breakerKey),
	}
}

func (d *OutboxWebhookDispatcher) Configured() bool {
	return d != nil && strings.TrimSpace(d.webhookURL) != ""
}

func (d *OutboxWebhookDispatcher) Dispatch(ctx context.Context, event outbox.Event) error {
	if d == nil || !d.Configured() {
		return ErrOutboxWebhookNotConfigured
	}
	if ctx == nil {
		ctx = context.Background()
	}

	payload := json.RawMessage(bytes.TrimSpace(event.Payload))
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}
	body, err := json.Marshal(OutboxWebhookEnvelope{
		ID:            event.ID,
		EventKey:      event.EventKey,
		EventType:     event.EventType,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Payload:       payload,
		Attempts:      event.Attempts,
		CreatedAt:     event.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("encode outbox webhook envelope: %w", err)
	}

	resp, err := d.client.Do(ctx, func() (*http.Request, error) {
		clonedBody := bytes.NewReader(body)
		request, createErr := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, clonedBody)
		if createErr != nil {
			return nil, createErr
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Outbox-Event-Key", event.EventKey)
		request.Header.Set("X-Outbox-Event-Type", event.EventType)
		request.Header.Set("Idempotency-Key", event.EventKey)
		if d.token != "" {
			request.Header.Set("Authorization", "Bearer "+d.token)
		}
		return request, nil
	})
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return fmt.Errorf("send outbox webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("outbox webhook returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
