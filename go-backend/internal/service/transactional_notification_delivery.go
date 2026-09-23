package service

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/outbox"
)

// TransactionalNotificationSender is the provider-neutral portion of the
// existing email service needed by the notification worker. Keeping this
// interface small lets tests use a recording sender without coupling the
// domain event pipeline to SMTP configuration.
type TransactionalNotificationSender interface {
	SendEmail(to []string, subject, body string) error
}

// RenderedTransactionalNotificationSender is implemented by the current email
// service so the worker can preserve both rendered HTML and plain text. A
// sender that only implements TransactionalNotificationSender receives the
// rendered plain-text body as a compatibility fallback.
type RenderedTransactionalNotificationSender interface {
	SendRenderedEmail(to []string, subject, htmlBody, textBody string) error
}

var (
	ErrTransactionalNotificationTemplateServiceRequired = errors.New("transactional notification template service is required")
	ErrTransactionalNotificationSenderRequired          = errors.New("transactional notification sender is required")
)

// deliverTransactionalNotification renders and sends one canonical event.
// The event payload remains the source of truth; no current order or case
// lookup is performed here. Outbox retries therefore replay the same snapshot
// rather than rebuilding a different message from mutable state.
func deliverTransactionalNotification(
	ctx context.Context,
	event outbox.Event,
	templateService *TransactionalNotificationTemplateService,
	sender TransactionalNotificationSender,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if templateService == nil {
		return ErrTransactionalNotificationTemplateServiceRequired
	}
	if sender == nil {
		return ErrTransactionalNotificationSenderRequired
	}

	plan, err := BuildTransactionalNotificationDeliveryPlan(event, TransactionalNotificationDeliveryContext{})
	if err != nil {
		return fmt.Errorf("build transactional notification delivery plan: %w", err)
	}
	if plan.TemplateCode == "" {
		// Valid but intentionally silent facts (for example after-sales
		// inspecting/cancelled/exception) do not create a customer message.
		return nil
	}

	rendered, err := templateService.Render(plan.TemplateCode, plan.Locale, plan.Variables)
	if err != nil {
		return fmt.Errorf("render transactional notification %q: %w", plan.TemplateCode, err)
	}

	if renderedSender, ok := sender.(RenderedTransactionalNotificationSender); ok {
		if err := renderedSender.SendRenderedEmail(
			[]string{plan.RecipientEmail},
			rendered.Subject,
			rendered.HTML,
			rendered.Text,
		); err != nil {
			return fmt.Errorf("send transactional notification %q: %w", plan.TemplateCode, err)
		}
		return nil
	}

	if err := sender.SendEmail([]string{plan.RecipientEmail}, rendered.Subject, rendered.Text); err != nil {
		return fmt.Errorf("send transactional notification %q: %w", plan.TemplateCode, err)
	}
	return nil
}
