package service

import (
	"commerce-platform/internal/domain/outbox"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerServiceDedicatedPathStillHandlesConversationMessages(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "agent@example.test", "agent", "support")
	require.NoError(t, db.Create(&user.AgentProfile{
		AgentID: "agent-profile",
		UserID:  &agent.ID,
		Name:    "Agent",
		Status:  "active",
	}).Error)

	owner := CustomerServiceOwner{VisitorSessionHash: "signed-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	require.NotNil(t, conversation)
	assert.Equal(t, customerServiceTicketCategory, conversation.Category)

	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(ticketConversationID(conversation), owner, "hello", agent.ID, "text", "", "")
	require.NoError(t, err)
	require.NotNil(t, customerMessage)

	err = ticketService.AddCustomerServiceAgentMessage(&ticket.TicketMessage{
		TicketID: conversation.ID,
		Content:  "reply",
	}, agent.ID, false)
	require.NoError(t, err)

	messages, err := ticketService.GetPublicCustomerServiceMessages(ticketConversationID(conversation), owner, 50, 0)
	require.NoError(t, err)
	require.Len(t, messages, 2)
	assert.False(t, messages[0].IsStaff)
	assert.True(t, messages[1].IsStaff)
}

func TestCustomerServiceConversationFallsBackToActiveSupportUserWithoutProfile(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	fallbackAgent := createTicketBoundaryUser(t, db, "fallback@example.test", "fallback", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "visitor-without-profile"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, 0)
	require.NoError(t, err)
	require.NotNil(t, conversation)
	assert.Equal(t, fallbackAgent.ID, conversation.UserID)
	assert.Equal(t, uint(0), conversation.AssignedTo)

	_, message, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation),
		owner,
		"hello without a configured profile",
		0,
		"text",
		"",
		"",
	)
	require.NoError(t, err)
	require.NotNil(t, message)
	assert.Equal(t, fallbackAgent.ID, message.UserID)
}

func TestCustomerServiceMessagesCreateRealtimeOutboxEvents(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "outbox-agent@example.test", "outbox-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "outbox-visitor-hash"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation),
		owner,
		"customer outbox message",
		agent.ID,
		"text",
		"",
		"",
	)
	require.NoError(t, err)

	agentMessage := &ticket.TicketMessage{TicketID: conversation.ID, Content: "agent outbox reply"}
	require.NoError(t, ticketService.AddCustomerServiceAgentMessage(agentMessage, agent.ID, false))

	var events []outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Order("id ASC").Find(&events).Error)
	require.Len(t, events, 2)

	expectedActors := map[uint]string{
		customerMessage.ID: "customer",
		agentMessage.ID:    "agent",
	}
	for _, event := range events {
		var payload outbox.CustomerServiceRealtimePayload
		require.NoError(t, json.Unmarshal(event.Payload, &payload))

		var messagePayload map[string]uint
		require.NoError(t, json.Unmarshal(payload.Payload, &messagePayload))
		messageID := messagePayload["message_id"]

		assert.Equal(t, CustomerServiceMessageCreatedEventID(messageID), event.EventKey)
		assert.Equal(t, outbox.AggregateTypeCustomerServiceConversation, event.AggregateType)
		assert.Equal(t, strconv.FormatUint(uint64(conversation.ID), 10), event.AggregateID)
		assert.Equal(t, CustomerServiceEventMessageCreated, payload.Type)
		assert.Equal(t, CustomerServiceMessageCreatedEventID(messageID), payload.EventID)
		assert.Equal(t, conversation.ID, payload.TicketID)
		assert.Equal(t, ticketConversationID(conversation), payload.ConversationID)
		assert.Equal(t, expectedActors[messageID], payload.Actor.Kind)
	}
}

func TestCustomerServiceMessageRollsBackWhenRealtimeOutboxWriteFails(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "outbox-failure-agent@example.test", "outbox-failure-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "outbox-failure-visitor"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))
	_, message, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation),
		owner,
		"this must not persist",
		agent.ID,
		"text",
		"",
		"",
	)
	require.Error(t, err)
	require.Nil(t, message)

	var messageCount int64
	require.NoError(t, db.Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&messageCount).Error)
	assert.Zero(t, messageCount)
}

