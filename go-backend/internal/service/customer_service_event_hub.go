package service

import (
	"context"
	"sync"
	"time"

	"commerce-platform/internal/pkg/metrics"
)

const (
	customerServiceRecentEventRetention = time.Hour
	customerServiceRecentEventLimit     = 10000
)

type CustomerServiceEventSubscription struct {
	id     uint64
	events <-chan CustomerServiceRealtimeEvent
	cancel func()
}

type customerServiceRecentEvent struct {
	seenAt   time.Time
	streamID string
}

type CustomerServiceEventHub struct {
	mu                      sync.RWMutex
	nextID                  uint64
	inboxSubscribers        map[uint64]chan CustomerServiceRealtimeEvent
	conversationSubscribers map[uint]map[uint64]chan CustomerServiceRealtimeEvent
	recentEventIDs          map[string]customerServiceRecentEvent
	replayProvider          CustomerServiceEventReplayer
	bufferSize              int
}

func NewCustomerServiceEventHub() *CustomerServiceEventHub {
	return &CustomerServiceEventHub{
		inboxSubscribers:        make(map[uint64]chan CustomerServiceRealtimeEvent),
		conversationSubscribers: make(map[uint]map[uint64]chan CustomerServiceRealtimeEvent),
		recentEventIDs:          make(map[string]customerServiceRecentEvent),
		bufferSize:              32,
	}
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

	h.mu.Lock()
	if !h.rememberEventLocked(event) {
		h.mu.Unlock()
		return
	}
	inboxSubscribers := cloneCustomerServiceSubscribers(h.inboxSubscribers)
	conversationSubscribers := cloneCustomerServiceSubscribers(h.conversationSubscribers[event.TicketID])
	h.mu.Unlock()

	for _, subscriber := range inboxSubscribers {
		recordCustomerServiceHubDelivery("inbox", nonBlockingCustomerServiceEventSend(subscriber, event))
	}
	for _, subscriber := range conversationSubscribers {
		recordCustomerServiceHubDelivery("conversation", nonBlockingCustomerServiceEventSend(subscriber, event))
	}
}

func (h *CustomerServiceEventHub) ConfigureReplayProvider(provider CustomerServiceEventReplayer) {
	if h == nil {
		return
	}

	h.mu.Lock()
	h.replayProvider = provider
	h.mu.Unlock()
}

func (h *CustomerServiceEventHub) ReplayAfter(ctx context.Context, afterID string, limit int) ([]CustomerServiceRealtimeEvent, error) {
	if h == nil {
		return nil, nil
	}

	h.mu.RLock()
	provider := h.replayProvider
	h.mu.RUnlock()
	if provider == nil {
		return nil, nil
	}
	return provider.ReplayAfter(ctx, afterID, limit)
}

func (h *CustomerServiceEventHub) rememberEventLocked(event CustomerServiceRealtimeEvent) bool {
	if event.EventID == "" {
		return true
	}

	now := time.Now().UTC()
	if previous, exists := h.recentEventIDs[event.EventID]; exists && now.Sub(previous.seenAt) < customerServiceRecentEventRetention {
		// Keep the normal immediate-local + Outbox recovery dedupe, but emit a
		// later Stream-backed copy once to advance connected client cursors.
		if previous.streamID == "" && event.StreamID != "" {
			h.recentEventIDs[event.EventID] = customerServiceRecentEvent{
				seenAt:   now,
				streamID: event.StreamID,
			}
			return true
		}
		return false
	}

	if len(h.recentEventIDs) >= customerServiceRecentEventLimit {
		cutoff := now.Add(-customerServiceRecentEventRetention)
		for id, recent := range h.recentEventIDs {
			if recent.seenAt.Before(cutoff) {
				delete(h.recentEventIDs, id)
			}
		}
		for len(h.recentEventIDs) >= customerServiceRecentEventLimit {
			var oldestID string
			var oldestAt time.Time
			for id, recent := range h.recentEventIDs {
				if oldestID == "" || recent.seenAt.Before(oldestAt) {
					oldestID = id
					oldestAt = recent.seenAt
				}
			}
			if oldestID == "" {
				break
			}
			delete(h.recentEventIDs, oldestID)
		}
	}
	h.recentEventIDs[event.EventID] = customerServiceRecentEvent{
		seenAt:   now,
		streamID: event.StreamID,
	}
	return true
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

func nonBlockingCustomerServiceEventSend(subscriber chan CustomerServiceRealtimeEvent, event CustomerServiceRealtimeEvent) bool {
	select {
	case subscriber <- event:
		return true
	default:
		// Realtime is an acceleration layer. Dropping an overloaded subscriber is
		// safer than blocking the HTTP write path; clients reconcile through HTTP.
		return false
	}
}

func recordCustomerServiceHubDelivery(scope string, delivered bool) {
	result := "dropped"
	if delivered {
		result = "delivered"
	}
	metrics.CustomerServiceRealtimeHubDeliveries.WithLabelValues(scope, result).Inc()
}
