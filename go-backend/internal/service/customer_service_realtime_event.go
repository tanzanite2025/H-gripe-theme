package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	CustomerServiceEventMessageCreated = "conversation.message.created"
	CustomerServiceEventMessagesRead   = "conversation.messages.read"
	CustomerServiceEventAssigned       = "conversation.assigned"
	CustomerServiceEventStatusChanged  = "conversation.status.changed"
	CustomerServiceEventContextUpdated = "conversation.context.updated"
	CustomerServiceEventTyping         = "conversation.typing"
	CustomerServiceEventHeartbeat      = "heartbeat"
)

type CustomerServiceRealtimeAudience string

const (
	CustomerServiceRealtimeAudienceBoth       CustomerServiceRealtimeAudience = "both"
	CustomerServiceRealtimeAudiencePublic     CustomerServiceRealtimeAudience = "public"
	CustomerServiceRealtimeAudienceBackoffice CustomerServiceRealtimeAudience = "backoffice"
)

type CustomerServiceRealtimeActor struct {
	Kind      string `json:"kind"`
	UserID    *uint  `json:"user_id,omitempty"`
	Anonymous bool   `json:"anonymous,omitempty"`
}

// CustomerServiceRealtimeEvent is the display-safe invalidation contract
// shared by the transactional Outbox, relay, and browser transports. It must
// not become a second source of message or conversation truth.
type CustomerServiceRealtimeEvent struct {
	Type     string                          `json:"type"`
	EventID  string                          `json:"event_id"`
	Audience CustomerServiceRealtimeAudience `json:"audience,omitempty"`
	// StreamID is delivery metadata. It is deliberately not part of the
	// browser event body; WebSocket sends it in the envelope cursor field.
	StreamID       string                       `json:"-"`
	TicketID       uint                         `json:"ticket_id,omitempty"`
	ConversationID string                       `json:"conversation_id,omitempty"`
	OccurredAt     time.Time                    `json:"occurred_at"`
	Actor          CustomerServiceRealtimeActor `json:"actor"`
	Payload        interface{}                  `json:"payload,omitempty"`
}

// CustomerServiceMessageCreatedPayload is intentionally minimal. Clients use
// the message ID to reconcile through their authorized HTTP history endpoint.
type CustomerServiceMessageCreatedPayload struct {
	MessageID uint `json:"message_id"`
}

// CustomerServiceEventReplayer returns a bounded set of durable events after a
// realtime cursor. Implementations must treat the cursor as untrusted input.
type CustomerServiceEventReplayer interface {
	ReplayAfter(ctx context.Context, afterID string, limit int) ([]CustomerServiceRealtimeEvent, error)
}

// CustomerServiceRealtimeEventDeduper is scoped to one realtime connection. It
// prevents a replayed event and the already-buffered live subscription from
// causing two client refreshes while retaining a fixed memory bound.
type CustomerServiceRealtimeEventDeduper struct {
	seen  map[string]string
	limit int
}

func CustomerServiceMessageCreatedEventID(messageID uint) string {
	return fmt.Sprintf("customer_service.message.created:%d", messageID)
}

func CustomerServiceMessagesReadEventID(ticketID, recipientUserID, assignmentVersion, lastReadMessageID uint) string {
	return fmt.Sprintf("customer_service.messages.read:%d:%d:%d:%d", ticketID, recipientUserID, assignmentVersion, lastReadMessageID)
}

func CustomerServiceConversationAssignedEventID(ticketID, recipientUserID, assignmentVersion uint) string {
	return fmt.Sprintf("customer_service.conversation.assigned:%d:%d:%d", ticketID, recipientUserID, assignmentVersion)
}

func CustomerServiceConversationStatusChangedEventID(ticketID, statusVersion uint) string {
	return fmt.Sprintf("customer_service.conversation.status.changed:%d:%d", ticketID, statusVersion)
}

// NewCustomerServiceMessageCreatedEvent is the sole producer for durable
// message-created invalidations. Immediate Hub delivery and Outbox recovery
// must therefore have the same wire contract and deterministic identity.
func NewCustomerServiceMessageCreatedEvent(
	ticketID uint,
	conversationID string,
	messageID uint,
	occurredAt time.Time,
	actor CustomerServiceRealtimeActor,
) CustomerServiceRealtimeEvent {
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	return CustomerServiceRealtimeEvent{
		Type:           CustomerServiceEventMessageCreated,
		EventID:        CustomerServiceMessageCreatedEventID(messageID),
		Audience:       CustomerServiceRealtimeAudienceBoth,
		TicketID:       ticketID,
		ConversationID: conversationID,
		OccurredAt:     occurredAt.UTC(),
		Actor:          actor,
		Payload:        CustomerServiceMessageCreatedPayload{MessageID: messageID},
	}
}