func TestCustomerServiceInboxStatesKeepAdminReadsIndependent(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	supportAgent := createTicketBoundaryUser(t, db, "inbox-support@example.test", "inbox-support", "support")
	adminA := createTicketBoundaryUser(t, db, "inbox-admin-a@example.test", "inbox-admin-a", "admin")
	adminB := createTicketBoundaryUser(t, db, "inbox-admin-b@example.test", "inbox-admin-b", "manager")
	owner := CustomerServiceOwner{VisitorSessionHash: "inbox-independent-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, supportAgent.ID)
	require.NoError(t, err)
	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation),
		owner,
		"Please help with my wheel order",
		supportAgent.ID,
		"text",
		"",
		"",
	)
	require.NoError(t, err)

	var states []ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ?", conversation.ID).Find(&states).Error)
	require.Len(t, states, 1)
	assert.Equal(t, supportAgent.ID, states[0].RecipientUserID)
	assert.Equal(t, 1, states[0].UnreadCount)

	adminAUnread, total, err := ticketService.ListCustomerServiceConversationsForAgent(
		1,
		20,
		adminA.ID,
		true,
		CustomerServiceConversationListInput{UnreadOnly: true},
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, adminAUnread, 1)
	assert.Equal(t, 1, adminAUnread[0].CustomerServiceUnreadCount)

	adminBUnread, total, err := ticketService.ListCustomerServiceConversationsForAgent(
		1,
		20,
		adminB.ID,
		true,
		CustomerServiceConversationListInput{UnreadOnly: true},
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, adminBUnread, 1)
	assert.Equal(t, 1, adminBUnread[0].CustomerServiceUnreadCount)

	require.NoError(t, ticketService.MarkCustomerServiceMessagesReadForAgent(conversation.ID, adminA.ID, true))

	var persistedMessage ticket.TicketMessage
	require.NoError(t, db.First(&persistedMessage, customerMessage.ID).Error)
	assert.False(t, persistedMessage.IsRead)

	var adminAState ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, adminA.ID).First(&adminAState).Error)
	assert.Equal(t, customerMessage.ID, adminAState.LastReadMessageID)
	assert.Zero(t, adminAState.UnreadCount)

	adminAUnread, total, err = ticketService.ListCustomerServiceConversationsForAgent(
		1,
		20,
		adminA.ID,
		true,
		CustomerServiceConversationListInput{UnreadOnly: true},
	)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, adminAUnread)

	adminBUnread, total, err = ticketService.ListCustomerServiceConversationsForAgent(
		1,
		20,
		adminB.ID,
		true,
		CustomerServiceConversationListInput{UnreadOnly: true},
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, adminBUnread, 1)
	assert.Equal(t, 1, adminBUnread[0].CustomerServiceUnreadCount)
}

