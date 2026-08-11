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
	client     *http.Client
}

func NewOutboxWebhookDispatcher(webhookURL, token string, client *http.Client) *OutboxWebhookDispatcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &OutboxWebhookDispatcher{
		webhookURL: strings.TrimSpace(webhookURL),
		token:      strings.TrimSpace(token),
		client:     client,
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create outbox webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Outbox-Event-Key", event.EventKey)
	req.Header.Set("X-Outbox-Event-Type", event.EventType)
	if d.token != "" {
		req.Header.Set("Authorization", "Bearer "+d.token)
	}

	resp, err := d.client.Do(req)
	if err != nil {
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
