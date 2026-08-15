package service

import (
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCustomerServiceConversationAccessDenied = errors.New("conversation access denied")
	ErrCustomerServiceOwnerRequired            = errors.New("conversation owner is required")
	ErrCustomerServiceAgentAccessDenied        = errors.New("agent conversation access denied")
)

type CustomerServiceOwner struct {
	UserID             *uint
	VisitorSessionHash string
}

type CustomerServiceConversationListInput struct {
	AssignedTo *uint
	GroupID    *uint
	Status     string
	UnreadOnly bool
	Identity   string
	Search     string
}

func (s *TicketService) GetCustomerServiceConversations(page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ListCustomerServiceConversationsForAgent(page, pageSize, 0, true, CustomerServiceConversationListInput{})
}

func (s *TicketService) GetCustomerServiceConversationsForAgent(page, pageSize int, agentUserID uint, canViewAll bool) ([]ticket.Ticket, int64, error) {
	return s.ListCustomerServiceConversationsForAgent(page, pageSize, agentUserID, canViewAll, CustomerServiceConversationListInput{})
}

func (s *TicketService) ListCustomerServiceConversationsForAgent(page, pageSize int, agentUserID uint, canViewAll bool, input CustomerServiceConversationListInput) ([]ticket.Ticket, int64, error) {
	filters := repository.CustomerServiceConversationFilters{
		AssignedTo: input.AssignedTo,
		GroupID:    input.GroupID,
		Status:     input.Status,
		UnreadOnly: input.UnreadOnly,
		Identity:   input.Identity,
		Search:     input.Search,
	}

	if !canViewAll {
		if agentUserID == 0 {
			return nil, 0, ErrCustomerServiceAgentAccessDenied
		}
		filters.AssignedTo = &agentUserID
	}

	return s.ticketRepo.FindCustomerServiceConversations(page, pageSize, filters)
}

func (s *TicketService) ListCustomerServiceConversationsInWindowForAgent(start, end time.Time, agentUserID uint, canViewAll bool) ([]ticket.Ticket, error) {
	filters := repository.CustomerServiceConversationFilters{}
	if !canViewAll {
		if agentUserID == 0 {
			return nil, ErrCustomerServiceAgentAccessDenied
		}
		filters.AssignedTo = &agentUserID
	}
	return s.ticketRepo.FindCustomerServiceConversationsInWindow(start, end, filters)
}

func (s *TicketService) GetCustomerServiceMessagesForAgent(ticketID uint, agentUserID uint, canViewAll bool) ([]ticket.TicketMessage, error) {
	if _, err := s.getAgentAccessibleCustomerServiceConversation(ticketID, agentUserID, canViewAll); err != nil {
		return nil, err
	}
	return s.ticketRepo.FindMessagesByTicketID(ticketID)
}

func (s *TicketService) GetCustomerServiceConversationForAgent(ticketID uint, agentUserID uint, canViewAll bool) (*ticket.Ticket, error) {
	return s.getAgentAccessibleCustomerServiceConversation(ticketID, agentUserID, canViewAll)
}

func (s *TicketService) AddCustomerServiceAgentMessage(m *ticket.TicketMessage, agentUserID uint, canViewAll bool) error {
	if m == nil {
		return errors.New("message is required")
	}
	t, err := s.getAgentAccessibleCustomerServiceConversation(m.TicketID, agentUserID, canViewAll)
	if err != nil {
		return err
	}

	m.UserID = agentUserID
	m.IsStaff = true
	if strings.TrimSpace(m.MessageType) == "" {
		m.MessageType = "text"
	}
	if err := s.ticketRepo.CreateTicketMessage(m); err != nil {
		return err
	}
	if t.Status == "closed" {
		return s.updateTicketStatus(t.ID, "in_progress")
	}
	if t.Status == "open" || t.Status == "" {
		return s.updateTicketStatus(t.ID, "in_progress")
	}
	return nil
}

func (s *TicketService) MarkCustomerServiceMessagesReadForAgent(ticketID uint, agentUserID uint, canViewAll bool) error {
	if _, err := s.getAgentAccessibleCustomerServiceConversation(ticketID, agentUserID, canViewAll); err != nil {
		return err
	}
	return s.MarkMessagesAsRead(ticketID, true)
}

func (s *TicketService) TransferCustomerServiceConversationForAgent(ticketID uint, fromAgentUserID uint, canViewAll bool, toAgentUserID uint) error {
	if toAgentUserID == 0 {
		return ErrCustomerServiceAgentAccessDenied
	}
	if _, err := s.getAgentAccessibleCustomerServiceConversation(ticketID, fromAgentUserID, canViewAll); err != nil {
		return err
	}
	if err := s.assignTicket(ticketID, toAgentUserID); err != nil {
		return err
	}
	return s.updateTicketStatus(ticketID, "in_progress")
}

