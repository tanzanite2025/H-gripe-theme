package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
)

var ErrOrderPaidOutboxWebhookNotConfigured = errors.New("order paid outbox webhook is not configured")

type OrderPaidOutboxWebhookHandler struct {
	dispatcher *OutboxWebhookDispatcher
}

func NewOrderPaidOutboxWebhookHandlerFromEnv() *OrderPaidOutboxWebhookHandler {
	return newOrderPaidOutboxWebhookHandler(
		resilience.HTTPRetryPolicy{},
		nil,
	)
}

func NewOrderPaidOutboxWebhookHandlerFromEnvWithResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *OrderPaidOutboxWebhookHandler {
	return newOrderPaidOutboxWebhookHandler(retry, breaker)
}

func newOrderPaidOutboxWebhookHandler(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *OrderPaidOutboxWebhookHandler {
	url := firstNonEmptyEnv(
		"ORDER_PAID_OUTBOX_WEBHOOK_URL",
		"OUTBOX_ORDER_PAID_WEBHOOK_URL",
		"ERP_ORDER_PAID_WEBHOOK_URL",
	)
	token := firstNonEmptyEnv(
		"ORDER_PAID_OUTBOX_WEBHOOK_TOKEN",
		"OUTBOX_ORDER_PAID_WEBHOOK_TOKEN",
		"ERP_ORDER_PAID_WEBHOOK_TOKEN",
	)
	if breaker != nil {
		return &OrderPaidOutboxWebhookHandler{
			dispatcher: NewOutboxWebhookDispatcherWithResilience(url, token, nil, retry, breaker),
		}
	}
	return &OrderPaidOutboxWebhookHandler{
		dispatcher: NewOutboxWebhookDispatcher(
			url,
			token,
			nil,
		),
	}
}

func (h *OrderPaidOutboxWebhookHandler) Configured() bool {
	return h != nil && h.dispatcher != nil && h.dispatcher.Configured()
}

func (h *OrderPaidOutboxWebhookHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || !h.Configured() {
		return ErrOrderPaidOutboxWebhookNotConfigured
	}
	return h.dispatcher.Dispatch(ctx, event)
}
