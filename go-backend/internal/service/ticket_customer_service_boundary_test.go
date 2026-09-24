package service

import (
	"commerce-platform/internal/domain/outbox"
	"encoding/json"
	"strconv"
	"strings"
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

func TestCustomerServiceMessagePageUsesStableDatabaseOrderAndTrueTotal(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "page-agent@example.test", "page-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "page-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	createdAt := time.Date(2026, time.September, 11, 10, 0, 0, 0, time.UTC)
	for _, content := range []string{"first", "second", "third"} {
		require.NoError(t, ticketService.ticketRepo.CreateTicketMessage(&ticket.TicketMessage{
			TicketID:  conversation.ID,
			Content:   content,
			CreatedAt: createdAt,
		}))
	}

	messages, total, err := ticketService.GetPublicCustomerServiceMessagesPage(ticketConversationID(conversation), owner, 2, 1)
	require.NoError(t, err)
	assert.EqualValues(t, 3, total)
	require.Len(t, messages, 2)
	assert.Equal(t, "second", messages[0].Content)
	assert.Equal(t, "third", messages[1].Content)
}

func TestCustomerServiceAgentMessagePreservesVideoType(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "agent-video@example.test", "agent-video", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "video-visitor-hash"}

	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	videoMessage := &ticket.TicketMessage{
		TicketID:    conversation.ID,
		Content:     "Product demo video",
		MessageType: "video",
		Metadata:    `{"url":"https://example.test/uploads/demo.mp4"}`,
	}
	require.NoError(t, ticketService.AddCustomerServiceAgentMessage(videoMessage, agent.ID, false))
	assert.Equal(t, "video", videoMessage.MessageType)

	messages, err := ticketService.GetPublicCustomerServiceMessages(ticketConversationID(conversation), owner, 50, 0)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, "video", messages[0].MessageType)
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
	assert.Nil(t, message.UserID)
}

func TestCustomerServiceInboxArchiveIsRecipientScopedAndRecoverable(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "archive-agent@example.test", "archive-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "archive-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	_, _, err = ticketService.AddPublicCustomerServiceMessage(ticketConversationID(conversation), owner, "archive me", agent.ID, "text", "", "")
	require.NoError(t, err)

	mutation, err := ticketService.SetCustomerServiceConversationArchivedForAgent(conversation.ID, agent.ID, false, true)
	require.NoError(t, err)
	require.NotNil(t, mutation)
	assert.Equal(t, CustomerServiceEventInboxStateChanged, mutation.Event.Type)
	assert.Equal(t, CustomerServiceRealtimeAudienceBackoffice, mutation.Event.Audience)
	var archiveEvent outbox.Event
	require.NoError(t, db.Where("event_key = ?", mutation.Event.EventID).First(&archiveEvent).Error)

	var state ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).First(&state).Error)
	require.NotNil(t, state.ArchivedAt)

	inbox, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{View: "inbox"})
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, inbox)

	archived, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{View: "archived"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, archived, 1)
	require.NotNil(t, archived[0].CustomerServiceInboxArchivedAt)

	restoreMutation, err := ticketService.SetCustomerServiceConversationArchivedForAgent(conversation.ID, agent.ID, false, false)
	require.NoError(t, err)
	require.NotNil(t, restoreMutation)
	var restoreEvent outbox.Event
	require.NoError(t, db.Where("event_key = ?", restoreMutation.Event.EventID).First(&restoreEvent).Error)
	state = ticket.CustomerServiceInboxState{}
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).First(&state).Error)
	assert.Nil(t, state.ArchivedAt)
}

