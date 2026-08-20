package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
)

var ErrVerifiedConversionOutboxWebhookNotConfigured = errors.New("verified conversion outbox webhook is not configured")

type VerifiedConversionOutboxWebhookHandler struct {
	dispatcher *OutboxWebhookDispatcher
}

func NewVerifiedConversionOutboxWebhookHandlerFromEnv() *VerifiedConversionOutboxWebhookHandler {
	return newVerifiedConversionOutboxWebhookHandler(
		resilience.HTTPRetryPolicy{},
		nil,
	)
}

func NewVerifiedConversionOutboxWebhookHandlerFromEnvWithResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *VerifiedConversionOutboxWebhookHandler {
	return newVerifiedConversionOutboxWebhookHandler(retry, breaker)
}

func newVerifiedConversionOutboxWebhookHandler(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *VerifiedConversionOutboxWebhookHandler {
	url := firstNonEmptyEnv(
		"VERIFIED_CONVERSION_OUTBOX_WEBHOOK_URL",
		"OUTBOX_VERIFIED_CONVERSION_WEBHOOK_URL",
	)
	token := firstNonEmptyEnv(
		"VERIFIED_CONVERSION_OUTBOX_WEBHOOK_TOKEN",
		"OUTBOX_VERIFIED_CONVERSION_WEBHOOK_TOKEN",
	)
	if breaker != nil {
		return &VerifiedConversionOutboxWebhookHandler{
			dispatcher: NewOutboxWebhookDispatcherWithResilience(url, token, nil, retry, breaker),
		}
	}
	return &VerifiedConversionOutboxWebhookHandler{
		dispatcher: NewOutboxWebhookDispatcher(
			url,
			token,
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
