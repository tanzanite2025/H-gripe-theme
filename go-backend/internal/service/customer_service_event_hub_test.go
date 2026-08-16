package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceEventHubPublishesToInboxAndConversation(t *testing.T) {
	hub := NewCustomerServiceEventHub()
	inbox := hub.SubscribeInbox()
	conversationOne := hub.SubscribeConversation(1)
	conversationTwo := hub.SubscribeConversation(2)
	t.Cleanup(inbox.Cancel)
	t.Cleanup(conversationOne.Cancel)
	t.Cleanup(conversationTwo.Cancel)

	event := NewCustomerServiceRealtimeEvent(
		CustomerServiceEventMessageCreated,
		1,
		"conversation-one",
		CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
		map[string]interface{}{"message": "hello"},
	)
	hub.Publish(event)

	assertCustomerServiceRealtimeEvent(t, inbox.Events(), event)
	assertCustomerServiceRealtimeEvent(t, conversationOne.Events(), event)
	assertNoCustomerServiceRealtimeEvent(t, conversationTwo.Events())
}

func TestCustomerServiceEventHubCancelStopsDelivery(t *testing.T) {
	hub := NewCustomerServiceEventHub()
	subscription := hub.SubscribeConversation(1)
	subscription.Cancel()

	hub.Publish(NewCustomerServiceRealtimeEvent(
		CustomerServiceEventMessageCreated,
		1,
		"conversation-one",
		CustomerServiceRealtimeActor{Kind: "agent"},
		nil,
	))

	_, ok := <-subscription.Events()
	assert.False(t, ok)
}

func TestCustomerServiceEventHubPublishesStreamCursorUpgradeOnce(t *testing.T) {
	hub := NewCustomerServiceEventHub()
	subscription := hub.SubscribeConversation(1)
	t.Cleanup(subscription.Cancel)

	event := NewCustomerServiceRealtimeEventWithID(
		CustomerServiceMessageCreatedEventID(481),
		CustomerServiceEventMessageCreated,
		1,
		"conversation-one",
		CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
		nil,
	)
	hub.Publish(event)
	assertCustomerServiceRealtimeEvent(t, subscription.Events(), event)

	event.StreamID = "1750000000000-0"
	hub.Publish(event)
	upgraded := readCustomerServiceRealtimeHubEvent(t, subscription.Events())
	assert.Equal(t, event.StreamID, upgraded.StreamID)

	hub.Publish(event)
	assertNoCustomerServiceRealtimeEvent(t, subscription.Events())
}

func TestCustomerServiceRealtimeEventAudienceSeparatesProducts(t *testing.T) {
	publicOnly := NewCustomerServiceRealtimeEventWithIDAndAudience(
		"public-event",
		CustomerServiceEventTyping,
		1,
		"conversation-one",
		CustomerServiceRealtimeActor{Kind: "agent"},
		CustomerServiceRealtimeAudiencePublic,
		nil,
	)
	backofficeOnly := NewCustomerServiceRealtimeEventWithIDAndAudience(
		"backoffice-event",
		CustomerServiceEventMessagesRead,
		1,
		"conversation-one",
		CustomerServiceRealtimeActor{Kind: "agent"},
		CustomerServiceRealtimeAudienceBackoffice,
		nil,
	)
	legacyBoth := CustomerServiceRealtimeEvent{Audience: ""}

	assert.True(t, publicOnly.DeliversTo(CustomerServiceRealtimeAudiencePublic))
	assert.False(t, publicOnly.DeliversTo(CustomerServiceRealtimeAudienceBackoffice))
	assert.False(t, backofficeOnly.DeliversTo(CustomerServiceRealtimeAudiencePublic))
	assert.True(t, backofficeOnly.DeliversTo(CustomerServiceRealtimeAudienceBackoffice))
	assert.True(t, legacyBoth.DeliversTo(CustomerServiceRealtimeAudiencePublic))
	assert.True(t, legacyBoth.DeliversTo(CustomerServiceRealtimeAudienceBackoffice))
}

func readCustomerServiceRealtimeHubEvent(t *testing.T, events <-chan CustomerServiceRealtimeEvent) CustomerServiceRealtimeEvent {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for customer-service realtime hub event")
		return CustomerServiceRealtimeEvent{}
	}
}

func assertCustomerServiceRealtimeEvent(t *testing.T, events <-chan CustomerServiceRealtimeEvent, expected CustomerServiceRealtimeEvent) {
	t.Helper()

	select {
	case actual := <-events:
		require.Equal(t, expected.EventID, actual.EventID)
		require.Equal(t, expected.Type, actual.Type)
		require.Equal(t, expected.TicketID, actual.TicketID)
		require.Equal(t, expected.ConversationID, actual.ConversationID)
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for customer-service realtime event")
	}
}

func assertNoCustomerServiceRealtimeEvent(t *testing.T, events <-chan CustomerServiceRealtimeEvent) {
	t.Helper()

	select {
	case actual := <-events:
		t.Fatalf("unexpected customer-service realtime event: %#v", actual)
	case <-time.After(50 * time.Millisecond):
	}
}
