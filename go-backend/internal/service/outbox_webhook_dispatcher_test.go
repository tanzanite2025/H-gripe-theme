package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tanzanite/internal/domain/outbox"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestOutboxWebhookDispatcherDeliversAuthenticatedEnvelope(t *testing.T) {
	received := make(chan OutboxWebhookEnvelope, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, "Bearer alert-token", request.Header.Get("Authorization"))
		require.Equal(t, "risk:stripe:2", request.Header.Get("X-Outbox-Event-Key"))
		require.Equal(t, outbox.EventTypePaymentRiskLevelChanged, request.Header.Get("X-Outbox-Event-Type"))

		var envelope OutboxWebhookEnvelope
		require.NoError(t, json.NewDecoder(request.Body).Decode(&envelope))
		received <- envelope
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	dispatcher := NewOutboxWebhookDispatcher(server.URL, "alert-token", server.Client())
	event := outbox.Event{
		ID:            7,
		EventKey:      "risk:stripe:2",
		EventType:     outbox.EventTypePaymentRiskLevelChanged,
		AggregateType: outbox.AggregateTypePaymentRiskProvider,
		AggregateID:   "stripe",
		Payload:       datatypes.JSON([]byte(`{"provider":"stripe","current_level":"warning"}`)),
		Attempts:      1,
		CreatedAt:     time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC),
	}

	require.NoError(t, dispatcher.Dispatch(context.Background(), event))
	envelope := <-received
	require.Equal(t, event.EventKey, envelope.EventKey)
	require.Equal(t, event.EventType, envelope.EventType)
	require.JSONEq(t, string(event.Payload), string(envelope.Payload))
}

func TestPaymentRiskAlertOutboxWebhookHandlerRequiresConfiguredDispatcher(t *testing.T) {
	handler := &PaymentRiskAlertOutboxWebhookHandler{}
	require.ErrorIs(t, handler.Handle(context.Background(), outbox.Event{}), ErrPaymentRiskAlertOutboxWebhookNotConfigured)
}