func TestCustomerServiceReadCreatesRealtimeOutboxEventAndSkipsNoop(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "read-outbox-agent@example.test", "read-outbox-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "read-outbox-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "Please mark this read", agent.ID, "text", "", "",
	)
	require.NoError(t, err)

	mutation, err := ticketService.MarkCustomerServiceMessagesReadForAgentWithRealtimeEvent(conversation.ID, agent.ID, false)
	require.NoError(t, err)
	require.NotNil(t, mutation)
	assert.Equal(t, CustomerServiceEventMessagesRead, mutation.Event.Type)
	assert.Equal(t, CustomerServiceMessagesReadEventID(conversation.ID, agent.ID, 1, customerMessage.ID), mutation.Event.EventID)

	var events []outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Order("id ASC").Find(&events).Error)
	require.Len(t, events, 2)

	var payload outbox.CustomerServiceRealtimePayload
	require.NoError(t, json.Unmarshal(events[1].Payload, &payload))
	assert.Equal(t, CustomerServiceEventMessagesRead, payload.Type)
	assert.Equal(t, mutation.Event.EventID, payload.EventID)
	assert.Equal(t, conversation.ID, payload.TicketID)
	assert.Equal(t, agent.ID, *payload.Actor.UserID)

	var readPayload struct {
		ReaderKind        string `json:"reader_kind"`
		ReadByUserID      uint   `json:"read_by_user_id"`
		AssignmentVersion uint   `json:"assignment_version"`
		LastReadMessageID uint   `json:"last_read_message_id"`
	}
	require.NoError(t, json.Unmarshal(payload.Payload, &readPayload))
	assert.Equal(t, "agent", readPayload.ReaderKind)
	assert.Equal(t, agent.ID, readPayload.ReadByUserID)
	assert.Equal(t, uint(1), readPayload.AssignmentVersion)
	assert.Equal(t, customerMessage.ID, readPayload.LastReadMessageID)

	mutation, err = ticketService.MarkCustomerServiceMessagesReadForAgentWithRealtimeEvent(conversation.ID, agent.ID, false)
	require.NoError(t, err)
	assert.Nil(t, mutation)

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Count(&eventCount).Error)
	assert.EqualValues(t, 2, eventCount)
}

func TestCustomerServiceReadAfterReassignmentUsesNewRealtimeEventID(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agentA := createTicketBoundaryUser(t, db, "reassigned-read-agent-a@example.test", "reassigned-read-agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "reassigned-read-agent-b@example.test", "reassigned-read-agent-b", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "reassigned-read-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentA.ID)
	require.NoError(t, err)
	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "Reassigned agents must get a new read event", agentA.ID, "text", "", "",
	)
	require.NoError(t, err)

	firstRead, err := ticketService.MarkCustomerServiceMessagesReadForAgentWithRealtimeEvent(conversation.ID, agentA.ID, false)
	require.NoError(t, err)
	require.NotNil(t, firstRead)
	require.NoError(t, ticketService.TransferCustomerServiceConversationForAgent(conversation.ID, agentA.ID, false, agentB.ID))
	require.NoError(t, ticketService.TransferCustomerServiceConversationForAgent(conversation.ID, agentB.ID, false, agentA.ID))

	secondRead, err := ticketService.MarkCustomerServiceMessagesReadForAgentWithRealtimeEvent(conversation.ID, agentA.ID, false)
	require.NoError(t, err)
	require.NotNil(t, secondRead)
	assert.Equal(t, CustomerServiceMessagesReadEventID(conversation.ID, agentA.ID, 2, customerMessage.ID), secondRead.Event.EventID)
	assert.NotEqual(t, firstRead.Event.EventID, secondRead.Event.EventID)

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Count(&eventCount).Error)
	assert.EqualValues(t, 5, eventCount)
}

func TestCustomerServiceReadRollsBackWhenRealtimeOutboxWriteFails(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "read-outbox-failure-agent@example.test", "read-outbox-failure-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "read-outbox-failure-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "Do not partially mark me read", agent.ID, "text", "", "",
	)
	require.NoError(t, err)
	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))

	mutation, err := ticketService.MarkCustomerServiceMessagesReadForAgentWithRealtimeEvent(conversation.ID, agent.ID, false)
	require.Error(t, err)
	assert.Nil(t, mutation)

	var state ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).First(&state).Error)
	assert.Zero(t, state.LastReadMessageID)
	assert.Equal(t, 1, state.UnreadCount)
	assert.NotEqual(t, customerMessage.ID, state.LastReadMessageID)
}

