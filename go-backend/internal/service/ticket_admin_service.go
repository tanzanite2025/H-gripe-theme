package service

import (
	"errors"
	"tanzanite/internal/domain/ticket"
)

var (
	ErrInvalidTicketStatus   = errors.New("invalid ticket status")
	ErrInvalidTicketPriority = errors.New("invalid ticket priority")
)

type TicketAdminUpdateInput struct {
	Subject    string
	Priority   string
	Status     string
	AssignedTo *uint
}

func (s *TicketService) UpdateAdminTicket(id uint, input TicketAdminUpdateInput) (*ticket.Ticket, error) {
	if input.Priority != "" && !validTicketPriority(input.Priority) {
		return nil, ErrInvalidTicketPriority
	}
	if input.Status != "" && !validTicketStatus(input.Status) {
		return nil, ErrInvalidTicketStatus
	}

	existingTicket, err := s.ticketRepo.FindTicketByID(id)
	if err != nil {
		return nil, err
	}

	assignedChanged := false
	if input.Subject != "" {
		existingTicket.Subject = input.Subject
	}
	if input.Priority != "" {
		existingTicket.Priority = input.Priority
	}
	if input.Status != "" {
		existingTicket.Status = input.Status
	}
	if input.AssignedTo != nil && existingTicket.AssignedTo != *input.AssignedTo {
		existingTicket.AssignedTo = *input.AssignedTo
		assignedChanged = true
	}

	if err := s.ticketRepo.UpdateTicket(existingTicket); err != nil {
		return nil, err
	}

	switch {
	case input.Status != "":
		if err := s.ticketRepo.UpdateTicketStatus(id, input.Status); err != nil {
			return nil, err
		}
		existingTicket.Status = input.Status
	case assignedChanged:
		if err := s.ticketRepo.UpdateTicketStatus(id, "in_progress"); err != nil {
			return nil, err
		}
		existingTicket.Status = "in_progress"
	}

	return existingTicket, nil
}

func (s *TicketService) GetAdminTicketStats() (map[string]interface{}, error) {
	return s.ticketRepo.GetStats()
}

func validTicketStatus(status string) bool {
	switch status {
	case "open", "in_progress", "resolved", "closed":
		return true
	default:
		return false
	}
}

func validTicketPriority(priority string) bool {
	switch priority {
	case "low", "medium", "high", "urgent":
		return true
	default:
		return false
	}
}
