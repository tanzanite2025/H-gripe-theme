package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
)

var ErrPaymentRefundOutboxWebhookNotConfigured = errors.New("payment refund outbox webhook is not configured")

// PaymentRefundOutboxWebhookHandler is the transport boundary for refund
// lifecycle events. The local transaction remains authoritative; this handler
// only delivers the committed event to an external consumer.
type PaymentRefundOutboxWebhookHandler struct {
	dispatcher *OutboxWebhookDispatcher
}

func NewPaymentRefundOutboxWebhookHandlerFromEnvWithResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *PaymentRefundOutboxWebhookHandler {
	url := firstNonEmptyEnv(
		"PAYMENT_REFUND_OUTBOX_WEBHOOK_URL",
		"OUTBOX_PAYMENT_REFUND_WEBHOOK_URL",
	)
	token := firstNonEmptyEnv(
		"PAYMENT_REFUND_OUTBOX_WEBHOOK_TOKEN",
		"OUTBOX_PAYMENT_REFUND_WEBHOOK_TOKEN",
	)
	if breaker != nil {
		return &PaymentRefundOutboxWebhookHandler{
			dispatcher: NewOutboxWebhookDispatcherWithResilience(url, token, nil, retry, breaker),
		}
	}
	return &PaymentRefundOutboxWebhookHandler{
		dispatcher: NewOutboxWebhookDispatcher(url, token, nil),
	}
}

func (h *PaymentRefundOutboxWebhookHandler) Configured() bool {
	return h != nil && h.dispatcher != nil && h.dispatcher.Configured()
}

func (h *PaymentRefundOutboxWebhookHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || !h.Configured() {
		return ErrPaymentRefundOutboxWebhookNotConfigured
	}
	switch event.EventType {
	case outbox.EventTypePaymentRefundPending,
		outbox.EventTypePaymentRefundCompleted,
		outbox.EventTypePaymentRefundFailed:
		return h.dispatcher.Dispatch(ctx, event)
	default:
		return errors.New("unsupported payment refund outbox event type")
	}
}
