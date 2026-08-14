package service

import (
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
)

const customerServiceTicketCategory = "customer_service"

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

func (s *TicketService) createTicket(t *ticket.Ticket) error {
	t.Status = "open"
	t.Priority = "medium"
	return s.ticketRepo.CreateTicket(t)
}

func (s *TicketService) assignTicket(id, assignedTo uint) error {
	return s.ticketRepo.AssignTicket(id, assignedTo)
}

func (s *TicketService) updateTicketStatus(id uint, status string) error {
	return s.ticketRepo.UpdateTicketStatus(id, status)
}
