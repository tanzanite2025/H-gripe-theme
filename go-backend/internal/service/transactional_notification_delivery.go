package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
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

// TransactionalNotificationProviderIdentity is optional metadata supplied by
// a runtime sender when it can identify the selected outbound channel.
type TransactionalNotificationProviderIdentity interface {
	TransactionalNotificationProviderCode() string
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
	return deliverTransactionalNotificationWithRecordService(ctx, event, templateService, sender, nil)
}

func deliverTransactionalNotificationWithRecordService(
	ctx context.Context,
	event outbox.Event,
	templateService *TransactionalNotificationTemplateService,
	sender TransactionalNotificationSender,
	recordService *TransactionalNotificationDeliveryRecordService,
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

	referenceType, referenceNumber := transactionalNotificationReference(plan)
	providerCode := ""
	if identifiedSender, ok := sender.(TransactionalNotificationProviderIdentity); ok {
		providerCode = identifiedSender.TransactionalNotificationProviderCode()
	}
	attemptedAt := time.Now().UTC()
	if event.Attempts <= 0 {
		event.Attempts = 1
	}
	if recordService != nil {
		// A missing audit row must never turn a successful SMTP delivery into an
		// Outbox retry, so audit writes are deliberately best effort.
		_ = recordService.StartEmailDeliveryRecord(
			plan.IdempotencyKey,
			event.ID,
			event.EventType,
			plan.TemplateCode,
			plan.Locale,
			rendered.TemplateVersion,
			plan.RecipientEmail,
			rendered.Subject,
			referenceType,
			referenceNumber,
			event.Attempts,
			providerCode,
			attemptedAt,
		)
	}

	if renderedSender, ok := sender.(RenderedTransactionalNotificationSender); ok {
		if err := renderedSender.SendRenderedEmail(
			[]string{plan.RecipientEmail},
			rendered.Subject,
			rendered.HTML,
			rendered.Text,
		); err != nil {
			if recordService != nil {
				status := notificationDeliveryRecordFailureStatus(err)
				_ = recordService.MarkEmailDeliveryRecordFailed(plan.IdempotencyKey, status, err.Error(), time.Now().UTC())
			}
			return fmt.Errorf("send transactional notification %q: %w", plan.TemplateCode, err)
		}
		if recordService != nil {
			_ = recordService.MarkEmailDeliveryRecordSent(plan.IdempotencyKey, time.Now().UTC())
		}
		return nil
	}

	if err := sender.SendEmail([]string{plan.RecipientEmail}, rendered.Subject, rendered.Text); err != nil {
		if recordService != nil {
			status := notificationDeliveryRecordFailureStatus(err)
			_ = recordService.MarkEmailDeliveryRecordFailed(plan.IdempotencyKey, status, err.Error(), time.Now().UTC())
		}
		return fmt.Errorf("send transactional notification %q: %w", plan.TemplateCode, err)
	}
	if recordService != nil {
		_ = recordService.MarkEmailDeliveryRecordSent(plan.IdempotencyKey, time.Now().UTC())
	}
	return nil
}

func transactionalNotificationReference(plan TransactionalNotificationDeliveryPlan) (string, string) {
	if value := strings.TrimSpace(plan.Variables["after_sales_case_number"]); value != "" {
		return "after_sales", value
	}
	if value := strings.TrimSpace(plan.Variables["order_number"]); value != "" {
		return "order", value
	}
	return "", ""
}

func notificationDeliveryRecordFailureStatus(err error) string {
	if errors.Is(err, resilience.ErrExternalOutcomeUnknown) {
		return notification.EmailDeliveryStatusUnknown
	}
	return notification.EmailDeliveryStatusFailed
}
