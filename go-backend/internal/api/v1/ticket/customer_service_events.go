package ticket

import (
	"net/http"
	"strings"
	"time"

	ticketdomain "tanzanite/internal/domain/ticket"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

const customerServiceSSEHeartbeatInterval = 25 * time.Second

func (h *Handler) StreamPublicCustomerServiceEvents(c *gin.Context) {
	if h.customerServiceEvents == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "customer_service_events_unavailable"})
		return
	}

	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id_required"})
		return
	}

	t, err := h.ticketService.GetPublicCustomerServiceConversation(conversationID, h.existingPublicCustomerOwner(c))
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	subscription := h.customerServiceEvents.SubscribeConversation(t.ID)
	streamCustomerServiceSSE(c, subscription)
}

func (h *Handler) publishPublicCustomerServiceEvent(eventType string, t *ticketdomain.Ticket, actor service.CustomerServiceRealtimeActor, payload interface{}) {
	if h.customerServiceEvents == nil || t == nil {
		return
	}
	h.customerServiceEvents.Publish(service.NewCustomerServiceRealtimeEvent(
		eventType,
		t.ID,
		publicConversationID(t),
		actor,
		payload,
	))
}

func publicCustomerServiceRealtimeActor(owner service.CustomerServiceOwner) service.CustomerServiceRealtimeActor {
	return service.CustomerServiceRealtimeActor{
		Kind:      "customer",
		UserID:    owner.UserID,
		Anonymous: owner.UserID == nil,
	}
}

func streamCustomerServiceSSE(c *gin.Context, subscription *service.CustomerServiceEventSubscription) {
	if subscription == nil || subscription.Events() == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "customer_service_events_unavailable"})
		return
	}
	defer subscription.Cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming_unsupported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	heartbeat := time.NewTicker(customerServiceSSEHeartbeatInterval)
	defer heartbeat.Stop()

	c.SSEvent(service.CustomerServiceEventHeartbeat, service.NewCustomerServiceHeartbeatEvent())
	flusher.Flush()

	for {
		select {
		case event, ok := <-subscription.Events():
			if !ok {
				return
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
