package ticket

import "tanzanite/internal/service"

type Handler struct {
	ticketService         *service.TicketService
	visitorProfileService *service.VisitorProfileService
	customerServiceEvents *service.CustomerServiceEventHub
	allowedOrigins        []string
	visitorSecret         []byte
}

type Options struct {
	AllowedOrigins        []string
	VisitorSecret         string
	VisitorProfileService *service.VisitorProfileService
	CustomerServiceEvents *service.CustomerServiceEventHub
}

func NewHandler(ticketService *service.TicketService, opts ...Options) *Handler {
	options := Options{}
	if len(opts) > 0 {
		options = opts[0]
	}
	return &Handler{
		ticketService:         ticketService,
		visitorProfileService: options.VisitorProfileService,
		customerServiceEvents: options.CustomerServiceEvents,
		allowedOrigins:        append([]string(nil), options.AllowedOrigins...),
		visitorSecret:         []byte(options.VisitorSecret),
	}
}
