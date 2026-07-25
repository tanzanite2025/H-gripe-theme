package service

import (
	"sync"
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

type CustomerServiceRealtimeActor struct {
	Kind      string `json:"kind"`
	UserID    *uint  `json:"user_id,omitempty"`
	Anonymous bool   `json:"anonymous,omitempty"`
}

type CustomerServiceRealtimeEvent struct {
	Type           string                       `json:"type"`
	EventID        string                       `json:"event_id"`
	TicketID       uint                         `json:"ticket_id,omitempty"`
	ConversationID string                       `json:"conversation_id,omitempty"`
	OccurredAt     time.Time                    `json:"occurred_at"`
	Actor          CustomerServiceRealtimeActor `json:"actor"`
	Payload        interface{}                  `json:"payload,omitempty"`
}

type CustomerServiceEventSubscription struct {
	id     uint64
	events <-chan CustomerServiceRealtimeEvent
	cancel func()
}

type CustomerServiceEventHub struct {
	mu                      sync.RWMutex
	nextID                  uint64
	inboxSubscribers        map[uint64]chan CustomerServiceRealtimeEvent
	conversationSubscribers map[uint]map[uint64]chan CustomerServiceRealtimeEvent
	bufferSize              int
}

func NewCustomerServiceEventHub() *CustomerServiceEventHub {
	return &CustomerServiceEventHub{
		inboxSubscribers:        make(map[uint64]chan CustomerServiceRealtimeEvent),
		conversationSubscribers: make(map[uint]map[uint64]chan CustomerServiceRealtimeEvent),
		bufferSize:              32,
	}
}

func NewCustomerServiceRealtimeEvent(eventType string, ticketID uint, conversationID string, actor CustomerServiceRealtimeActor, payload interface{}) CustomerServiceRealtimeEvent {
	return CustomerServiceRealtimeEvent{
		Type:           eventType,
		EventID:        uuid.NewString(),
		TicketID:       ticketID,
		ConversationID: conversationID,
		OccurredAt:     time.Now().UTC(),
		Actor:          actor,
		Payload:        payload,
	}
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

func (h *CustomerServiceEventHub) SubscribeInbox() *CustomerServiceEventSubscription {
	if h == nil {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id, events := h.nextSubscriptionLocked()
	h.inboxSubscribers[id] = events

	return &CustomerServiceEventSubscription{
		id:     id,
		events: events,
		cancel: func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if subscriber, ok := h.inboxSubscribers[id]; ok {
				delete(h.inboxSubscribers, id)
				close(subscriber)
			}
		},
	}
}

func (h *CustomerServiceEventHub) SubscribeConversation(ticketID uint) *CustomerServiceEventSubscription {
	if h == nil || ticketID == 0 {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id, events := h.nextSubscriptionLocked()
	if h.conversationSubscribers[ticketID] == nil {
		h.conversationSubscribers[ticketID] = make(map[uint64]chan CustomerServiceRealtimeEvent)
	}
	h.conversationSubscribers[ticketID][id] = events

	return &CustomerServiceEventSubscription{
		id:     id,
		events: events,
		cancel: func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			subscribers := h.conversationSubscribers[ticketID]
			if subscribers == nil {
				return
			}
			if subscriber, ok := subscribers[id]; ok {
				delete(subscribers, id)
				close(subscriber)
			}
			if len(subscribers) == 0 {
				delete(h.conversationSubscribers, ticketID)
			}
		},
	}
}

func (h *CustomerServiceEventHub) Publish(event CustomerServiceRealtimeEvent) {
	if h == nil || event.Type == "" {
		return
	}

	h.mu.RLock()
	inboxSubscribers := cloneCustomerServiceSubscribers(h.inboxSubscribers)
	conversationSubscribers := cloneCustomerServiceSubscribers(h.conversationSubscribers[event.TicketID])
	h.mu.RUnlock()

	for _, subscriber := range inboxSubscribers {
		nonBlockingCustomerServiceEventSend(subscriber, event)
	}
	for _, subscriber := range conversationSubscribers {
		nonBlockingCustomerServiceEventSend(subscriber, event)
	}
}

func (s *CustomerServiceEventSubscription) Events() <-chan CustomerServiceRealtimeEvent {
	if s == nil {
		return nil
	}
	return s.events
}

func (s *CustomerServiceEventSubscription) Cancel() {
	if s == nil || s.cancel == nil {
		return
	}
	s.cancel()
}

func (h *CustomerServiceEventHub) nextSubscriptionLocked() (uint64, chan CustomerServiceRealtimeEvent) {
	h.nextID++
	return h.nextID, make(chan CustomerServiceRealtimeEvent, h.bufferSize)
}

func cloneCustomerServiceSubscribers(source map[uint64]chan CustomerServiceRealtimeEvent) []chan CustomerServiceRealtimeEvent {
	if len(source) == 0 {
		return nil
	}
	subscribers := make([]chan CustomerServiceRealtimeEvent, 0, len(source))
	for _, subscriber := range source {
		subscribers = append(subscribers, subscriber)
	}
	return subscribers
}

func nonBlockingCustomerServiceEventSend(subscriber chan CustomerServiceRealtimeEvent, event CustomerServiceRealtimeEvent) {
	select {
	case subscriber <- event:
	default:
		// Realtime is an acceleration layer. Dropping an overloaded subscriber is
		// safer than blocking the HTTP write path; clients keep HTTP polling/fetch
		// as the durable fallback.
	}
}