func TestCustomerServiceTransferResetsNewAssigneeInboxState(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agentA := createTicketBoundaryUser(t, db, "transfer-agent-a@example.test", "transfer-agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "transfer-agent-b@example.test", "transfer-agent-b", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "transfer-inbox-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentA.ID)
	require.NoError(t, err)
	_, firstMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "First question", agentA.ID, "text", "", "",
	)
	require.NoError(t, err)
	_, secondMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "Second question", agentA.ID, "text", "", "",
	)
	require.NoError(t, err)
	require.Greater(t, secondMessage.ID, firstMessage.ID)
	require.NoError(t, ticketService.MarkCustomerServiceMessagesReadForAgent(conversation.ID, agentA.ID, false))

	require.NoError(t, ticketService.TransferCustomerServiceConversationForAgent(
		conversation.ID,
		agentA.ID,
		false,
		agentB.ID,
	))

	var transferred ticket.Ticket
	require.NoError(t, db.First(&transferred, conversation.ID).Error)
	assert.Equal(t, agentB.ID, transferred.AssignedTo)
	assert.Equal(t, "in_progress", transferred.Status)

	var oldAssigneeState ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agentA.ID).First(&oldAssigneeState).Error)
	assert.Equal(t, secondMessage.ID, oldAssigneeState.LastReadMessageID)
	assert.Zero(t, oldAssigneeState.UnreadCount)

	var newAssigneeState ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agentB.ID).First(&newAssigneeState).Error)
	assert.Zero(t, newAssigneeState.LastReadMessageID)
	assert.Equal(t, 2, newAssigneeState.UnreadCount)

	newAssigneeUnread, total, err := ticketService.ListCustomerServiceConversationsForAgent(
		1,
		20,
		agentB.ID,
		false,
		CustomerServiceConversationListInput{UnreadOnly: true},
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, newAssigneeUnread, 1)
	assert.Equal(t, 2, newAssigneeUnread[0].CustomerServiceUnreadCount)

	oldAssigneeConversations, total, err := ticketService.ListCustomerServiceConversationsForAgent(
		1,
		20,
		agentA.ID,
		false,
		CustomerServiceConversationListInput{},
	)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, oldAssigneeConversations)
}

func TestCustomerServiceTransferCreatesVersionedRealtimeOutboxEvents(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agentA := createTicketBoundaryUser(t, db, "transfer-outbox-agent-a@example.test", "transfer-outbox-agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "transfer-outbox-agent-b@example.test", "transfer-outbox-agent-b", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "transfer-outbox-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentA.ID)
	require.NoError(t, err)
	_, _, err = ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "Please transfer me", agentA.ID, "text", "", "",
	)
	require.NoError(t, err)

	toB, err := ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(conversation.ID, agentA.ID, false, agentB.ID)
	require.NoError(t, err)
	require.NotNil(t, toB)
	assert.Equal(t, CustomerServiceEventAssigned, toB.Event.Type)
	assert.Equal(t, CustomerServiceConversationAssignedEventID(conversation.ID, agentB.ID, 1), toB.Event.EventID)

	toA, err := ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(conversation.ID, agentB.ID, false, agentA.ID)
	require.NoError(t, err)
	require.NotNil(t, toA)
	assert.Equal(t, CustomerServiceConversationAssignedEventID(conversation.ID, agentA.ID, 2), toA.Event.EventID)

	var events []outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Order("id ASC").Find(&events).Error)
	require.Len(t, events, 3)

	var assignmentPayload outbox.CustomerServiceRealtimePayload
	require.NoError(t, json.Unmarshal(events[1].Payload, &assignmentPayload))
	assert.Equal(t, CustomerServiceEventAssigned, assignmentPayload.Type)
	assert.Equal(t, toB.Event.EventID, assignmentPayload.EventID)

	var eventPayload struct {
		AssignedTo        uint   `json:"assigned_to"`
		AssignedByUserID  uint   `json:"assigned_by_user_id"`
		AssignmentVersion uint   `json:"assignment_version"`
		Status            string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(assignmentPayload.Payload, &eventPayload))
	assert.Equal(t, agentB.ID, eventPayload.AssignedTo)
	assert.Equal(t, agentA.ID, eventPayload.AssignedByUserID)
	assert.Equal(t, uint(1), eventPayload.AssignmentVersion)
	assert.Equal(t, "in_progress", eventPayload.Status)
}

