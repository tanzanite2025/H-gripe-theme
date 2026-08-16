package ticket

import (
	"net/http"
	"strings"
	"time"

	"commerce-platform/internal/api/realtime"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// StreamPublicCustomerServiceWebSocket upgrades one customer-owned
// conversation after applying the same ownership check as the HTTP history
// endpoint. The socket carries invalidations and transient controls only.
func (h *Handler) StreamPublicCustomerServiceWebSocket(c *gin.Context) {
	if h.customerServiceEvents == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "customer_service_events_unavailable"})
		return
	}

	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id_required"})
		return
	}

	owner := h.existingPublicCustomerOwner(c)
	conversation, err := h.ticketService.GetPublicCustomerServiceConversation(conversationID, owner)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	subscription := h.customerServiceEvents.SubscribeConversation(conversation.ID)
	replay := h.replayPublicCustomerServiceWebSocketEvents(c)
	realtime.ServeCustomerServiceWebSocket(c.Writer, c.Request, realtime.CustomerServiceWebSocketOptions{
		CheckOrigin: func(request *http.Request) bool {
			return realtime.CustomerServiceWebSocketOriginAllowed(request, h.allowedOrigins)
		},
		Subscription: subscription,
		Replay:       replay,
		AllowEvent: func(event service.CustomerServiceRealtimeEvent) bool {
			return event.TicketID == conversation.ID &&
				event.DeliversTo(service.CustomerServiceRealtimeAudiencePublic)
		},
		HandleControl: func(control realtime.CustomerServiceWebSocketControl) {
			if control.Type != "typing" {
				return
			}

			isTyping := true
			if control.IsTyping != nil {
				isTyping = *control.IsTyping
			}
			displayName := "Visitor"
			if owner.UserID != nil {
				displayName = "Customer"
			}

			h.publishPublicCustomerServiceEventToAudience(
				service.CustomerServiceEventTyping,
				conversation,
				publicCustomerServiceRealtimeActor(owner),
				service.CustomerServiceRealtimeAudienceBackoffice,
				gin.H{
					"is_typing":    isTyping,
					"display_name": displayName,
					"expires_at":   time.Now().UTC().Add(5 * time.Second),
				},
			)
		},
	})
}
