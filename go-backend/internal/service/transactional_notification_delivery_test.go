package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingTransactionalNotificationSender struct {
	to       []string
	subject  string
	htmlBody string
	textBody string
}

func (s *recordingTransactionalNotificationSender) SendEmail(to []string, subject, body string) error {
	s.to = append([]string(nil), to...)
	s.subject = subject
	s.textBody = body
	return nil
}

func (s *recordingTransactionalNotificationSender) SendRenderedEmail(to []string, subject, htmlBody, textBody string) error {
	s.to = append([]string(nil), to...)
	s.subject = subject
	s.htmlBody = htmlBody
	s.textBody = textBody
	return nil
}

func TestCanonicalDomainEventHandlerBuildsRendersAndSendsNotification(t *testing.T) {
	templateService := newTransactionalNotificationTemplateServiceForTest(t)
	definition, err := LookupTransactionalNotificationTemplate(NotificationTemplateOrderConfirmation)
	require.NoError(t, err)
	_, err = templateService.Save(SaveTransactionalNotificationTemplateInput{
		Code:              definition.Code,
		Locale:            "en",
		Category:          definition.Category,
		Name:              "Order confirmation",
		SubjectTemplate:   "Order {{order_number}} confirmed",
		BodyHTML:          "<p>{{order_number}} {{customer_name}}</p>",
		BodyText:          "{{order_number}} {{customer_name}} {{order_amount}}",
		AllowedVariables:  definition.AllowedVariables,
		RequiredVariables: definition.RequiredVariables,
		IsEnabled:         true,
		Version:           1,
	})
	require.NoError(t, err)

	payload, err := json.Marshal(outbox.OrderPaymentSucceededPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{
			SchemaVersion:  1,
			OccurredAt:     time.Date(2026, 9, 20, 1, 2, 3, 0, time.UTC),
			IdempotencyKey: "payment_transaction:txn-1",
		},
		NotificationAudienceSnapshot: outbox.NotificationAudienceSnapshot{
			RecipientEmail: "buyer@example.com",
			Locale:         "en-US",
			CustomerName:   "Ada Buyer",
		},
		OrderID:              7,
		OrderNumber:          "ORD-7",
		PaymentTransactionID: "txn-1",
		AmountMinor:          12500,
		Currency:             "USD",
	})
	require.NoError(t, err)

	sender := &recordingTransactionalNotificationSender{}
	handler := NewCanonicalDomainEventOutboxHandlerWithSender(templateService, sender)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		ID:        99,
		EventKey:  "order.payment_succeeded:7:txn-1",
		EventType: outbox.EventTypeOrderPaymentSucceeded,
		Payload:   payload,
	}))

	assert.Equal(t, []string{"buyer@example.com"}, sender.to)
	assert.Equal(t, "Order ORD-7 confirmed", sender.subject)
	assert.Contains(t, sender.htmlBody, "Ada Buyer")
	assert.Contains(t, sender.textBody, "ORD-7")
	assert.Contains(t, sender.textBody, "125.00")
}

func TestDeliverTransactionalNotificationUsesTextCompatibilitySender(t *testing.T) {
	templateService := newTransactionalNotificationTemplateServiceForTest(t)
	definition, err := LookupTransactionalNotificationTemplate(NotificationTemplateOrderDelivered)
	require.NoError(t, err)
	_, err = templateService.Save(SaveTransactionalNotificationTemplateInput{
		Code:              definition.Code,
		Locale:            "en",
		Category:          definition.Category,
		Name:              "Order delivered",
		SubjectTemplate:   "Order {{order_number}} delivered",
		BodyText:          "Delivered at {{delivered_at}}",
		AllowedVariables:  definition.AllowedVariables,
		RequiredVariables: definition.RequiredVariables,
		IsEnabled:         true,
		Version:           1,
	})
	require.NoError(t, err)

	payload, err := json.Marshal(outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{
			SchemaVersion:  1,
			OccurredAt:     time.Date(2026, 9, 20, 1, 2, 3, 0, time.UTC),
			IdempotencyKey: "delivery_fact:7:TRK-7",
		},
		NotificationAudienceSnapshot: outbox.NotificationAudienceSnapshot{
			RecipientEmail: "buyer@example.com",
		},
		OrderID:           7,
		OrderNumber:       "ORD-7",
		DeliveredAt:       time.Date(2026, 9, 20, 1, 2, 3, 0, time.UTC),
		NewShippingStatus: "delivered",
	})
	require.NoError(t, err)

	sender := &textOnlyTransactionalNotificationSender{}
	require.NoError(t, deliverTransactionalNotification(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDelivered,
		Payload:   payload,
	}, templateService, sender))
	assert.Equal(t, "Order ORD-7 delivered", sender.subject)
	assert.Contains(t, sender.body, "Delivered at")
}

type textOnlyTransactionalNotificationSender struct {
	subject string
	body    string
}

func (s *textOnlyTransactionalNotificationSender) SendEmail(_ []string, subject, body string) error {
	s.subject = subject
	s.body = body
	return nil
}
