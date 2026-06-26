package ticket

import "tanzanite/internal/service"

type Handler struct {
	ticketService *service.TicketService
}

func NewHandler(ticketService *service.TicketService) *Handler {
	return &Handler{
		ticketService: ticketService,
	}
}
