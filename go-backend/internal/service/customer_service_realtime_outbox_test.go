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

func TestCustomerServiceRealtimeOutboxHandlerPublishesOnceForDuplicateDelivery(t *testing.T) {
	hub := NewCustomerServiceEventHub()
	subscription := hub.SubscribeInbox()
	t.Cleanup(subscription.Cancel)

	messagePayload, err := json.Marshal(map[string]uint{"message_id": 481})
	require.NoError(t, err)
	payload, err := json.Marshal(outbox.CustomerServiceRealtimePayload{
		Type:           CustomerServiceEventMessageCreated,
		EventID:        CustomerServiceMessageCreatedEventID(481),
		TicketID:       123,
		ConversationID: "public-conversation-id",
		OccurredAt:     time.Now().UTC(),
		Actor: outbox.CustomerServiceRealtimeActor{
			Kind:      "customer",
			Anonymous: true,
		},
		Payload: messagePayload,
	})
	require.NoError(t, err)

	handler := NewCustomerServiceRealtimeOutboxHandler(hub)
	event := outbox.Event{
		EventType: outbox.EventTypeCustomerServiceRealtime,
		Payload:   payload,
	}

	require.NoError(t, handler.Handle(context.Background(), event))
	actual := readCustomerServiceRealtimeEvent(t, subscription.Events())
	assert.Equal(t, CustomerServiceMessageCreatedEventID(481), actual.EventID)
	assert.Equal(t, uint(123), actual.TicketID)
	assert.Equal(t, "customer", actual.Actor.Kind)

	require.NoError(t, handler.Handle(context.Background(), event))
	assertNoCustomerServiceRealtimeEvent(t, subscription.Events())
}

func TestCustomerServiceRealtimeOutboxHandlerPreservesAudience(t *testing.T) {
	hub := NewCustomerServiceEventHub()
	subscription := hub.SubscribeInbox()
	t.Cleanup(subscription.Cancel)

	payload, err := json.Marshal(outbox.CustomerServiceRealtimePayload{
		Type:           CustomerServiceEventMessagesRead,
		EventID:        "customer_service.messages.read:123:12:1:481",
		Audience:       string(CustomerServiceRealtimeAudienceBackoffice),
		TicketID:       123,
		ConversationID: "public-conversation-id",
		OccurredAt:     time.Now().UTC(),
		Actor: outbox.CustomerServiceRealtimeActor{
			Kind: "agent",
		},
		Payload: json.RawMessage(`{"last_read_message_id":481}`),
	})
	require.NoError(t, err)

	handler := NewCustomerServiceRealtimeOutboxHandler(hub)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeCustomerServiceRealtime,
		Payload:   payload,
	}))

	actual := readCustomerServiceRealtimeEvent(t, subscription.Events())
	assert.Equal(t, CustomerServiceRealtimeAudienceBackoffice, actual.Audience)
	assert.False(t, actual.DeliversTo(CustomerServiceRealtimeAudiencePublic))
	assert.True(t, actual.DeliversTo(CustomerServiceRealtimeAudienceBackoffice))
}

func TestCustomerServiceRealtimeOutboxHandlerPreservesCanonicalMessagePayload(t *testing.T) {
	hub := NewCustomerServiceEventHub()
	subscription := hub.SubscribeInbox()
	t.Cleanup(subscription.Cancel)

	occurredAt := time.Date(2026, 8, 16, 11, 0, 0, 0, time.UTC)
	source := NewCustomerServiceMessageCreatedEvent(
		123,
		"public-conversation-id",
		481,
		occurredAt,
		CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
	)
	payload, err := json.Marshal(outbox.CustomerServiceRealtimePayload{
		Type:           source.Type,
		EventID:        source.EventID,
		Audience:       string(source.Audience),
		TicketID:       source.TicketID,
		ConversationID: source.ConversationID,
		OccurredAt:     source.OccurredAt,
		Actor: outbox.CustomerServiceRealtimeActor{
			Kind:      source.Actor.Kind,
			UserID:    source.Actor.UserID,
			Anonymous: source.Actor.Anonymous,
		},
		Payload: json.RawMessage(`{"message_id":481}`),
	})
	require.NoError(t, err)

	handler := NewCustomerServiceRealtimeOutboxHandler(hub)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeCustomerServiceRealtime,
		Payload:   payload,
	}))

	actual := readCustomerServiceRealtimeEvent(t, subscription.Events())
	assert.Equal(t, source.Type, actual.Type)
	assert.Equal(t, source.EventID, actual.EventID)
	assert.Equal(t, source.Audience, actual.Audience)
	assert.Equal(t, source.TicketID, actual.TicketID)
	assert.Equal(t, source.ConversationID, actual.ConversationID)
	assert.Equal(t, source.OccurredAt, actual.OccurredAt)
	assert.Equal(t, source.Actor, actual.Actor)
	assert.Equal(t, map[string]interface{}{"message_id": float64(481)}, actual.Payload)
}

func readCustomerServiceRealtimeEvent(t *testing.T, events <-chan CustomerServiceRealtimeEvent) CustomerServiceRealtimeEvent {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for customer-service realtime event")
		return CustomerServiceRealtimeEvent{}
	}
}
