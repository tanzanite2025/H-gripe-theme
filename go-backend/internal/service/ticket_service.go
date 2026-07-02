package service

import (
	"errors"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/repository"
)

type TicketService struct {
	ticketRepo *repository.TicketRepository
	userRepo   *repository.UserRepository
}

func NewTicketService(ticketRepo *repository.TicketRepository, userRepo *repository.UserRepository) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}

func (s *TicketService) CreateTicket(t *ticket.Ticket) error {
	t.Status = "open"
	t.Priority = "medium"
	return s.ticketRepo.CreateTicket(t)
}

func (s *TicketService) GetTicket(id uint, userID uint, isStaff bool) (*ticket.Ticket, error) {
	t, err := s.ticketRepo.FindTicketByID(id)
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
	existing, err := s.ticketRepo.FindTicketByID(t.ID)
	if err != nil {
		return err
	}

	if !isStaff && existing.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.UpdateTicket(t)
}

func (s *TicketService) UpdateTicketStatus(id uint, status string) error {
	validStatuses := []string{"open", "in_progress", "resolved", "closed"}
	isValid := false
	for _, candidate := range validStatuses {
		if candidate == status {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("invalid status")
	}

	return s.ticketRepo.UpdateTicketStatus(id, status)
}

func (s *TicketService) AssignTicket(id, assignedTo uint) error {
	if err := s.ticketRepo.AssignTicket(id, assignedTo); err != nil {
		return err
	}

	return s.ticketRepo.UpdateTicketStatus(id, "in_progress")
}

func (s *TicketService) CloseTicket(id uint, userID uint, isStaff bool) error {
	t, err := s.ticketRepo.FindTicketByID(id)
	if err != nil {
		return err
	}

	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	if t.Status != "resolved" {
		return errors.New("only resolved tickets can be closed")
	}

	return s.ticketRepo.UpdateTicketStatus(id, "closed")
}

func (s *TicketService) DeleteTicket(id uint, userID uint, isStaff bool) error {
	t, err := s.ticketRepo.FindTicketByID(id)
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