func TestCustomerMessageReturnsArchivedConversationToInbox(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "reopen-archive-agent@example.test", "reopen-archive-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "reopen-archive-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	_, _, err = ticketService.AddPublicCustomerServiceMessage(ticketConversationID(conversation), owner, "first", agent.ID, "text", "", "")
	require.NoError(t, err)
	archive := true
	mutation, statusVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{
			Status:                "closed",
			ExpectedStatusVersion: 1,
			Archive:               &archive,
			ReasonCode:            "operator_close_and_archive",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, mutation)
	require.Len(t, mutation.RealtimeEvents(), 2)
	assert.Equal(t, uint(2), statusVersion)

	_, _, err = ticketService.AddPublicCustomerServiceMessage(ticketConversationID(conversation), owner, "customer followed up", agent.ID, "text", "", "")
	require.NoError(t, err)

	var persisted ticket.Ticket
	require.NoError(t, db.First(&persisted, conversation.ID).Error)
	assert.Equal(t, "open", persisted.Status)
	assert.Equal(t, uint(3), persisted.StatusVersion)
	assert.Nil(t, persisted.ClosedAt)
	assert.Nil(t, persisted.ResolvedAt)

	var state ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).First(&state).Error)
	assert.Nil(t, state.ArchivedAt)
	var reopenedEvent outbox.Event
	require.NoError(t, db.Where("event_key = ?", CustomerServiceConversationStatusChangedEventID(conversation.ID, 3)).First(&reopenedEvent).Error)
	var restoredEvents int64
	require.NoError(t, db.Model(&outbox.Event{}).
		Where("aggregate_id = ? AND event_key LIKE ?", strconv.FormatUint(uint64(conversation.ID), 10), "%:restored:%").
		Count(&restoredEvents).Error)
	assert.EqualValues(t, 1, restoredEvents)
	inbox, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{View: "inbox"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, inbox, 1)
}

func TestCustomerServiceCloseArchiveAndReopenRestoreUseOneVersionedTransaction(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "status-agent@example.test", "status-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "status-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	archive := true
	closeMutation, statusVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{
			Status:                "closed",
			ExpectedStatusVersion: 1,
			Archive:               &archive,
			ReasonCode:            "operator_close_and_archive",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, closeMutation)
	require.Len(t, closeMutation.RealtimeEvents(), 2)
	assert.Equal(t, CustomerServiceEventStatusChanged, closeMutation.RealtimeEvents()[0].Type)
	assert.Equal(t, CustomerServiceEventInboxStateChanged, closeMutation.RealtimeEvents()[1].Type)
	assert.Equal(t, uint(2), statusVersion)

	var closed ticket.Ticket
	require.NoError(t, db.First(&closed, conversation.ID).Error)
	assert.Equal(t, "closed", closed.Status)
	assert.Equal(t, uint(2), closed.StatusVersion)
	assert.NotNil(t, closed.ResolvedAt)
	assert.NotNil(t, closed.ClosedAt)

	var archivedState ticket.CustomerServiceInboxState
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).First(&archivedState).Error)
	assert.NotNil(t, archivedState.ArchivedAt)

	restore := false
	reopenMutation, statusVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{
			Status:                "open",
			ExpectedStatusVersion: 2,
			Archive:               &restore,
			ReasonCode:            "operator_reopen",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, reopenMutation)
	require.Len(t, reopenMutation.RealtimeEvents(), 2)
	assert.Equal(t, uint(3), statusVersion)

	var reopened ticket.Ticket
	require.NoError(t, db.First(&reopened, conversation.ID).Error)
	assert.Equal(t, "open", reopened.Status)
	assert.Equal(t, uint(3), reopened.StatusVersion)
	assert.Nil(t, reopened.ResolvedAt)
	assert.Nil(t, reopened.ClosedAt)
	archivedState = ticket.CustomerServiceInboxState{}
	require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).First(&archivedState).Error)
	assert.Nil(t, archivedState.ArchivedAt)

	noopMutation, statusVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{
			Status:                "open",
			ExpectedStatusVersion: 3,
			Archive:               &restore,
			ReasonCode:            "manual_status_change",
		},
	)
	require.NoError(t, err)
	assert.Nil(t, noopMutation)
	assert.Equal(t, uint(3), statusVersion)

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("aggregate_id = ?", strconv.FormatUint(uint64(conversation.ID), 10)).Count(&eventCount).Error)
	assert.EqualValues(t, 5, eventCount)
}