func (s *TicketService) getAgentAccessibleCustomerServiceConversation(ticketID uint, agentUserID uint, canViewAll bool) (*ticket.Ticket, error) {
	if ticketID == 0 {
		return nil, ErrCustomerServiceAgentAccessDenied
	}

	t, err := s.ticketRepo.FindTicketByID(ticketID)
	if err != nil {
		return nil, err
	}
	if t.Category != customerServiceTicketCategory {
		return nil, ErrCustomerServiceAgentAccessDenied
	}
	if canViewAll {
		return t, nil
	}
	if agentUserID > 0 && t.AssignedTo == agentUserID {
		return t, nil
	}
	return nil, ErrCustomerServiceAgentAccessDenied
}

func (s *TicketService) HasPublicCustomerServiceConversation(owner CustomerServiceOwner) (bool, string, uint, error) {
	t, err := s.findCustomerServiceConversationByOwner(owner)
	if repository.IsRecordNotFound(err) {
		return false, "", 0, nil
	}
	if err != nil {
		return false, "", 0, err
	}
	return true, ticketConversationID(t), t.AssignedTo, nil
}

func (s *TicketService) GetOrCreatePublicCustomerServiceConversation(owner CustomerServiceOwner, agentID uint) (*ticket.Ticket, error) {
	owner = normalizeCustomerServiceOwner(owner)
	if !owner.Valid() {
		return nil, ErrCustomerServiceOwnerRequired
	}

	t, err := s.findCustomerServiceConversationByOwner(owner)
	if err == nil {
		if err := s.updateCustomerServiceConversationOwner(t, owner, agentID); err != nil {
			return nil, err
		}
		return t, nil
	}
	if !repository.IsRecordNotFound(err) {
		return nil, err
	}

	persistedUserID, err := s.customerServicePersistedUserID(owner.UserID, agentID)
	if err != nil {
		return nil, err
	}

	conversationID := uuid.NewString()
	t = &ticket.Ticket{
		UserID:             persistedUserID,
		CustomerUserID:     owner.UserID,
		ConversationID:     &conversationID,
		VisitorSessionHash: owner.VisitorSessionHash,
		Subject:            "Customer service chat",
		Category:           customerServiceTicketCategory,
		Priority:           "medium",
		Status:             "open",
		AssignedTo:         agentID,
		Tags:               customerServiceConversationTag(conversationID),
	}
	if err := s.createTicket(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) AddPublicCustomerServiceMessage(conversationID string, owner CustomerServiceOwner, message string, agentID uint, messageType string, metadata string, attachments string) (*ticket.Ticket, *ticket.TicketMessage, error) {
	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return nil, nil, err
	}

	persistedUserID := t.UserID
	if owner.UserID != nil && *owner.UserID > 0 {
		persistedUserID = *owner.UserID
	}

	msg := &ticket.TicketMessage{
		TicketID:    t.ID,
		UserID:      persistedUserID,
		IsStaff:     false,
		Content:     message,
		MessageType: normalizeCustomerServiceMessageType(messageType),
		Metadata:    metadata,
		Attachments: attachments,
		IsRead:      false,
		IsInternal:  false,
	}
	if err := s.ticketRepo.CreateTicketMessage(msg); err != nil {
		return nil, nil, err
	}

	return t, msg, nil
}

func (s *TicketService) GetPublicCustomerServiceMessages(conversationID string, owner CustomerServiceOwner, limit, offset int) ([]ticket.TicketMessage, error) {
	t, err := s.getAccessibleCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	messages, err := s.ticketRepo.FindMessagesByTicketID(t.ID)
	if err != nil {
		return nil, err
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset >= len(messages) {
		return []ticket.TicketMessage{}, nil
	}
	end := offset + limit
	if end > len(messages) {
		end = len(messages)
	}
	return messages[offset:end], nil
}

func (s *TicketService) GetPublicCustomerServiceConversation(conversationID string, owner CustomerServiceOwner) (*ticket.Ticket, error) {
	return s.getAccessibleCustomerServiceConversation(conversationID, owner)
}

func (s *TicketService) CanAccessCustomerServiceConversation(conversationID string, owner CustomerServiceOwner) (bool, error) {
	_, err := s.getAccessibleCustomerServiceConversation(conversationID, owner)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrCustomerServiceConversationAccessDenied) || repository.IsRecordNotFound(err) {
		return false, nil
	}
	return false, err
}

func (s *TicketService) getOrCreateAccessibleCustomerServiceConversation(conversationID string, owner CustomerServiceOwner, agentID uint) (*ticket.Ticket, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return s.GetOrCreatePublicCustomerServiceConversation(owner, agentID)
	}

	t, err := s.getAccessibleCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	if err := s.updateCustomerServiceConversationOwner(t, normalizeCustomerServiceOwner(owner), agentID); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TicketService) getAccessibleCustomerServiceConversation(conversationID string, owner CustomerServiceOwner) (*ticket.Ticket, error) {
	owner = normalizeCustomerServiceOwner(owner)
	if strings.TrimSpace(conversationID) == "" || !owner.Valid() {
		return nil, ErrCustomerServiceConversationAccessDenied
	}

	t, err := s.ticketRepo.FindCustomerServiceConversationByConversationID(strings.TrimSpace(conversationID))
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrCustomerServiceConversationAccessDenied
		}
		return nil, err
	}

	if customerServiceOwnerMatches(t, owner) {
		return t, nil
	}
	return nil, ErrCustomerServiceConversationAccessDenied
}

