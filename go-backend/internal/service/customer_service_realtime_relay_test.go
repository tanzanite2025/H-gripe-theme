package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceRealtimeRelayPublishesOnceAndReplays(t *testing.T) {
	client := newCustomerServiceRealtimeRelayTestClient(t)
	relay := newCustomerServiceRealtimeRelayForTest(t, client, NewCustomerServiceEventHub())
	firstEvent := customerServiceRealtimeRelayTestEvent(481, 123)
	secondEvent := customerServiceRealtimeRelayTestEvent(482, 123)

	firstID, err := relay.Publish(context.Background(), firstEvent)
	require.NoError(t, err)
	duplicateID, err := relay.Publish(context.Background(), firstEvent)
	require.NoError(t, err)
	assert.Equal(t, firstID, duplicateID)
	secondID, err := relay.Publish(context.Background(), secondEvent)
	require.NoError(t, err)

	length, err := client.XLen(context.Background(), relay.config.Stream).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(2), length)

	replayed, err := relay.ReplayAfter(context.Background(), firstID, 20)
	require.NoError(t, err)
	require.Len(t, replayed, 1)
	assert.Equal(t, secondID, replayed[0].StreamID)
	assert.Equal(t, secondEvent.EventID, replayed[0].EventID)
	assert.Equal(t, secondEvent.TicketID, replayed[0].TicketID)
}

func TestCustomerServiceRealtimeRelayFansOutToEveryLocalHub(t *testing.T) {
	client := newCustomerServiceRealtimeRelayTestClient(t)
	hubA := NewCustomerServiceEventHub()
	hubB := NewCustomerServiceEventHub()
	relayA := newCustomerServiceRealtimeRelayForTest(t, client, hubA)
	relayB := newCustomerServiceRealtimeRelayForTest(t, client, hubB)
	require.NoError(t, relayA.Start(context.Background()))
	require.NoError(t, relayB.Start(context.Background()))
	t.Cleanup(relayA.Stop)
	t.Cleanup(relayB.Stop)

	subscriptionA := hubA.SubscribeConversation(123)
	subscriptionB := hubB.SubscribeConversation(123)
	t.Cleanup(subscriptionA.Cancel)
	t.Cleanup(subscriptionB.Cancel)

	streamID, err := relayA.Publish(context.Background(), customerServiceRealtimeRelayTestEvent(482, 123))
	require.NoError(t, err)

	assertCustomerServiceRealtimeRelayEvent(t, subscriptionA.Events(), streamID, CustomerServiceMessageCreatedEventID(482))
	assertCustomerServiceRealtimeRelayEvent(t, subscriptionB.Events(), streamID, CustomerServiceMessageCreatedEventID(482))
}

func TestCustomerServiceRealtimeRelayRejectsInvalidReplayCursor(t *testing.T) {
	client := newCustomerServiceRealtimeRelayTestClient(t)
	relay := newCustomerServiceRealtimeRelayForTest(t, client, NewCustomerServiceEventHub())

	_, err := relay.ReplayAfter(context.Background(), "not-a-stream-id", 10)
	assert.True(t, errors.Is(err, errInvalidCustomerServiceRealtimeCursor))
}

func TestCustomerServiceRealtimeRelaySkipsTruncatedReplayWindow(t *testing.T) {
	client := newCustomerServiceRealtimeRelayTestClient(t)
	relay := newCustomerServiceRealtimeRelayForTest(t, client, NewCustomerServiceEventHub())
	_, err := relay.Publish(context.Background(), customerServiceRealtimeRelayTestEvent(483, 123))
	require.NoError(t, err)

	replayed, err := relay.ReplayAfter(context.Background(), "0-0", 20)
	require.NoError(t, err)
	assert.Empty(t, replayed)
}

func TestCustomerServiceRealtimeRelayReplaysAudience(t *testing.T) {
	client := newCustomerServiceRealtimeRelayTestClient(t)
	relay := newCustomerServiceRealtimeRelayForTest(t, client, NewCustomerServiceEventHub())
	event := customerServiceRealtimeRelayTestEvent(484, 123)
	event.Audience = CustomerServiceRealtimeAudienceBackoffice

	streamID, err := relay.Publish(context.Background(), event)
	require.NoError(t, err)

	replayed, err := relay.ReplayAfter(context.Background(), streamID, 20)
	require.NoError(t, err)
	assert.Empty(t, replayed)

	followup := customerServiceRealtimeRelayTestEvent(485, 123)
	followup.Audience = CustomerServiceRealtimeAudiencePublic
	_, err = relay.Publish(context.Background(), followup)
	require.NoError(t, err)

	replayed, err = relay.ReplayAfter(context.Background(), streamID, 20)
	require.NoError(t, err)
	require.Len(t, replayed, 1)
	assert.Equal(t, CustomerServiceRealtimeAudiencePublic, replayed[0].Audience)
}

func newCustomerServiceRealtimeRelayTestClient(t *testing.T) redis.UniversalClient {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

func newCustomerServiceRealtimeRelayForTest(t *testing.T, client redis.UniversalClient, hub *CustomerServiceEventHub) *CustomerServiceRealtimeRelay {
	t.Helper()

	relay, err := NewCustomerServiceRealtimeRelay(client, hub, CustomerServiceRealtimeRelayConfig{
		Stream:         "customer_service:{realtime-test}:v1",
		StreamMaxLen:   100,
		ReplayLimit:    20,
		ConsumerBlock:  100 * time.Millisecond,
		DedupRetention: time.Hour,
	})
	require.NoError(t, err)
	return relay
}

func customerServiceRealtimeRelayTestEvent(messageID, ticketID uint) CustomerServiceRealtimeEvent {
	return NewCustomerServiceRealtimeEventWithID(
		CustomerServiceMessageCreatedEventID(messageID),
		CustomerServiceEventMessageCreated,
		ticketID,
		"public-conversation-id",
		CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
		map[string]uint{"message_id": messageID},
	)
}

func assertCustomerServiceRealtimeRelayEvent(t *testing.T, events <-chan CustomerServiceRealtimeEvent, streamID, eventID string) {
	t.Helper()

	select {
	case event := <-events:
		assert.Equal(t, streamID, event.StreamID)
		assert.Equal(t, eventID, event.EventID)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for relay event")
	}
}