func TestCustomerServiceSameOwnerTransferCreatesVersionedStatusRealtimeEvent(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "same-owner-status-agent@example.test", "same-owner-status-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "same-owner-status-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	require.Equal(t, "open", conversation.Status)
	require.Equal(t, uint(1), conversation.StatusVersion)

	mutation, err := ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(
		conversation.ID,
		agent.ID,
		false,
		agent.ID,
	)
	require.NoError(t, err)
	require.NotNil(t, mutation)
	assert.Equal(t, CustomerServiceEventStatusChanged, mutation.Event.Type)
	assert.Equal(t, CustomerServiceRealtimeAudienceBackoffice, mutation.Event.Audience)
	assert.Equal(t, CustomerServiceConversationStatusChangedEventID(conversation.ID, 2), mutation.Event.EventID)

	var persisted ticket.Ticket
	require.NoError(t, db.First(&persisted, conversation.ID).Error)
	assert.Equal(t, agent.ID, persisted.AssignedTo)
	assert.Equal(t, "in_progress", persisted.Status)
	assert.Equal(t, uint(2), persisted.StatusVersion)

	var events []outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Order("id ASC").Find(&events).Error)
	require.Len(t, events, 1)
	assert.Equal(t, mutation.Event.EventID, events[0].EventKey)

	var payload outbox.CustomerServiceRealtimePayload
	require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
	assert.Equal(t, CustomerServiceEventStatusChanged, payload.Type)
	assert.Equal(t, string(CustomerServiceRealtimeAudienceBackoffice), payload.Audience)
	assert.Equal(t, mutation.Event.EventID, payload.EventID)

	var statusPayload struct {
		PreviousStatus string `json:"previous_status"`
		Status         string `json:"status"`
		StatusVersion  uint   `json:"status_version"`
		Reason         string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(payload.Payload, &statusPayload))
	assert.Equal(t, "open", statusPayload.PreviousStatus)
	assert.Equal(t, "in_progress", statusPayload.Status)
	assert.Equal(t, uint(2), statusPayload.StatusVersion)
	assert.Equal(t, "same_owner_transfer", statusPayload.Reason)

	mutation, err = ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(
		conversation.ID,
		agent.ID,
		false,
		agent.ID,
	)
	require.NoError(t, err)
	assert.Nil(t, mutation)

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Count(&eventCount).Error)
	assert.EqualValues(t, 1, eventCount)
}

func TestCustomerServiceSameOwnerTransferRollsBackStatusWhenRealtimeOutboxWriteFails(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "same-owner-status-failure-agent@example.test", "same-owner-status-failure-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "same-owner-status-failure-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))

	mutation, err := ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(
		conversation.ID,
		agent.ID,
		false,
		agent.ID,
	)
	require.Error(t, err)
	assert.Nil(t, mutation)

	var persisted ticket.Ticket
	require.NoError(t, db.First(&persisted, conversation.ID).Error)
	assert.Equal(t, "open", persisted.Status)
	assert.Equal(t, uint(1), persisted.StatusVersion)
}

func TestCustomerServiceTransferRollsBackWhenRealtimeOutboxWriteFails(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agentA := createTicketBoundaryUser(t, db, "transfer-outbox-failure-agent-a@example.test", "transfer-outbox-failure-agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "transfer-outbox-failure-agent-b@example.test", "transfer-outbox-failure-agent-b", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "transfer-outbox-failure-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentA.ID)
	require.NoError(t, err)
	_, _, err = ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "This transfer must roll back", agentA.ID, "text", "", "",
	)
	require.NoError(t, err)
	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))

	mutation, err := ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(conversation.ID, agentA.ID, false, agentB.ID)
	require.Error(t, err)
	assert.Nil(t, mutation)

	var persisted ticket.Ticket
	require.NoError(t, db.First(&persisted, conversation.ID).Error)
	assert.Equal(t, agentA.ID, persisted.AssignedTo)
	assert.Equal(t, "open", persisted.Status)

	var newAssigneeState ticket.CustomerServiceInboxState
	assert.Error(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agentB.ID).First(&newAssigneeState).Error)
}