func TestCustomerServiceStatusRejectsMissingReasonCode(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "reason-code-agent@example.test", "reason-code-agent", "support")
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(CustomerServiceOwner{VisitorSessionHash: "reason-code-visitor"}, agent.ID)
	require.NoError(t, err)

	_, _, err = ticketService.UpdateCustomerServiceConversationStatusForAgent(conversation.ID, agent.ID, false, CustomerServiceConversationStatusInput{
		Status:                "resolved",
		ExpectedStatusVersion: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCustomerServiceInvalidStatus)
}

func TestCustomerServiceStatusCommandRejectsStaleVersionAndInvalidTransition(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "status-conflict-agent@example.test", "status-conflict-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "status-conflict-visitor"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	mutation, statusVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{Status: "in_progress", ExpectedStatusVersion: 1, ReasonCode: "manual_status_change"},
	)
	require.NoError(t, err)
	require.NotNil(t, mutation)
	assert.Equal(t, uint(2), statusVersion)

	mutation, statusVersion, err = ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{Status: "closed", ExpectedStatusVersion: 1, ReasonCode: "operator_close_and_archive"},
	)
	assert.ErrorIs(t, err, repository.ErrTicketStatusVersionConflict)
	assert.Nil(t, mutation)
	assert.Zero(t, statusVersion)

	mutation, statusVersion, err = ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{Status: "open", ExpectedStatusVersion: 2, ReasonCode: "operator_reopen"},
	)
	assert.ErrorIs(t, err, ErrCustomerServiceInvalidStatusTransition)
	assert.Nil(t, mutation)
	assert.Zero(t, statusVersion)

	var persisted ticket.Ticket
	require.NoError(t, db.First(&persisted, conversation.ID).Error)
	assert.Equal(t, "in_progress", persisted.Status)
	assert.Equal(t, uint(2), persisted.StatusVersion)
}

func TestCustomerServiceResolveThenClosePreservesResolutionTimestamp(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "resolve-agent@example.test", "resolve-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "resolve-visitor"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)

	_, resolvedVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{Status: "resolved", ExpectedStatusVersion: 1, ReasonCode: "operator_resolve"},
	)
	require.NoError(t, err)
	assert.Equal(t, uint(2), resolvedVersion)

	var resolved ticket.Ticket
	require.NoError(t, db.First(&resolved, conversation.ID).Error)
	require.NotNil(t, resolved.ResolvedAt)
	assert.Nil(t, resolved.ClosedAt)
	resolutionTime := *resolved.ResolvedAt

	_, closedVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{Status: "closed", ExpectedStatusVersion: 2, ReasonCode: "operator_close_and_archive"},
	)
	require.NoError(t, err)
	assert.Equal(t, uint(3), closedVersion)

	var closed ticket.Ticket
	require.NoError(t, db.First(&closed, conversation.ID).Error)
	require.NotNil(t, closed.ResolvedAt)
	require.NotNil(t, closed.ClosedAt)
	assert.Equal(t, resolutionTime, *closed.ResolvedAt)
}

func TestCustomerServiceCloseAndArchiveRollsBackWhenOutboxWriteFails(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "status-rollback-agent@example.test", "status-rollback-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "status-rollback-visitor"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))

	archive := true
	mutation, statusVersion, err := ticketService.UpdateCustomerServiceConversationStatusForAgent(
		conversation.ID,
		agent.ID,
		false,
		CustomerServiceConversationStatusInput{Status: "closed", ExpectedStatusVersion: 1, Archive: &archive},
	)
	require.Error(t, err)
	assert.Nil(t, mutation)
	assert.Zero(t, statusVersion)

	var persisted ticket.Ticket
	require.NoError(t, db.First(&persisted, conversation.ID).Error)
	assert.Equal(t, "open", persisted.Status)
	assert.Equal(t, uint(1), persisted.StatusVersion)
	assert.Nil(t, persisted.ResolvedAt)
	assert.Nil(t, persisted.ClosedAt)

	var stateCount int64
	require.NoError(t, db.Model(&ticket.CustomerServiceInboxState{}).
		Where("ticket_id = ? AND recipient_user_id = ?", conversation.ID, agent.ID).
		Count(&stateCount).Error)
	assert.Zero(t, stateCount)
}

