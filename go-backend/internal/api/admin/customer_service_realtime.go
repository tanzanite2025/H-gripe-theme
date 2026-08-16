package admin

import (
	"strconv"
	"strings"
	"time"

	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/apierror"
	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func parseAdminCustomerServiceWebSocketTicketID(c *gin.Context) (uint, bool, bool) {
	rawID := strings.TrimSpace(c.Query("conversation_id"))
	if rawID == "" {
		return 0, false, true
	}

	id, err := strconv.ParseUint(rawID, 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "Invalid conversation ID")
		return 0, true, false
	}
	return uint(id), true, true
}

func (h *TicketHandler) publishAdminCustomerServiceMessageCreated(conversation *ticketdomain.Ticket, messageID uint, occurredAt time.Time, actor service.CustomerServiceRealtimeActor) {
	if h.customerServiceEvents == nil || conversation == nil || messageID == 0 {
		return
	}
	h.customerServiceEvents.Publish(service.NewCustomerServiceMessageCreatedEvent(
		conversation.ID,
		adminCustomerServicePublicConversationID(conversation),
		messageID,
		occurredAt,
		actor,
	))
}

func (h *TicketHandler) publishAdminCustomerServiceEventToAudience(eventType string, conversation *ticketdomain.Ticket, ticketID uint, actor service.CustomerServiceRealtimeActor, audience service.CustomerServiceRealtimeAudience, payload interface{}, eventIDs ...string) {
	if h.customerServiceEvents == nil {
		return
	}
	conversationID := ""
	if conversation != nil {
		ticketID = conversation.ID
		conversationID = adminCustomerServicePublicConversationID(conversation)
	}
	eventID := ""
	if len(eventIDs) > 0 {
		eventID = strings.TrimSpace(eventIDs[0])
	}
	h.customerServiceEvents.Publish(service.NewCustomerServiceRealtimeEventWithIDAndAudience(
		eventID,
		eventType,
		ticketID,
		conversationID,
		actor,
		audience,
		payload,
	))
}

func adminCustomerServiceRealtimeActor(userID uint) service.CustomerServiceRealtimeActor {
	return service.CustomerServiceRealtimeActor{
		Kind:   "agent",
		UserID: &userID,
	}
}

func adminCustomerServicePublicConversationID(item *ticketdomain.Ticket) string {
	if item == nil {
		return ""
	}
	if item.ConversationID != nil {
		if conversationID := strings.TrimSpace(*item.ConversationID); conversationID != "" {
			return conversationID
		}
	}
	if strings.HasPrefix(item.Tags, "conversation_id:") {
		return strings.TrimSpace(strings.TrimPrefix(item.Tags, "conversation_id:"))
	}
	return ""
}

func (h *TicketHandler) adminCustomerServiceEventAllowed(agentUserID uint, canViewAll bool) func(service.CustomerServiceRealtimeEvent) bool {
	return func(event service.CustomerServiceRealtimeEvent) bool {
		if event.Type == service.CustomerServiceEventHeartbeat || event.TicketID == 0 {
			return true
		}
		if !event.DeliversTo(service.CustomerServiceRealtimeAudienceBackoffice) {
			return false
		}
		if canViewAll {
			return true
		}
		_, err := h.ticketService.GetCustomerServiceConversationForAgent(event.TicketID, agentUserID, false)
		return err == nil
	}
}

func (h *TicketHandler) replayAdminCustomerServiceWebSocketEvents(c *gin.Context) []service.CustomerServiceRealtimeEvent {
	cursor := strings.TrimSpace(c.Query("last_event_id"))
	if cursor == "" {
		return nil
	}

	events, err := h.customerServiceEvents.ReplayAfter(c.Request.Context(), cursor, 0)
	if err != nil {
		// A malformed or expired cursor is never an authorization failure. The
		// browser reconciles through the scoped HTTP APIs after reconnect.
		appLogger.Warn("customer-service websocket replay unavailable", zap.Error(err))
		return nil
	}
	return events
}
