package service

import (
	"testing"

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

func TestCustomerServiceConversationListFiltersUseBackendSource(t *testing.T) {
	db, ticketService := newTestTicketBoundaryService(t)
	customer := createTicketBoundaryUser(t, db, "member@example.test", "member", "user")
	agentA := createTicketBoundaryUser(t, db, "agent-a@example.test", "agent-a", "support")
	agentB := createTicketBoundaryUser(t, db, "agent-b@example.test", "agent-b", "support")

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
	))

	return db, NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
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
