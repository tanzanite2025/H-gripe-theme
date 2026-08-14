package admin

import "commerce-platform/internal/service"

type TicketHandler struct {
	ticketService          *service.TicketService
	customerServiceContext *service.CustomerServiceContextService
	customerServiceEvents  *service.CustomerServiceEventHub
	mediaService           *service.MediaService
}

func NewTicketHandler(ticketService *service.TicketService, customerServiceContext *service.CustomerServiceContextService, customerServiceEvents *service.CustomerServiceEventHub, mediaService ...*service.MediaService) *TicketHandler {
	var mediaSvc *service.MediaService
	if len(mediaService) > 0 {
		mediaSvc = mediaService[0]
	}
	return &TicketHandler{
		ticketService:          ticketService,
		customerServiceContext: customerServiceContext,
		customerServiceEvents:  customerServiceEvents,
		mediaService:           mediaSvc,
	}
}
