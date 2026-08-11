package ticket

import "commerce-platform/internal/service"

type Handler struct {
	ticketService         *service.TicketService
	mediaService          *service.MediaService
	visitorProfileService *service.VisitorProfileService
	customerServiceEvents *service.CustomerServiceEventHub
	allowedOrigins        []string
	visitorSecret         []byte
}

type Options struct {
	MediaService          *service.MediaService
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
		mediaService:          options.MediaService,
		visitorProfileService: options.VisitorProfileService,
		customerServiceEvents: options.CustomerServiceEvents,
		allowedOrigins:        append([]string(nil), options.AllowedOrigins...),
		visitorSecret:         []byte(options.VisitorSecret),
	}
}