func TestCustomerServiceBulkArchiveIsAtomicAndDeduplicated(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "bulk-archive-agent@example.test", "bulk-archive-agent", "support")
	firstConversationID := "bulk-archive-first"
	secondConversationID := "bulk-archive-second"
	first := ticket.Ticket{
		TicketNumber:       "TK-BULK-ARCHIVE-1",
		UserID:             agent.ID,
		ConversationID:     &firstConversationID,
		VisitorSessionHash: "bulk-archive-first-owner",
		Subject:            "Bulk archive first",
		Category:           customerServiceTicketCategory,
		Status:             "open",
		AssignedTo:         agent.ID,
	}
	second := ticket.Ticket{
		TicketNumber:       "TK-BULK-ARCHIVE-2",
		UserID:             agent.ID,
		ConversationID:     &secondConversationID,
		VisitorSessionHash: "bulk-archive-second-owner",
		Subject:            "Bulk archive second",
		Category:           customerServiceTicketCategory,
		Status:             "open",
		AssignedTo:         agent.ID,
	}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)

	result, err := ticketService.ArchiveCustomerServiceConversationsForAgent([]uint{second.ID, first.ID, second.ID}, agent.ID, false)
	require.NoError(t, err)
	assert.Equal(t, []uint{first.ID, second.ID}, result.ArchivedConversationIDs)
	require.NotNil(t, result.Mutation)
	require.Len(t, result.Mutation.RealtimeEvents(), 2)

	for _, conversationID := range []uint{first.ID, second.ID} {
		var state ticket.CustomerServiceInboxState
		require.NoError(t, db.Where("ticket_id = ? AND recipient_user_id = ?", conversationID, agent.ID).First(&state).Error)
		assert.NotNil(t, state.ArchivedAt)
	}

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&eventCount).Error)
	assert.EqualValues(t, 2, eventCount)
}

func TestCustomerServiceDefaultInboxExcludesClosedHistory(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "closed-view-agent@example.test", "closed-view-agent", "support")
	owner := CustomerServiceOwner{VisitorSessionHash: "closed-view-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	_, _, err = ticketService.AddPublicCustomerServiceMessage(ticketConversationID(conversation), owner, "close this history", agent.ID, "text", "", "")
	require.NoError(t, err)
	closedAt := time.Now().UTC()
	require.NoError(t, db.Model(&ticket.Ticket{}).Where("id = ?", conversation.ID).Updates(map[string]interface{}{
		"status":    "closed",
		"closed_at": closedAt,
	}).Error)

	inbox, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{})
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, inbox)

	closed, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{View: "closed"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, closed, 1)
	assert.Equal(t, "closed", closed[0].Status)
}

