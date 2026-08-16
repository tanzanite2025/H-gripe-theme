package ticket

import (
	"strings"
	"time"

	ticketdomain "commerce-platform/internal/domain/ticket"
	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) publishPublicCustomerServiceMessageCreated(t *ticketdomain.Ticket, messageID uint, occurredAt time.Time, actor service.CustomerServiceRealtimeActor) {
	if h.customerServiceEvents == nil || t == nil || messageID == 0 {
		return
	}
	h.customerServiceEvents.Publish(service.NewCustomerServiceMessageCreatedEvent(
		t.ID,
		publicConversationID(t),
		messageID,
		occurredAt,
		actor,
	))
}

func (h *Handler) publishPublicCustomerServiceEventToAudience(eventType string, t *ticketdomain.Ticket, actor service.CustomerServiceRealtimeActor, audience service.CustomerServiceRealtimeAudience, payload interface{}, eventIDs ...string) {
	if h.customerServiceEvents == nil || t == nil {
		return
	}
	eventID := ""
	if len(eventIDs) > 0 {
		eventID = strings.TrimSpace(eventIDs[0])
	}
	h.customerServiceEvents.Publish(service.NewCustomerServiceRealtimeEventWithIDAndAudience(
		eventID,
		eventType,
		t.ID,
		publicConversationID(t),
		actor,
		audience,
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

func (h *Handler) replayPublicCustomerServiceWebSocketEvents(c *gin.Context) []service.CustomerServiceRealtimeEvent {
	cursor := strings.TrimSpace(c.Query("last_event_id"))
	if cursor == "" {
		return nil
	}

	events, err := h.customerServiceEvents.ReplayAfter(c.Request.Context(), cursor, 0)
	if err != nil {
		// A malformed or expired cursor is never an authorization failure. The
		// browser reconciles through the scoped HTTP history API after reconnect.
		appLogger.Warn("public customer-service websocket replay unavailable", zap.Error(err))
		return nil
	}
	return events
}