func NewCustomerServiceRealtimeEvent(eventType string, ticketID uint, conversationID string, actor CustomerServiceRealtimeActor, payload interface{}) CustomerServiceRealtimeEvent {
	return NewCustomerServiceRealtimeEventWithIDAndAudience("", eventType, ticketID, conversationID, actor, CustomerServiceRealtimeAudienceBoth, payload)
}

func NewCustomerServiceRealtimeEventWithID(eventID, eventType string, ticketID uint, conversationID string, actor CustomerServiceRealtimeActor, payload interface{}) CustomerServiceRealtimeEvent {
	return NewCustomerServiceRealtimeEventWithIDAndAudience(eventID, eventType, ticketID, conversationID, actor, CustomerServiceRealtimeAudienceBoth, payload)
}

func NewCustomerServiceRealtimeEventWithIDAndAudience(eventID, eventType string, ticketID uint, conversationID string, actor CustomerServiceRealtimeActor, audience CustomerServiceRealtimeAudience, payload interface{}) CustomerServiceRealtimeEvent {
	if eventID == "" {
		eventID = uuid.NewString()
	}
	return CustomerServiceRealtimeEvent{
		Type:           eventType,
		EventID:        eventID,
		Audience:       normalizeCustomerServiceRealtimeAudience(audience),
		TicketID:       ticketID,
		ConversationID: conversationID,
		OccurredAt:     time.Now().UTC(),
		Actor:          actor,
		Payload:        payload,
	}
}

// DeliversTo reports whether an event is intended for a browser product. An
// empty audience is the compatibility form used by pre-audience message
// events and is intentionally treated as both products.
func (e CustomerServiceRealtimeEvent) DeliversTo(audience CustomerServiceRealtimeAudience) bool {
	eventAudience := normalizeCustomerServiceRealtimeAudience(e.Audience)
	return eventAudience == CustomerServiceRealtimeAudienceBoth || eventAudience == audience
}

func NewCustomerServiceHeartbeatEvent() CustomerServiceRealtimeEvent {
	return NewCustomerServiceRealtimeEvent(
		CustomerServiceEventHeartbeat,
		0,
		"",
		CustomerServiceRealtimeActor{Kind: "system"},
		map[string]interface{}{"server_time": time.Now().UTC()},
	)
}

func NewCustomerServiceRealtimeEventDeduper(limit int) *CustomerServiceRealtimeEventDeduper {
	if limit <= 0 {
		limit = 1024
	}
	return &CustomerServiceRealtimeEventDeduper{
		seen:  make(map[string]string, limit),
		limit: limit,
	}
}

func (d *CustomerServiceRealtimeEventDeduper) First(event CustomerServiceRealtimeEvent) bool {
	if d == nil || event.EventID == "" {
		return true
	}
	if streamID, exists := d.seen[event.EventID]; exists {
		// A locally published durable message initially has no Stream ID. Let
		// the later relay delivery through once so this connection advances
		// its cursor; browser clients dedupe the business event before refresh.
		if streamID == "" && event.StreamID != "" {
			d.seen[event.EventID] = event.StreamID
			return true
		}
		return false
	}
	if len(d.seen) >= d.limit {
		clear(d.seen)
	}
	d.seen[event.EventID] = event.StreamID
	return true
}

func validateCustomerServiceRealtimeEvent(event CustomerServiceRealtimeEvent) error {
	if strings.TrimSpace(event.Type) == "" || strings.TrimSpace(event.EventID) == "" || event.TicketID == 0 {
		return errors.New("customer-service realtime event is incomplete")
	}
	if !validCustomerServiceRealtimeAudience(event.Audience) {
		return fmt.Errorf("customer-service realtime event has invalid audience %q", event.Audience)
	}
	return nil
}

func normalizeCustomerServiceRealtimeAudience(audience CustomerServiceRealtimeAudience) CustomerServiceRealtimeAudience {
	switch audience {
	case CustomerServiceRealtimeAudiencePublic, CustomerServiceRealtimeAudienceBackoffice:
		return audience
	case "", CustomerServiceRealtimeAudienceBoth:
		return CustomerServiceRealtimeAudienceBoth
	default:
		return audience
	}
}

func validCustomerServiceRealtimeAudience(audience CustomerServiceRealtimeAudience) bool {
	switch audience {
	case "", CustomerServiceRealtimeAudienceBoth, CustomerServiceRealtimeAudiencePublic, CustomerServiceRealtimeAudienceBackoffice:
		return true
	default:
		return false
	}
}
