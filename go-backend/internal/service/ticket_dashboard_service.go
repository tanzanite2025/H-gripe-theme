package service

import "tanzanite/internal/domain/ticket"

func (s *TicketService) GetDashboard() (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	stats, err := s.ticketRepo.GetTicketStats(0)
	if err != nil {
		return nil, err
	}
	dashboard["stats"] = stats

	totalTickets := stats["open"] + stats["in_progress"] + stats["resolved"] + stats["closed"]
	dashboard["total_tickets"] = totalTickets
	dashboard["pending_tickets"] = stats["open"] + stats["in_progress"]
	dashboard["urgent_tickets"] = stats["urgent"]

	if totalTickets > 0 {
		dashboard["resolved_rate"] = float64(stats["resolved"]+stats["closed"]) / float64(totalTickets) * 100
	} else {
		dashboard["resolved_rate"] = 0.0
	}

	return dashboard, nil
}

func (s *TicketService) GetRecentTickets(limit int) ([]ticket.Ticket, error) {
	tickets, _, err := s.ticketRepo.FindAllTickets(1, limit, "", "")
	return tickets, err
}
