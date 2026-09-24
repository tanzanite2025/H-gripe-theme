package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/email"
)

// OrderDisputeContactEmailOutboxHandler delivers manually requested dispute
// emails after the request transaction commits. SMTP failures are returned to
// the dispatcher so the durable event can retry or dead-letter normally.
type OrderDisputeContactEmailOutboxHandler struct {
	sender email.EmailService
}

func NewOrderDisputeContactEmailOutboxHandler(sender email.EmailService) *OrderDisputeContactEmailOutboxHandler {
	return &OrderDisputeContactEmailOutboxHandler{sender: sender}
}

func (h *OrderDisputeContactEmailOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.sender == nil {
		return errors.New("order dispute contact email sender is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.EventType != outbox.EventTypeOrderDisputeContactEmail {
		return fmt.Errorf("unsupported order dispute contact email event type %s", event.EventType)
	}
	var payload outbox.OrderDisputeContactEmailPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode order dispute contact email event: %w", err)
	}
	if payload.OrderID == 0 || strings.TrimSpace(payload.Provider) == "" || payload.DisputeID == 0 || strings.TrimSpace(payload.Subject) == "" || strings.TrimSpace(payload.Body) == "" || payload.RequestedAt.IsZero() {
		return errors.New("order dispute contact email event is incomplete")
	}
	recipient := strings.TrimSpace(payload.RecipientEmail)
	parsed, err := mail.ParseAddress(recipient)
	if err != nil || parsed.Address != recipient {
		return errors.New("order dispute contact email recipient is invalid")
	}
	return h.sender.SendEmail([]string{recipient}, strings.TrimSpace(payload.Subject), payload.Body)
}
