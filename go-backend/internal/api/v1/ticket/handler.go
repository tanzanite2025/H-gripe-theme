package ticket

import "tanzanite/internal/service"

type Handler struct {
	ticketService  *service.TicketService
	allowedOrigins []string
	visitorSecret  []byte
}

type Options struct {
	AllowedOrigins []string
	VisitorSecret  string
}

func NewHandler(ticketService *service.TicketService, opts ...Options) *Handler {
	options := Options{}
	if len(opts) > 0 {
		options = opts[0]
	}
	return &Handler{
		ticketService:  ticketService,
		allowedOrigins: append([]string(nil), options.AllowedOrigins...),
		visitorSecret:  []byte(options.VisitorSecret),
	}
}
