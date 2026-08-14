package service

func (s *TicketService) MarkMessagesAsRead(ticketID uint, isStaff bool) error {
	return s.ticketRepo.MarkMessagesAsRead(ticketID, isStaff)
}
