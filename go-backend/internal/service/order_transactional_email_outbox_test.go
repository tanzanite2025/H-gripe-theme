package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingTransactionalOrderEmailSender struct {
	confirmations []outbox.OrderConfirmationEmailPayload
	shipments     []outbox.OrderShippingNotificationEmailPayload
	err           error
}

func (s *recordingTransactionalOrderEmailSender) SendOrderConfirmation(_ string, data interface{}) error {
	if s.err != nil {
		return s.err
	}
	payload, ok := data.(outbox.OrderConfirmationEmailPayload)
	if !ok {
		return errors.New("unexpected order confirmation payload")
	}
	s.confirmations = append(s.confirmations, payload)
	return nil
}

func (s *recordingTransactionalOrderEmailSender) SendShippingNotification(_ string, data interface{}) error {
	if s.err != nil {
		return s.err
	}
	payload, ok := data.(outbox.OrderShippingNotificationEmailPayload)
	if !ok {
		return errors.New("unexpected shipping notification payload")
	}
	s.shipments = append(s.shipments, payload)
	return nil
}

func TestOrderTransactionalEmailOutboxHandlerSendsConfirmationAndShippingEmails(t *testing.T) {
	sender := &recordingTransactionalOrderEmailSender{}
	handler := NewOrderTransactionalEmailOutboxHandler(sender)
	paidAt := time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)
	shippedAt := paidAt.Add(48 * time.Hour)

	confirmationPayload, err := json.Marshal(outbox.OrderConfirmationEmailPayload{
		RecipientEmail: "ada.rider@example.test",
		CustomerName:   "Ada Rider",
		OrderID:        10,
		OrderNumber:    "ORD-EMAIL-10",
		Amount:         1500,
		Currency:       "USD",
		PaidAt:         paidAt,
	})
	require.NoError(t, err)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderConfirmationEmail,
		Payload:   confirmationPayload,
	}))
	require.Len(t, sender.confirmations, 1)
	assert.Equal(t, "ORD-EMAIL-10", sender.confirmations[0].OrderNumber)
	assert.Equal(t, "USD", sender.confirmations[0].Currency)

	shippingPayload, err := json.Marshal(outbox.OrderShippingNotificationEmailPayload{
		RecipientEmail: "ada.rider@example.test",
		CustomerName:   "Ada Rider",
		OrderID:        10,
		OrderNumber:    "ORD-EMAIL-10",
		CarrierName:    "DHL",
		TrackingNumber: "DHL-TRACK-10",
		TrackingURL:    "https://tracking.example.test/DHL-TRACK-10",
		ShippedAt:      shippedAt,
	})
	require.NoError(t, err)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderShippingNotificationEmail,
		Payload:   shippingPayload,
	}))
	require.Len(t, sender.shipments, 1)
	assert.Equal(t, "DHL-TRACK-10", sender.shipments[0].TrackingNumber)
	assert.Equal(t, "https://tracking.example.test/DHL-TRACK-10", sender.shipments[0].TrackingURL)
}

func TestOrderTransactionalEmailOutboxHandlerReturnsErrorsForInvalidPayloadAndSendFailure(t *testing.T) {
	handler := NewOrderTransactionalEmailOutboxHandler(&recordingTransactionalOrderEmailSender{})
	require.Error(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderConfirmationEmail,
		Payload:   []byte(`{"order_id":1}`),
	}))

	sender := &recordingTransactionalOrderEmailSender{err: errors.New("smtp unavailable")}
	handler = NewOrderTransactionalEmailOutboxHandler(sender)
	payload, err := json.Marshal(outbox.OrderConfirmationEmailPayload{
		RecipientEmail: "ada.rider@example.test",
		OrderID:        10,
		OrderNumber:    "ORD-EMAIL-10",
		Amount:         1500,
		Currency:       "USD",
		PaidAt:         time.Now().UTC(),
	})
	require.NoError(t, err)

	err = handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderConfirmationEmail,
		Payload:   payload,
	})
	require.ErrorIs(t, err, sender.err)
}

func TestResolveOrderTrackingURLAcceptsOnlyHTTPURLs(t *testing.T) {
	assert.Equal(
		t,
		"https://tracking.example.test/track/ABC%20123",
		resolveOrderTrackingURL("https://tracking.example.test/track/{tracking_number}", "ABC 123"),
	)
	assert.Empty(t, resolveOrderTrackingURL("javascript:alert(1)", "ABC123"))
	assert.Empty(t, resolveOrderTrackingURL("mailto:support@example.test", "ABC123"))
}
