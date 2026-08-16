package admin

import (
	"commerce-platform/internal/service"
	"strings"
)

type TicketHandler struct {
	ticketService            *service.TicketService
	customerServiceContext   *service.CustomerServiceContextService
	customerServiceAnalytics *service.CustomerServiceAnalyticsService
	customerServiceEvents    *service.CustomerServiceEventHub
	mediaService             *service.MediaService
	allowedOrigins           []string
}

func (h *TicketHandler) ConfigureAllowedOrigins(allowedOrigins []string) {
	if h == nil {
		return
	}
	h.allowedOrigins = make([]string, 0, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if value := strings.TrimSpace(origin); value != "" {
			h.allowedOrigins = append(h.allowedOrigins, value)
		}
	}
}

func NewTicketHandler(
	ticketService *service.TicketService,
	customerServiceContext *service.CustomerServiceContextService,
	customerServiceAnalytics *service.CustomerServiceAnalyticsService,
	customerServiceEvents *service.CustomerServiceEventHub,
	mediaService ...*service.MediaService,
) *TicketHandler {
	var mediaSvc *service.MediaService
	if len(mediaService) > 0 {
		mediaSvc = mediaService[0]
	}
	return &TicketHandler{
		ticketService:            ticketService,
		customerServiceContext:   customerServiceContext,
		customerServiceAnalytics: customerServiceAnalytics,
		customerServiceEvents:    customerServiceEvents,
		mediaService:             mediaSvc,
	}
}
