package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceMessageCreatedEventUsesCanonicalInvalidationPayload(t *testing.T) {
	actorUserID := uint(17)
	occurredAt := time.Date(2026, 8, 16, 10, 30, 0, 0, time.UTC)
	event := NewCustomerServiceMessageCreatedEvent(
		123,
		"public-conversation-id",
		481,
		occurredAt,
		CustomerServiceRealtimeActor{Kind: "agent", UserID: &actorUserID},
	)

	assert.Equal(t, CustomerServiceEventMessageCreated, event.Type)
	assert.Equal(t, CustomerServiceMessageCreatedEventID(481), event.EventID)
	assert.Equal(t, CustomerServiceRealtimeAudienceBoth, event.Audience)
	assert.Equal(t, uint(123), event.TicketID)
	assert.Equal(t, "public-conversation-id", event.ConversationID)
	assert.Equal(t, occurredAt, event.OccurredAt)
	assert.Equal(t, CustomerServiceRealtimeActor{Kind: "agent", UserID: &actorUserID}, event.Actor)
	assert.Equal(t, CustomerServiceMessageCreatedPayload{MessageID: 481}, event.Payload)

	raw, err := json.Marshal(event)
	require.NoError(t, err)
	var decoded struct {
		Payload map[string]interface{} `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, map[string]interface{}{"message_id": float64(481)}, decoded.Payload)
}

func TestCustomerServiceMessageCreatedEventNormalizesMissingOccurrenceTime(t *testing.T) {
	event := NewCustomerServiceMessageCreatedEvent(
		123,
		"public-conversation-id",
		481,
		time.Time{},
		CustomerServiceRealtimeActor{Kind: "customer", Anonymous: true},
	)

	assert.False(t, event.OccurredAt.IsZero())
	assert.True(t, event.OccurredAt.Location() == time.UTC)
}
