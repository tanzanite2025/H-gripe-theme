package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/outbox"
)

var ErrVerifiedConversionOutboxWebhookNotConfigured = errors.New("verified conversion outbox webhook is not configured")

type VerifiedConversionOutboxWebhookHandler struct {
	dispatcher *OutboxWebhookDispatcher
}

func NewVerifiedConversionOutboxWebhookHandlerFromEnv() *VerifiedConversionOutboxWebhookHandler {
	return &VerifiedConversionOutboxWebhookHandler{
		dispatcher: NewOutboxWebhookDispatcher(
			firstNonEmptyEnv(
				"VERIFIED_CONVERSION_OUTBOX_WEBHOOK_URL",
				"OUTBOX_VERIFIED_CONVERSION_WEBHOOK_URL",
			),
			firstNonEmptyEnv(
				"VERIFIED_CONVERSION_OUTBOX_WEBHOOK_TOKEN",
				"OUTBOX_VERIFIED_CONVERSION_WEBHOOK_TOKEN",
			),
			nil,
		),
	}
}

func (h *VerifiedConversionOutboxWebhookHandler) Configured() bool {
	return h != nil && h.dispatcher != nil && h.dispatcher.Configured()
}

func (h *VerifiedConversionOutboxWebhookHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || !h.Configured() {
		return ErrVerifiedConversionOutboxWebhookNotConfigured
	}
	return h.dispatcher.Dispatch(ctx, event)
}