func TestCustomerServicePublicConversationOwnerUpdateResetsNewAssigneeInboxState(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agentA := createTicketBoundaryUser(t, db, "owner-update-agent-a@example.test", "owner-update-agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "owner-update-agent-b@example.test", "owner-update-agent-b", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "owner-update-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentA.ID)
	require.NoError(t, err)
	_, customerMessage, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation), owner, "Route me again", agentA.ID, "text", "", "",
	)
	require.NoError(t, err)
	require.NoError(t, ticketService.MarkCustomerServiceMessagesReadForAgent(conversation.ID, agentA.ID, false))

	updatedConversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentB.ID)
	require.NoError(t, err)
	assert.Equal(t, conversation.ID, updatedConversation.ID)
	assert.Equal(t, agentB.ID, updatedConversation.AssignedTo)

	var newAssigneeState ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agentB.ID).First(&newAssigneeState).Error)
	assert.Zero(t, newAssigneeState.LastReadMessageID)
	assert.Equal(t, 1, newAssigneeState.UnreadCount)
	assert.NotEqual(t, customerMessage.ID, newAssigneeState.LastReadMessageID)
}

func TestCustomerServiceMessageRollsBackWhenInboxStateWriteFails(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "inbox-write-failure-agent@example.test", "inbox-write-failure-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "inbox-write-failure-visitor"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	require.NoError(t, db.Migrator().DropTable(&ticket.CustomerServiceInboxState{}))

	_, message, err := ticketService.AddPublicCustomerServiceMessage(
		ticketConversationID(conversation),
		owner,
		"this must roll back with the inbox state",
		agent.ID,
		"text",
		"",
		"",
	)
	require.Error(t, err)
	require.Nil(t, message)

	var messageCount int64
	require.NoError(t, db.Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&messageCount).Error)
	assert.Zero(t, messageCount)

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("aggregate_id = ?", strconv.FormatUint(uint64(conversation.ID), 10)).Count(&eventCount).Error)
	assert.Zero(t, eventCount)
}

func TestCustomerServiceConversationListFiltersUseBackendSource(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	customer := createTicketBoundaryUser(t, db, "member@example.test", "member", "user")
	agentA := createTicketBoundaryUser(t, db, "agent-a@example.test", "agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "agent-b@example.test", "agent-b", "support")

	emptyConversationID := "empty-conversation"
	emptyChat := ticket.Ticket{
		TicketNumber:   "TK-FILTER-EMPTY",
		UserID:         agentA.ID,
		ConversationID: &emptyConversationID,
		Subject:        "Empty customer service chat",
		Category:       customerServiceTicketCategory,
		Status:         "open",
		AssignedTo:     agentA.ID,
	}
	require.NoError(t, ticketService.createTicket(&emptyChat))

	memberConversationID := "member-conversation"
	memberChat := ticket.Ticket{
		TicketNumber:   "TK-FILTER-MEMBER",
		UserID:         customer.ID,
		CustomerUserID: &customer.ID,
		ConversationID: &memberConversationID,
		Subject:        "Member customer service chat",
		Category:       customerServiceTicketCategory,
		Status:         "open",
		AssignedTo:     agentA.ID,
	}
	require.NoError(t, ticketService.createTicket(&memberChat))
	require.NoError(t, ticketService.ticketRepo.CreateTicketMessage(&ticket.TicketMessage{
		TicketID: memberChat.ID,
		UserID:   customer.ID,
		Content:  "Need help with a tire order",
		IsStaff:  false,
		IsRead:   false,
	}))

	anonymousConversationID := "anonymous-conversation"
	anonymousChat := ticket.Ticket{
		TicketNumber:       "TK-FILTER-ANON",
		UserID:             agentB.ID,
		ConversationID:     &anonymousConversationID,
		VisitorSessionHash: "visitor-filter-hash",
		Subject:            "Anonymous customer service chat",
		Category:           customerServiceTicketCategory,
		Status:             "in_progress",
		AssignedTo:         agentB.ID,
	}
	require.NoError(t, ticketService.createTicket(&anonymousChat))
	require.NoError(t, ticketService.updateTicketStatus(anonymousChat.ID, "in_progress"))
	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "visitor-filter-hash",
		Email:                      "visitor-filter@example.test",
		CartSessionID:              "cart-filter-session",
	}).Error)
	require.NoError(t, ticketService.ticketRepo.CreateTicketMessage(&ticket.TicketMessage{
		TicketID: anonymousChat.ID,
		UserID:   agentB.ID,
		Content:  "Already handled",
		IsStaff:  false,
		IsRead:   true,
	}))

	allChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{})
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	require.Len(t, allChats, 2)
	assert.NotContains(t, []uint{allChats[0].ID, allChats[1].ID}, emptyChat.ID)

	accountChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{Identity: "account"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, accountChats, 1)
	assert.Equal(t, memberChat.ID, accountChats[0].ID)

	unreadChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{UnreadOnly: true})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, unreadChats, 1)
	assert.Equal(t, memberChat.ID, unreadChats[0].ID)

	visitorChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{Search: "visitor-filter@example.test"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, visitorChats, 1)
	assert.Equal(t, anonymousChat.ID, visitorChats[0].ID)

	pendingChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{Status: "pending"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, pendingChats, 1)
	assert.Equal(t, memberChat.ID, pendingChats[0].ID)

	activeChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{Status: "active"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, activeChats, 1)
	assert.Equal(t, anonymousChat.ID, activeChats[0].ID)

	forcedAssignee := agentB.ID
	scopedChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agentA.ID, false, CustomerServiceConversationListInput{AssignedTo: &forcedAssignee})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, scopedChats, 1)
	assert.Equal(t, memberChat.ID, scopedChats[0].ID)
}

