package service

import (
	"errors"
	"tanzanite/internal/domain/ticket"
)

func (s *TicketService) AddMessage(m *ticket.TicketMessage, userID uint, isStaff bool) error {
	t, err := s.getRegularTicket(m.TicketID)
	if err != nil {
		return err
	}

	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	m.UserID = userID
	m.IsStaff = isStaff

	if err := s.ticketRepo.CreateTicketMessage(m); err != nil {
		return err
	}

	if t.Status == "closed" {
		if err := s.updateTicketStatus(t.ID, "open"); err != nil {
			return err
		}
	}

	return nil
}

func (s *TicketService) GetMessages(ticketID uint, userID uint, isStaff bool) ([]ticket.TicketMessage, error) {
	t, err := s.getRegularTicket(ticketID)
	if err != nil {
		return nil, err
	}

	if !isStaff && t.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.ticketRepo.FindMessagesByTicketID(ticketID)
}

func (s *TicketService) DeleteMessage(id uint, userID uint, isStaff bool) error {
	m, err := s.ticketRepo.FindMessageByID(id)
	if err != nil {
		return err
	}

	if !isStaff && m.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.DeleteTicketMessage(id)
}

func (s *TicketService) CountUnreadMessages(ticketID uint, isStaff bool) (int64, error) {
	return s.ticketRepo.CountUnreadMessages(ticketID, isStaff)
}

func (s *TicketService) MarkMessagesAsRead(ticketID uint, isStaff bool) error {
	return s.ticketRepo.MarkMessagesAsRead(ticketID, isStaff)
}

func (s *TicketService) MarkRegularMessagesAsRead(ticketID uint, isStaff bool) error {
	if _, err := s.getRegularTicket(ticketID); err != nil {
		return err
	}
	return s.MarkMessagesAsRead(ticketID, isStaff)
}
