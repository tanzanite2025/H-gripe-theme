package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const adminCustomerServiceSSEHeartbeatInterval = 25 * time.Second

func (h *TicketHandler) StreamCustomerServiceEvents(c *gin.Context) {
	if h.customerServiceEvents == nil {
		apierror.RespondInternalError(c, errors.New("customer service event hub is not configured"))
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	scope := strings.ToLower(strings.TrimSpace(c.DefaultQuery("scope", "inbox")))
	ticketID, hasTicketID, ok := parseAdminCustomerServiceEventTicketID(c)
	if !ok {
		return
	}

	var subscription *service.CustomerServiceEventSubscription
	allowEvent := h.adminCustomerServiceEventAllowed(agentUserID, canViewAll)
	if scope == "conversation" || hasTicketID {
		if !hasTicketID {
			apierror.RespondBadRequest(c, "conversation_id is required for conversation event stream")
			return
		}
		if _, err := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll); err != nil {
			respondAdminCustomerServiceError(c, err)
			return
		}
		subscription = h.customerServiceEvents.SubscribeConversation(ticketID)
	} else {
		subscription = h.customerServiceEvents.SubscribeInbox()
	}

	streamAdminCustomerServiceSSE(c, subscription, allowEvent)
}

func parseAdminCustomerServiceEventTicketID(c *gin.Context) (uint, bool, bool) {
	rawID := strings.TrimSpace(c.Query("conversation_id"))
	if rawID == "" {
		rawID = strings.TrimSpace(c.Query("ticket_id"))
	}
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

func (h *TicketHandler) publishAdminCustomerServiceEvent(eventType string, conversation *ticketdomain.Ticket, ticketID uint, actor service.CustomerServiceRealtimeActor, payload interface{}) {
	if h.customerServiceEvents == nil {
		return
	}
	conversationID := ""
	if conversation != nil {
		ticketID = conversation.ID
		conversationID = adminCustomerServicePublicConversationID(conversation)
	}
	h.customerServiceEvents.Publish(service.NewCustomerServiceRealtimeEvent(
		eventType,
		ticketID,
		conversationID,
		actor,
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
		if canViewAll {
			return true
		}
		_, err := h.ticketService.GetCustomerServiceConversationForAgent(event.TicketID, agentUserID, false)
		return err == nil
	}
}

func streamAdminCustomerServiceSSE(c *gin.Context, subscription *service.CustomerServiceEventSubscription, allowEvent func(service.CustomerServiceRealtimeEvent) bool) {
	if subscription == nil || subscription.Events() == nil {
		apierror.RespondInternalError(c, errors.New("customer service event subscription is not available"))
		return
	}
	defer subscription.Cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		apierror.RespondInternalError(c, http.ErrNotSupported)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	heartbeat := time.NewTicker(adminCustomerServiceSSEHeartbeatInterval)
	defer heartbeat.Stop()

	c.SSEvent(service.CustomerServiceEventHeartbeat, service.NewCustomerServiceHeartbeatEvent())
	flusher.Flush()

	for {
		select {
		case event, ok := <-subscription.Events():
			if !ok {
				return
			}
			if allowEvent != nil && !allowEvent(event) {
				continue
			}
			c.SSEvent(event.Type, event)
			flusher.Flush()
		case <-heartbeat.C:
			c.SSEvent(service.CustomerServiceEventHeartbeat, service.NewCustomerServiceHeartbeatEvent())
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