func TestCustomerServiceConversationWindowKeepsLatestMessageBeforeWindow(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "window-agent@example.test", "window-agent", "support")
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	conversationID := "window-conversation"
	conversation := ticket.Ticket{
		TicketNumber:   "TK-WINDOW-BOUNDARY",
		UserID:         agent.ID,
		ConversationID: &conversationID,
		Subject:        "Cross-day reply interval",
		Category:       customerServiceTicketCategory,
		Status:         "in_progress",
		AssignedTo:     agent.ID,
		CreatedAt:      start.Add(-2 * time.Hour),
	}
	require.NoError(t, ticketService.createTicket(&conversation))

	require.NoError(t, ticketService.ticketRepo.CreateTicketMessage(&ticket.TicketMessage{
		TicketID:   conversation.ID,
		UserID:     agent.ID,
		Content:    "customer message before the day",
		CreatedAt:  start.Add(-30 * time.Minute),
		IsStaff:    false,
		IsInternal: false,
	}))
	require.NoError(t, ticketService.ticketRepo.CreateTicketMessage(&ticket.TicketMessage{
		TicketID:   conversation.ID,
		UserID:     agent.ID,
		Content:    "reply during the day",
		CreatedAt:  start.Add(30 * time.Minute),
		IsStaff:    true,
		IsInternal: false,
	}))

	conversations, err := ticketService.ListCustomerServiceConversationsInWindowForAgent(start, end, 0, true)
	require.NoError(t, err)
	require.Len(t, conversations, 1)
	require.Len(t, conversations[0].Messages, 2)
	assert.Equal(t, "customer message before the day", conversations[0].Messages[0].Content)
	assert.Equal(t, "reply during the day", conversations[0].Messages[1].Content)
}

func newTestTicketBoundaryService(t *testing.T) (*gorm.DB, *TicketService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&user.User{},
		&user.AgentProfile{},
		&visitor.Profile{},
		&ticket.Ticket{},
		&ticket.TicketMessage{},
		&ticket.CustomerServiceInboxState{},
		&outbox.Event{},
	))

	ticketService := NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	ticketService.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(db))
	return db, ticketService
}

func createTicketBoundaryUser(t *testing.T, db *gorm.DB, email, username, role string) user.User {
	t.Helper()

	item := user.User{
		Email:    email,
		Username: username,
		Password: "test-password",
		Role:     role,
		Status:   "active",
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}
