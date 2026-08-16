package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

type CustomerServiceRealtimeOutboxHandler struct {
	events *CustomerServiceEventHub
	relay  *CustomerServiceRealtimeRelay
}

func NewCustomerServiceRealtimeOutboxHandler(events *CustomerServiceEventHub, relay ...*CustomerServiceRealtimeRelay) *CustomerServiceRealtimeOutboxHandler {
	var realtimeRelay *CustomerServiceRealtimeRelay
	if len(relay) > 0 {
		realtimeRelay = relay[0]
	}
	return &CustomerServiceRealtimeOutboxHandler{events: events, relay: realtimeRelay}
}

func (h *CustomerServiceRealtimeOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.events == nil {
		metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("infrastructure_unavailable").Inc()
		return errors.New("customer-service realtime event hub is not configured")
	}
	if event.EventType != outbox.EventTypeCustomerServiceRealtime {
		metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("unsupported_event").Inc()
		return fmt.Errorf("unsupported customer-service realtime outbox event %s", event.EventType)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	var payload outbox.CustomerServiceRealtimePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("decode_failed").Inc()
		return fmt.Errorf("decode customer-service realtime outbox event: %w", err)
	}
	if payload.Type == "" || payload.EventID == "" || payload.TicketID == 0 {
		metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("invalid_payload").Inc()
		return errors.New("customer-service realtime outbox event is incomplete")
	}

	realtimeEvent := CustomerServiceRealtimeEvent{
		Type:           payload.Type,
		EventID:        payload.EventID,
		Audience:       CustomerServiceRealtimeAudience(payload.Audience),
		TicketID:       payload.TicketID,
		ConversationID: payload.ConversationID,
		OccurredAt:     payload.OccurredAt,
		Actor: CustomerServiceRealtimeActor{
			Kind:      payload.Actor.Kind,
			UserID:    payload.Actor.UserID,
			Anonymous: payload.Actor.Anonymous,
		},
		Payload: decodeCustomerServiceRealtimePayload(payload.Payload),
	}
	if err := validateCustomerServiceRealtimeEvent(realtimeEvent); err != nil {
		metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("invalid_payload").Inc()
		return fmt.Errorf("validate customer-service realtime outbox event: %w", err)
	}
	if h.relay != nil {
		_, err := h.relay.Publish(ctx, realtimeEvent)
		if err != nil {
			metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("relay_failed").Inc()
			return err
		}
		metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("relay_published").Inc()
		return nil
	}
	h.events.Publish(realtimeEvent)
	metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues("local_published").Inc()
	return nil
}

func enqueueCustomerServiceMessageCreatedOutboxEvent(
	repo *repository.OutboxRepository,
	conversation *ticket.Ticket,
	message *ticket.TicketMessage,
	actor CustomerServiceRealtimeActor,
) error {
	if repo == nil {
		return errors.New("customer-service realtime outbox repository is unavailable")
	}
	if conversation == nil || conversation.ID == 0 || message == nil || message.ID == 0 {
		return errors.New("customer-service realtime event requires a persisted conversation and message")
	}

	event := NewCustomerServiceMessageCreatedEvent(
		conversation.ID,
		ticketConversationID(conversation),
		message.ID,
		message.CreatedAt,
		actor,
	)
	return enqueueCustomerServiceRealtimeOutboxEvent(repo, conversation, event)
}

// enqueueCustomerServiceRealtimeOutboxEvent persists the compact event
// envelope together with its business mutation. It is intentionally shared by
// messages and future durable conversation operations so all transports keep
// one event contract and stable event identity.
func enqueueCustomerServiceRealtimeOutboxEvent(
	repo *repository.OutboxRepository,
	conversation *ticket.Ticket,
	event CustomerServiceRealtimeEvent,
) error {
	if repo == nil {
		return errors.New("customer-service realtime outbox repository is unavailable")
	}
	if conversation == nil || conversation.ID == 0 {
		return errors.New("customer-service realtime event requires a persisted conversation")
	}
	if event.Type == "" || event.EventID == "" || event.TicketID == 0 {
		return errors.New("customer-service realtime event is incomplete")
	}
	if !validCustomerServiceRealtimeAudience(event.Audience) {
		return fmt.Errorf("customer-service realtime event has invalid audience %q", event.Audience)
	}
	if event.TicketID != conversation.ID {
		return errors.New("customer-service realtime event ticket does not match conversation")
	}

	occurredAt := event.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	eventPayload, err := customerServiceRealtimePayloadBytes(event.Payload)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(outbox.CustomerServiceRealtimePayload{
		Type:           event.Type,
		EventID:        event.EventID,
		Audience:       string(normalizeCustomerServiceRealtimeAudience(event.Audience)),
		TicketID:       event.TicketID,
		ConversationID: event.ConversationID,
		OccurredAt:     occurredAt,
		Actor: outbox.CustomerServiceRealtimeActor{
			Kind:      event.Actor.Kind,
			UserID:    event.Actor.UserID,
			Anonymous: event.Actor.Anonymous,
		},
		Payload: eventPayload,
	})
	if err != nil {
		return err
	}

	return repo.CreateEvent(&outbox.Event{
		EventKey:      event.EventID,
		EventType:     outbox.EventTypeCustomerServiceRealtime,
		AggregateType: outbox.AggregateTypeCustomerServiceConversation,
		AggregateID:   strconv.FormatUint(uint64(conversation.ID), 10),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   occurredAt,
	})
}

func customerServiceRealtimePayloadBytes(value interface{}) (json.RawMessage, error) {
	if raw, ok := value.(json.RawMessage); ok {
		return raw, nil
	}
	if raw, ok := value.([]byte); ok {
		return json.RawMessage(raw), nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode customer-service realtime event payload: %w", err)
	}
	return raw, nil
}

func decodeCustomerServiceRealtimePayload(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	return payload
}
