package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/outbox"
)

var ErrPaymentRiskAlertOutboxWebhookNotConfigured = errors.New("payment risk alert outbox webhook is not configured")

type PaymentRiskAlertOutboxWebhookHandler struct {
	dispatcher *OutboxWebhookDispatcher
}

func NewPaymentRiskAlertOutboxWebhookHandlerFromEnv() *PaymentRiskAlertOutboxWebhookHandler {
	return &PaymentRiskAlertOutboxWebhookHandler{
		dispatcher: NewOutboxWebhookDispatcher(
			firstNonEmptyEnv("PAYMENT_RISK_ALERT_OUTBOX_WEBHOOK_URL"),
			firstNonEmptyEnv("PAYMENT_RISK_ALERT_OUTBOX_WEBHOOK_TOKEN"),
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
