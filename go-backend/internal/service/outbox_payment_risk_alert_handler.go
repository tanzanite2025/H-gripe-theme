package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
)

var ErrPaymentRiskAlertOutboxWebhookNotConfigured = errors.New("payment risk alert outbox webhook is not configured")

type PaymentRiskAlertOutboxWebhookHandler struct {
	dispatcher *OutboxWebhookDispatcher
}

func NewPaymentRiskAlertOutboxWebhookHandlerFromEnv() *PaymentRiskAlertOutboxWebhookHandler {
	return newPaymentRiskAlertOutboxWebhookHandler(
		resilience.HTTPRetryPolicy{},
		nil,
	)
}

func NewPaymentRiskAlertOutboxWebhookHandlerFromEnvWithResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *PaymentRiskAlertOutboxWebhookHandler {
	return newPaymentRiskAlertOutboxWebhookHandler(retry, breaker)
}

func newPaymentRiskAlertOutboxWebhookHandler(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) *PaymentRiskAlertOutboxWebhookHandler {
	url := firstNonEmptyEnv("PAYMENT_RISK_ALERT_OUTBOX_WEBHOOK_URL")
	token := firstNonEmptyEnv("PAYMENT_RISK_ALERT_OUTBOX_WEBHOOK_TOKEN")
	if breaker != nil {
		return &PaymentRiskAlertOutboxWebhookHandler{
			dispatcher: NewOutboxWebhookDispatcherWithResilience(url, token, nil, retry, breaker),
		}
	}
	return &PaymentRiskAlertOutboxWebhookHandler{
		dispatcher: NewOutboxWebhookDispatcher(
			url,
			token,
			nil,
		),
	}
}

func (h *PaymentRiskAlertOutboxWebhookHandler) Configured() bool {
	return h != nil && h.dispatcher != nil && h.dispatcher.Configured()
}

func (h *PaymentRiskAlertOutboxWebhookHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || !h.Configured() {
		return ErrPaymentRiskAlertOutboxWebhookNotConfigured
	}
	return h.dispatcher.Dispatch(ctx, event)
}
