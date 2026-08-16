package admin

import (
	"net/http"
	"strings"
	"time"

	"commerce-platform/internal/api/realtime"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// StreamCustomerServiceWebSocket exposes a staff-only inbox or a single
// authorized conversation. Inbox sockets cannot send typing controls because
// a control frame must never select a target conversation itself.
func (h *TicketHandler) StreamCustomerServiceWebSocket(c *gin.Context) {
	if h.customerServiceEvents == nil {
		apierror.RespondInternalError(c, errCustomerServiceEventsUnavailable)
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	canEdit := adminCustomerServiceCanEdit(c)
	scope := strings.ToLower(strings.TrimSpace(c.DefaultQuery("scope", "inbox")))
	if scope != "inbox" && scope != "conversation" {
		apierror.RespondBadRequest(c, "scope must be inbox or conversation")
		return
	}

	ticketID, hasTicketID, ok := parseAdminCustomerServiceWebSocketTicketID(c)
	if !ok {
		return
	}

	var subscription *service.CustomerServiceEventSubscription
	allowEvent := h.adminCustomerServiceEventAllowed(agentUserID, canViewAll)
	var conversationID uint
	if scope == "conversation" {
		if !hasTicketID {
			apierror.RespondBadRequest(c, "conversation_id is required for conversation websocket")
			return
		}
		if _, err := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll); err != nil {
			respondAdminCustomerServiceError(c, err)
			return
		}
		conversationID = ticketID
		subscription = h.customerServiceEvents.SubscribeConversation(ticketID)
		baseAllowEvent := allowEvent
		allowEvent = func(event service.CustomerServiceRealtimeEvent) bool {
			return event.TicketID == ticketID && baseAllowEvent(event)
		}
	} else {
		if hasTicketID {
			apierror.RespondBadRequest(c, "conversation_id is only valid for conversation websocket")
			return
		}
		subscription = h.customerServiceEvents.SubscribeInbox()
	}

	replay := h.replayAdminCustomerServiceWebSocketEvents(c)
	realtime.ServeCustomerServiceWebSocket(c.Writer, c.Request, realtime.CustomerServiceWebSocketOptions{
		CheckOrigin: func(request *http.Request) bool {
			return realtime.CustomerServiceWebSocketOriginAllowed(request, h.allowedOrigins)
		},
		Subscription: subscription,
		Replay:       replay,
		AllowEvent:   allowEvent,
		HandleControl: func(control realtime.CustomerServiceWebSocketControl) {
			if control.Type != "typing" || conversationID == 0 || !canEdit {
				return
			}

			// Assignments can change while a socket remains open. Re-check the
			// conversation scope before each transient control publication.
			conversation, err := h.ticketService.GetCustomerServiceConversationForAgent(conversationID, agentUserID, canViewAll)
			if err != nil {
				return
			}
			isTyping := true
			if control.IsTyping != nil {
				isTyping = *control.IsTyping
			}

			h.publishAdminCustomerServiceEventToAudience(
				service.CustomerServiceEventTyping,
				conversation,
				conversationID,
				adminCustomerServiceRealtimeActor(agentUserID),
				service.CustomerServiceRealtimeAudiencePublic,
				gin.H{
					"is_typing":    isTyping,
					"display_name": adminCustomerServiceAssigneeName(agentUserID),
					"expires_at":   time.Now().UTC().Add(5 * time.Second),
				},
			)
		},
	})
}

// apierror requires an error value for unavailable infrastructure. Keep it
// package-local so the WebSocket adapter uses the normal admin response shape.
var errCustomerServiceEventsUnavailable = customerServiceWebSocketError("customer service event hub is not configured")

type customerServiceWebSocketError string

func (e customerServiceWebSocketError) Error() string { return string(e) }
