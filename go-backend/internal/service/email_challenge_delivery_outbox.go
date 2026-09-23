package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/emailtoken"
)

// EmailChallengeDeliveryOutboxHandler sends one-time verification links only
// after the challenge transaction has committed.
type EmailChallengeDeliveryOutboxHandler struct {
	sender         EmailChallengeSender
	resultRecorder EmailDeliveryResultRecorder
	secret         string
	now            func() time.Time
}

type EmailDeliveryResultRecorder interface {
	RecordDeliveryResult(channel string, success bool)
}

func NewEmailChallengeDeliveryOutboxHandler(sender EmailChallengeSender, resultRecorder EmailDeliveryResultRecorder, secret string) *EmailChallengeDeliveryOutboxHandler {
	return &EmailChallengeDeliveryOutboxHandler{
		sender:         sender,
		resultRecorder: resultRecorder,
		secret:         strings.TrimSpace(secret),
		now:            time.Now,
	}
}

func (h *EmailChallengeDeliveryOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.sender == nil || h.secret == "" {
		return errors.New("email challenge sender is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.EventType != outbox.EventTypeEmailChallengeDelivery {
		return fmt.Errorf("unsupported email challenge event type %s", event.EventType)
	}
	var payload outbox.EmailChallengeDeliveryPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode email challenge delivery event: %w", err)
	}
	recipient := strings.TrimSpace(payload.RecipientEmail)
	parsed, err := mail.ParseAddress(recipient)
	if err != nil || parsed.Address != recipient {
		return errors.New("email challenge recipient is invalid")
	}
	if strings.TrimSpace(payload.DeliverySubject) == "" ||
		!strings.Contains(payload.BodyTemplate, emailChallengeTokenPlaceholder) ||
		strings.TrimSpace(payload.Purpose) == "" ||
		strings.TrimSpace(payload.ChallengeSubject) == "" ||
		strings.TrimSpace(payload.Nonce) == "" ||
		payload.ExpiresAt.IsZero() ||
		payload.RequestedAt.IsZero() {
		return errors.New("email challenge delivery event is incomplete")
	}
	now := time.Now().UTC()
	if h.now != nil {
		now = h.now().UTC()
	}
	if !payload.ExpiresAt.After(now) {
		return nil
	}
	token, err := emailtoken.Sign(h.secret, emailtoken.Claims{
		Purpose:   strings.TrimSpace(payload.Purpose),
		Email:     recipient,
		Subject:   payload.ChallengeSubject,
		Nonce:     strings.TrimSpace(payload.Nonce),
		ExpiresAt: payload.ExpiresAt.Unix(),
	})
	if err != nil {
		return fmt.Errorf("rebuild email challenge token: %w", err)
	}
	if event.AggregateID == "" || emailtoken.Hash(token) != event.AggregateID {
		return errors.New("email challenge delivery event does not match challenge token")
	}
	body := strings.ReplaceAll(payload.BodyTemplate, emailChallengeTokenPlaceholder, token)
	err = h.sender.SendEmail([]string{recipient}, strings.TrimSpace(payload.DeliverySubject), body)
	if h.resultRecorder != nil {
		h.resultRecorder.RecordDeliveryResult("email", err == nil)
	}
	return err
}