func (s *TicketService) findCustomerServiceConversationByOwner(owner CustomerServiceOwner) (*ticket.Ticket, error) {
	owner = normalizeCustomerServiceOwner(owner)
	if !owner.Valid() {
		return nil, repository.ErrRecordNotFound
	}

	if owner.UserID != nil {
		t, err := s.ticketRepo.FindCustomerServiceConversationByOwner(owner.UserID, "")
		if err == nil {
			return t, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if owner.VisitorSessionHash != "" {
		return s.ticketRepo.FindCustomerServiceConversationByOwner(nil, owner.VisitorSessionHash)
	}

	return nil, repository.ErrRecordNotFound
}

func (s *TicketService) updateCustomerServiceConversationOwner(t *ticket.Ticket, owner CustomerServiceOwner, agentID uint) error {
	changed := false
	if t.ConversationID == nil || strings.TrimSpace(*t.ConversationID) == "" {
		conversationID := uuid.NewString()
		t.ConversationID = &conversationID
		t.Tags = customerServiceConversationTag(conversationID)
		changed = true
	}
	if t.CustomerUserID == nil && owner.UserID != nil {
		t.CustomerUserID = owner.UserID
		changed = true
	}
	if t.VisitorSessionHash == "" && owner.VisitorSessionHash != "" {
		t.VisitorSessionHash = owner.VisitorSessionHash
		changed = true
	}
	if agentID > 0 && t.AssignedTo != agentID {
		t.AssignedTo = agentID
		changed = true
	}
	if t.Status == "" || t.Status == "closed" || t.Status == "resolved" {
		t.Status = "open"
		changed = true
	}
	if t.UserID == 0 {
		persistedUserID, err := s.customerServicePersistedUserID(owner.UserID, agentID)
		if err != nil {
			return err
		}
		t.UserID = persistedUserID
		changed = true
	}
	if !changed {
		return nil
	}
	return s.ticketRepo.UpdateTicket(t)
}

func (s *TicketService) customerServicePersistedUserID(userID *uint, agentID uint) (uint, error) {
	if userID != nil && *userID > 0 {
		return *userID, nil
	}
	if agentID > 0 {
		return agentID, nil
	}

	agents, err := s.ListCustomerServiceAgentProfiles(1)
	if err != nil {
		return 0, err
	}
	if len(agents) > 0 && agents[0].UserID != nil && *agents[0].UserID > 0 {
		return *agents[0].UserID, nil
	}

	// A public visitor message must still be persisted when the optional
	// customer-service profile layer has not been configured yet. The tickets
	// schema requires a real user_id, so fall back to the first active support
	// account while leaving the conversation unassigned.
	fallbackAgents, err := s.ListCustomerServiceAgents(1)
	if err != nil {
		return 0, err
	}
	if len(fallbackAgents) == 0 {
		return 0, errors.New("no active customer service user configured")
	}
	return fallbackAgents[0].ID, nil
}

func normalizeCustomerServiceOwner(owner CustomerServiceOwner) CustomerServiceOwner {
	owner.VisitorSessionHash = strings.TrimSpace(owner.VisitorSessionHash)
	if owner.UserID != nil && *owner.UserID == 0 {
		owner.UserID = nil
	}
	return owner
}

func (owner CustomerServiceOwner) Valid() bool {
	return owner.UserID != nil || strings.TrimSpace(owner.VisitorSessionHash) != ""
}

func customerServiceOwnerMatches(t *ticket.Ticket, owner CustomerServiceOwner) bool {
	if owner.UserID != nil && t.CustomerUserID != nil && *t.CustomerUserID == *owner.UserID {
		return true
	}
	return owner.VisitorSessionHash != "" && t.VisitorSessionHash == owner.VisitorSessionHash
}

func ticketConversationID(t *ticket.Ticket) string {
	if t == nil || t.ConversationID == nil {
		return ""
	}
	return strings.TrimSpace(*t.ConversationID)
}

func (s *TicketService) ListCustomerServiceAgents(limit int) ([]user.User, error) {
	return s.userRepo.FindCustomerServiceAgents(limit)
}

func (s *TicketService) ListCustomerServiceAgentProfiles(limit int) ([]user.AgentProfile, error) {
	return s.userRepo.FindCustomerServiceAgentProfiles(limit)
}

func (s *TicketService) ListCustomerServiceAgentGroups(limit int, includeInactive bool) ([]user.AgentGroup, error) {
	return s.userRepo.FindCustomerServiceAgentGroups(limit, includeInactive)
}

func customerServiceConversationTag(conversationID string) string {
	return "conversation_id:" + conversationID
}

func normalizeCustomerServiceMessageType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "product", "order", "image", "link", "faq", "config_confirm":
		return value
	default:
		return "text"
	}
}
