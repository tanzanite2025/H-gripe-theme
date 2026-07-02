package service

import (
	"errors"
	"strings"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrCustomerServiceConversationAccessDenied = errors.New("conversation access denied")
	ErrCustomerServiceOwnerRequired            = errors.New("conversation owner is required")
)

type CustomerServiceOwner struct {
	UserID             *uint
	VisitorSessionHash string
}

func (s *TicketService) GetCustomerServiceConversations(page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindCustomerServiceConversations(page, pageSize)
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
		Category:           "customer_service",
		Priority:           "medium",
		Status:             "open",
		AssignedTo:         agentID,
		Tags:               customerServiceConversationTag(conversationID),
	}
	if err := s.CreateTicket(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) AddPublicCustomerServiceMessage(conversationID string, owner CustomerServiceOwner, message string, agentID uint) (*ticket.Ticket, *ticket.TicketMessage, error) {
	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return nil, nil, err
	}

	persistedUserID := t.UserID
	if owner.UserID != nil && *owner.UserID > 0 {
		persistedUserID = *owner.UserID
	}

	msg := &ticket.TicketMessage{
		TicketID:   t.ID,
		UserID:     persistedUserID,
		IsStaff:    false,
		Content:    message,
		IsRead:     false,
		IsInternal: false,
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
	if len(agents) == 0 {
		return 0, errors.New("no customer service agents configured")
	}
	if agents[0].UserID == nil {
		return 0, errors.New("customer service agent is not linked to a Go user")
	}
	return *agents[0].UserID, nil
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

func customerServiceConversationTag(conversationID string) string {
	return "conversation_id:" + conversationID
}
