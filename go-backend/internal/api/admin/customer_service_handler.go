package admin

import (
	"strings"

	"commerce-platform/internal/api/realtime"
	"commerce-platform/internal/service"
)

type TicketHandler struct {
	ticketService                *service.TicketService
	customerServiceContext       *service.CustomerServiceContextService
	customerServiceAnalytics     *service.CustomerServiceAnalyticsService
	customerServiceRetention     *service.CustomerServiceRetentionService
	customerServiceEvents        *service.CustomerServiceEventHub
	mediaService                 *service.MediaService
	conversationAttachmentAccess *service.CustomerServiceConversationMessageAttachmentAccessService
	allowedOrigins               []string
	webSocketLimiter             *realtime.CustomerServiceWebSocketLimiter
}

func (h *TicketHandler) ConfigureCustomerServiceRetention(retention *service.CustomerServiceRetentionService) {
	if h == nil {
		return
	}
	h.customerServiceRetention = retention
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
		ticketService:                ticketService,
		customerServiceContext:       customerServiceContext,
		customerServiceAnalytics:     customerServiceAnalytics,
		customerServiceEvents:        customerServiceEvents,
		mediaService:                 mediaSvc,
		conversationAttachmentAccess: service.NewCustomerServiceConversationMessageAttachmentAccessService(ticketService, mediaSvc),
	}
}

func (h *TicketHandler) ConfigureCustomerServiceWebSocketLimiter(limiter *realtime.CustomerServiceWebSocketLimiter) {
	if h == nil {
		return
	}
	h.webSocketLimiter = limiter
}
