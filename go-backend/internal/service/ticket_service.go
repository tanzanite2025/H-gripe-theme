package service

import (
	"errors"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
)

const customerServiceTicketCategory = "customer_service"

var ErrTicketRouteMismatch = errors.New("ticket must be handled by its dedicated route")

type TicketService struct {
	ticketRepo *repository.TicketRepository
	userRepo   *repository.UserRepository
	faqRepo    *repository.FAQRepository
}

func NewTicketService(ticketRepo *repository.TicketRepository, userRepo *repository.UserRepository, faqRepos ...*repository.FAQRepository) *TicketService {
	service := &TicketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
	if len(faqRepos) > 0 {
		service.faqRepo = faqRepos[0]
	}
	return service
}

func (s *TicketService) CreateTicket(t *ticket.Ticket) error {
	if isCustomerServiceTicket(t) {
		return ErrTicketRouteMismatch
	}
	return s.createTicket(t)
}

func (s *TicketService) createTicket(t *ticket.Ticket) error {
	t.Status = "open"
	t.Priority = "medium"
	return s.ticketRepo.CreateTicket(t)
}

func (s *TicketService) GetTicket(id uint, userID uint, isStaff bool) (*ticket.Ticket, error) {
	t, err := s.getRegularTicket(id)
	if err != nil {
		return nil, err
	}

	if !isStaff && t.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if err := s.ticketRepo.MarkMessagesAsRead(id, isStaff); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) GetUserTickets(userID uint, page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindTicketsByUserID(userID, page, pageSize)
}

func (s *TicketService) GetAllTickets(page, pageSize int, status, priority string) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindAllTickets(page, pageSize, status, priority)
}

func (s *TicketService) GetAssignedTickets(assignedTo uint, page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindTicketsByAssignedTo(assignedTo, page, pageSize)
}

func (s *TicketService) UpdateTicket(t *ticket.Ticket, userID uint, isStaff bool) error {
	existing, err := s.getRegularTicket(t.ID)
	if err != nil {
		return err
	}

	if !isStaff && existing.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.UpdateTicket(t)
}

func (s *TicketService) UpdateTicketStatus(id uint, status string) error {
	if !validTicketStatus(status) {
		return ErrInvalidTicketStatus
	}

	t, err := s.getRegularTicket(id)
	if err != nil {
		return err
	}

	return s.updateTicketStatus(t.ID, status)
}

func (s *TicketService) UpdateTicketStatusForUser(id uint, userID uint, isStaff bool, status string) error {
	if !validTicketStatus(status) {
		return ErrInvalidTicketStatus
	}

	t, err := s.getRegularTicket(id)
	if err != nil {
		return err
	}

	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.updateTicketStatus(t.ID, status)
}

func (s *TicketService) AssignTicket(id, assignedTo uint) error {
	t, err := s.getRegularTicket(id)
	if err != nil {
		return err
	}

	if err := s.assignTicket(t.ID, assignedTo); err != nil {
		return err
	}

	return s.updateTicketStatus(t.ID, "in_progress")
}

func (s *TicketService) CloseTicket(id uint, userID uint, isStaff bool) error {
	t, err := s.getRegularTicket(id)
	if err != nil {
		return err
	}

	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	if t.Status != "resolved" {
		return errors.New("only resolved tickets can be closed")
	}

	return s.updateTicketStatus(id, "closed")
}

func (s *TicketService) DeleteTicket(id uint, userID uint, isStaff bool) error {
	t, err := s.getRegularTicket(id)
	if err != nil {
		return err
	}

	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.DeleteTicket(id)
}

func (s *TicketService) GetTicketStats(userID uint) (map[string]int64, error) {
	return s.ticketRepo.GetTicketStats(userID)
}

func (s *TicketService) getRegularTicket(id uint) (*ticket.Ticket, error) {
	t, err := s.ticketRepo.FindTicketByID(id)
	if err != nil {
		return nil, err
	}
	if isCustomerServiceTicket(t) {
		return nil, ErrTicketRouteMismatch
	}
	return t, nil
}

func (s *TicketService) assignTicket(id, assignedTo uint) error {
	return s.ticketRepo.AssignTicket(id, assignedTo)
}

func (s *TicketService) updateTicketStatus(id uint, status string) error {
	return s.ticketRepo.UpdateTicketStatus(id, status)
}

func isCustomerServiceTicket(t *ticket.Ticket) bool {
	return t != nil && t.Category == customerServiceTicketCategory
}
