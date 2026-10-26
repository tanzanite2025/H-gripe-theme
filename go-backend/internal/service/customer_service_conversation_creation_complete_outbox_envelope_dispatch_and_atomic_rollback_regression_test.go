package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestCustomerServiceConversationCreationDispatchesCompleteEnvelopeAndPreservesHistoricalDeadLetter(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "creation-envelope-agent@example.test", "creation-envelope-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "synthetic-creation-envelope-visitor"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	returnedConversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	require.Equal(t, conversation.ID, returnedConversation.ID)

	var pendingEvents []outbox.Event
	require.NoError(t, db.Find(&pendingEvents).Error)
	require.Len(t, pendingEvents, 1, "reopening the same conversation must not create another event")
	createdEvent := pendingEvents[0]
	var envelope outbox.CustomerServiceRealtimePayload
	require.NoError(t, json.Unmarshal(createdEvent.Payload, &envelope))
	require.Equal(t, CustomerServiceConversationCreatedEventID(conversation.ID), envelope.EventID)
	require.Equal(t, createdEvent.EventKey, envelope.EventID)
	require.Equal(t, CustomerServiceEventConversationCreated, envelope.Type)
	require.Equal(t, conversation.ID, envelope.TicketID)
	require.Equal(t, ticketConversationID(conversation), envelope.ConversationID)
	require.Equal(t, outbox.AggregateTypeCustomerServiceConversation, createdEvent.AggregateType)
	require.Equal(t, strconv.FormatUint(uint64(conversation.ID), 10), createdEvent.AggregateID)
	require.Equal(t, string(CustomerServiceRealtimeAudienceBoth), envelope.Audience)
	require.Equal(t, "system", envelope.Actor.Kind)
	require.False(t, envelope.OccurredAt.IsZero())
	var invalidation CustomerServiceConversationCreatedPayload
	require.NoError(t, json.Unmarshal(envelope.Payload, &invalidation))
	require.Equal(t, conversation.Status, invalidation.Status)
	require.Equal(t, agent.ID, invalidation.AssignedTo)

	historicalEvent := outbox.Event{
		EventKey: CustomerServiceConversationCreatedEventID(900000), EventType: outbox.EventTypeCustomerServiceRealtime,
		AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "900000",
		Payload: []byte(`{"status":"open","assigned_to":1}`), Status: outbox.EventStatusDeadLetter,
		Attempts: 10, MaxAttempts: 10, LastError: "customer-service realtime outbox event is incomplete",
	}
	require.NoError(t, db.Create(&historicalEvent).Error)
	var historicalBefore outbox.Event
	require.NoError(t, db.First(&historicalBefore, historicalEvent.ID).Error)

	hub := NewCustomerServiceEventHub()
	inboxSubscription := hub.SubscribeInbox()
	t.Cleanup(inboxSubscription.Cancel)
	conversationSubscription := hub.SubscribeConversation(conversation.ID)
	t.Cleanup(conversationSubscription.Cancel)
	handler := NewCustomerServiceRealtimeOutboxHandler(hub)
	dispatcher := NewOutboxService(repository.NewOutboxRepository(db))
	dispatcher.RegisterHandler(outbox.EventTypeCustomerServiceRealtime, handler.Handle)
	result, err := dispatcher.ProcessPending(context.Background(), time.Now().UTC().Add(time.Second), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Claimed)
	require.Equal(t, 1, result.Processed)
	require.Zero(t, result.Failed)
	require.Zero(t, result.DeadLetter)
	for _, received := range []CustomerServiceRealtimeEvent{
		readCustomerServiceRealtimeEvent(t, inboxSubscription.Events()),
		readCustomerServiceRealtimeEvent(t, conversationSubscription.Events()),
	} {
		require.Equal(t, envelope.EventID, received.EventID)
		require.Equal(t, conversation.ID, received.TicketID)
		require.Equal(t, CustomerServiceEventConversationCreated, received.Type)
		require.Equal(t, CustomerServiceRealtimeAudienceBoth, received.Audience)
		require.Equal(t, map[string]interface{}{"status": conversation.Status, "assigned_to": float64(agent.ID)}, received.Payload)
	}
	var deliveredEvent outbox.Event
	require.NoError(t, db.First(&deliveredEvent, createdEvent.ID).Error)
	require.Equal(t, outbox.EventStatusProcessed, deliveredEvent.Status)
	require.NotNil(t, deliveredEvent.ProcessedAt)
	require.NoError(t, handler.Handle(context.Background(), deliveredEvent))
	assertNoCustomerServiceRealtimeEvent(t, inboxSubscription.Events())
	assertNoCustomerServiceRealtimeEvent(t, conversationSubscription.Events())
	secondResult, err := dispatcher.ProcessPending(context.Background(), time.Now().UTC().Add(time.Second), 10)
	require.NoError(t, err)
	require.Zero(t, secondResult.Claimed)
	var historicalAfter outbox.Event
	require.NoError(t, db.First(&historicalAfter, historicalEvent.ID).Error)
	require.Equal(t, historicalBefore, historicalAfter, "fixing new events must never rewrite or replay a historical failure")
}

func TestCustomerServiceConversationCreationRollsBackWhenCompleteEventCannotBePersisted(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "creation-rollback-agent@example.test", "creation-rollback-agent", "support")
	require.NoError(t, db.Exec(`CREATE TRIGGER reject_synthetic_conversation_event BEFORE INSERT ON outbox_events BEGIN SELECT RAISE(ABORT, 'synthetic outbox write failure'); END`).Error)
	_, err := ticketService.GetOrCreatePublicCustomerServiceConversation(CustomerServiceOwner{VisitorSessionHash: "synthetic-creation-rollback-visitor"}, agent.ID)
	require.Error(t, err)
	for _, model := range []interface{}{&ticket.Ticket{}, &outbox.Event{}} {
		var count int64
		require.NoError(t, db.Model(model).Count(&count).Error)
		require.Zero(t, count, "conversation and event must either both commit or both roll back")
	}
}

func TestCustomerServiceConversationCreationEnvelopeHandlesUnassignedAndGeneralTickets(t *testing.T) {
	for _, category := range []string{customerServiceTicketCategory, "order"} {
		t.Run(category, func(t *testing.T) {
			db, ticketService := newTestTicketBoundaryService(t)
			customer := createTicketBoundaryUser(t, db, "creation-category-customer@example.test", "creation-category-customer", "customer")
			conversation := ticket.Ticket{UserID: customer.ID, Subject: "synthetic ticket", Category: category}
			require.NoError(t, ticketService.createTicket(&conversation))
			var events []outbox.Event
			require.NoError(t, db.Find(&events).Error)
			if category != customerServiceTicketCategory {
				require.Empty(t, events)
				return
			}
			require.Len(t, events, 1)
			hub := NewCustomerServiceEventHub()
			subscription := hub.SubscribeInbox()
			t.Cleanup(subscription.Cancel)
			require.NoError(t, NewCustomerServiceRealtimeOutboxHandler(hub).Handle(context.Background(), events[0]))
			received := readCustomerServiceRealtimeEvent(t, subscription.Events())
			require.Equal(t, conversation.ID, received.TicketID)
			require.Empty(t, received.ConversationID)
			require.Equal(t, map[string]interface{}{"status": "open"}, received.Payload)
		})
	}
}