func TestCustomerServiceDefaultInboxIncludesLegacyOpenAndActiveAliases(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	agent := createTicketBoundaryUser(t, db, "legacy-status-agent@example.test", "legacy-status-agent", "support")

	owner := CustomerServiceOwner{VisitorSessionHash: "legacy-status-visitor-hash"}
	conversation, err := ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agent.ID)
	require.NoError(t, err)
	_, _, err = ticketService.AddPublicCustomerServiceMessage(ticketConversationID(conversation), owner, "legacy status message", agent.ID, "text", "", "")
	require.NoError(t, err)

	for _, status := range []string{"pending", "active"} {
		require.NoError(t, db.Model(&ticket.Ticket{}).Where("id = ?", conversation.ID).Update("status", status).Error)
		inbox, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{View: "inbox"})
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		assert.Len(t, inbox, 1)
	}

	inbox, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, agent.ID, false, CustomerServiceConversationListInput{View: "inbox"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	assert.Len(t, inbox, 1)
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
	require.NoError(t, db.Where("event_type = ? AND event_key LIKE ?", outbox.EventTypeCustomerServiceRealtime, "customer_service.message.created:%").Order("id ASC").Find(&events).Error)
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
	require.Len(t, events, 3)

	var payload outbox.CustomerServiceRealtimePayload
	var readEvent outbox.Event
	for _, event := range events {
		if event.EventKey == mutation.Event.EventID {
			readEvent = event
			break
		}
	}
	require.NotEmpty(t, readEvent.EventKey)
	require.NoError(t, json.Unmarshal(readEvent.Payload, &payload))
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
	assert.EqualValues(t, 3, eventCount)
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
	assert.EqualValues(t, 7, eventCount)
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
	require.Len(t, toB.RealtimeEvents(), 2)
	assert.Equal(t, CustomerServiceEventStatusChanged, toB.RealtimeEvents()[1].Type)

	toA, err := ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(conversation.ID, agentB.ID, false, agentA.ID)
	require.NoError(t, err)
	require.NotNil(t, toA)
	assert.Equal(t, CustomerServiceConversationAssignedEventID(conversation.ID, agentA.ID, 2), toA.Event.EventID)

	var events []outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceRealtime).Order("id ASC").Find(&events).Error)
	require.Len(t, events, 5)

	var assignmentPayload outbox.CustomerServiceRealtimePayload
	var assignmentEvent outbox.Event
	for _, event := range events {
		if strings.Contains(event.EventKey, "conversation.assigned") && strings.Contains(event.EventKey, ":2:1") {
			assignmentEvent = event
			break
		}
	}
	require.NotEmpty(t, assignmentEvent.EventKey)
	require.NoError(t, json.Unmarshal(assignmentEvent.Payload, &assignmentPayload))
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
	require.Len(t, events, 2)
	var statusEvent outbox.Event
	for _, event := range events {
		if event.EventKey == mutation.Event.EventID {
			statusEvent = event
			break
		}
	}
	require.NotEmpty(t, statusEvent.EventKey)

	var payload outbox.CustomerServiceRealtimePayload
	require.NoError(t, json.Unmarshal(statusEvent.Payload, &payload))
	assert.Equal(t, CustomerServiceEventStatusChanged, payload.Type)
	assert.Equal(t, string(CustomerServiceRealtimeAudienceBackoffice), payload.Audience)
	assert.Equal(t, mutation.Event.EventID, payload.EventID)

	var statusPayload struct {
		PreviousStatus string `json:"previous_status"`
		Status         string `json:"status"`
		StatusVersion  uint   `json:"status_version"`
		ReasonCode     string `json:"reason_code"`
	}
	require.NoError(t, json.Unmarshal(payload.Payload, &statusPayload))
	assert.Equal(t, "open", statusPayload.PreviousStatus)
	assert.Equal(t, "in_progress", statusPayload.Status)
	assert.Equal(t, uint(2), statusPayload.StatusVersion)
	assert.Equal(t, "same_owner_transfer", statusPayload.ReasonCode)

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
	assert.EqualValues(t, 2, eventCount)
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
	// Conversation creation is a separate durable event; the failed message
	// transaction must not add any additional event.
	assert.EqualValues(t, 1, eventCount)
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
		UserID:   &customer.ID,
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
		UserID:   &agentB.ID,
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

	openChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{Status: "open"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, openChats, 1)
	assert.Equal(t, memberChat.ID, openChats[0].ID)

	inProgressChats, total, err := ticketService.ListCustomerServiceConversationsForAgent(1, 20, 0, true, CustomerServiceConversationListInput{Status: "in_progress"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, inProgressChats, 1)
	assert.Equal(t, anonymousChat.ID, inProgressChats[0].ID)

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
		UserID:     &agent.ID,
		Content:    "customer message before the day",
		CreatedAt:  start.Add(-30 * time.Minute),
		IsStaff:    false,
		IsInternal: false,
	}))
	require.NoError(t, ticketService.ticketRepo.CreateTicketMessage(&ticket.TicketMessage{
		TicketID:   conversation.ID,
		UserID:     &agent.ID,
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
