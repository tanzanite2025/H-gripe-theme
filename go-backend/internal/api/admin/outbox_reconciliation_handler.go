package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// OutboxReconciliationHandler exposes the deliberately manual transitions for
// events whose external side effect may already have happened. The normal
// worker never resumes an unknown event automatically.
type OutboxReconciliationHandler struct {
	outboxService *service.OutboxService
	auditRecorder adminAuditRecorder
}

func NewOutboxReconciliationHandler(outboxService *service.OutboxService) *OutboxReconciliationHandler {
	return &OutboxReconciliationHandler{outboxService: outboxService}
}

func (h *OutboxReconciliationHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditRecorder = recorder
}

type outboxUnknownEventResponse struct {
	ID             uint       `json:"id"`
	EventKey       string     `json:"event_key"`
	EventType      string     `json:"event_type"`
	AggregateType  string     `json:"aggregate_type"`
	AggregateID    string     `json:"aggregate_id"`
	Status         string     `json:"status"`
	Attempts       int        `json:"attempts"`
	MaxAttempts    int        `json:"max_attempts"`
	AvailableAt    time.Time  `json:"available_at"`
	LastAttemptAt  *time.Time `json:"last_attempt_at,omitempty"`
	UncertainAt    *time.Time `json:"uncertain_at,omitempty"`
	ReconcileAfter *time.Time `json:"reconcile_after,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (h *OutboxReconciliationHandler) ListUnknown(c *gin.Context) {
	if h == nil || h.outboxService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox reconciliation service is unavailable"})
		return
	}

	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 500"})
			return
		}
		limit = parsed
	}

	events, err := h.outboxService.ListUnknownEvents(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list unknown outbox events"})
		return
	}

	items := make([]outboxUnknownEventResponse, 0, len(events))
	for _, event := range events {
		items = append(items, toOutboxUnknownEventResponse(event))
	}
	response.Success(c, gin.H{"events": items, "count": len(items)})
}

type outboxReconciliationRequest struct {
	Note            string `json:"note"`
	NextAvailableAt string `json:"next_available_at"`
}

func (h *OutboxReconciliationHandler) Resume(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.outboxService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox reconciliation service is unavailable"})
		return
	}
	id, ok := parseOutboxEventID(c)
	if !ok {
		return
	}

	var req outboxReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		recordAdminAudit(h.auditRecorder, c, adminAuditEvent{
			StartedAt: startedAt, Action: adminAuditActionReconcile, Resource: "outbox_event",
			ResourceID: id, Status: adminAuditStatusFailed, ErrorMessage: err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "a reconciliation note is required"})
		return
	}
	note := strings.TrimSpace(req.Note)
	if len(note) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a reconciliation note is required"})
		return
	}
	nextAvailableAt, err := parseOptionalOutboxTime(req.NextAvailableAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "next_available_at must be RFC3339"})
		return
	}

	now := time.Now().UTC()
	if nextAvailableAt.IsZero() {
		nextAvailableAt = now
	}
	err = h.outboxService.ResumeUnknownEvent(id, nextAvailableAt, note, now)
	if err != nil {
		h.recordReconciliationFailure(c, startedAt, id, err)
		if errors.Is(err, repository.ErrOutboxUnknownNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown outbox event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resume unknown outbox event"})
		return
	}

	h.recordReconciliationSuccess(c, startedAt, id, "resume", note)
	response.Success(c, gin.H{"id": id, "status": outbox.EventStatusFailed, "available_at": nextAvailableAt})
}

func (h *OutboxReconciliationHandler) MarkProcessed(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.outboxService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox reconciliation service is unavailable"})
		return
	}
	id, ok := parseOutboxEventID(c)
	if !ok {
		return
	}

	var req outboxReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(strings.TrimSpace(req.Note)) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a reconciliation note is required"})
		return
	}
	note := strings.TrimSpace(req.Note)
	now := time.Now().UTC()
	if err := h.outboxService.MarkUnknownEventProcessed(id, note, now); err != nil {
		h.recordReconciliationFailure(c, startedAt, id, err)
		if errors.Is(err, repository.ErrOutboxUnknownNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown outbox event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark unknown outbox event processed"})
		return
	}

	h.recordReconciliationSuccess(c, startedAt, id, "mark_processed", note)
	response.Success(c, gin.H{"id": id, "status": outbox.EventStatusProcessed, "processed_at": now})
}

func (h *OutboxReconciliationHandler) recordReconciliationFailure(c *gin.Context, startedAt time.Time, id uint, err error) {
	recordAdminAudit(h.auditRecorder, c, adminAuditEvent{
		StartedAt: startedAt, Action: adminAuditActionReconcile, Resource: "outbox_event",
		ResourceID: id, Status: adminAuditStatusFailed, ErrorMessage: err.Error(),
	})
}

func (h *OutboxReconciliationHandler) recordReconciliationSuccess(c *gin.Context, startedAt time.Time, id uint, operation, note string) {
	recordAdminAudit(h.auditRecorder, c, adminAuditEvent{
		StartedAt: startedAt, Action: adminAuditActionReconcile, Resource: "outbox_event",
		ResourceID: id, Status: adminAuditStatusSuccess,
		NewValue: map[string]string{"operation": operation, "note": note},
	})
}

func parseOutboxEventID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outbox event ID"})
		return 0, false
	}
	return uint(id), true
}

func parseOptionalOutboxTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func toOutboxUnknownEventResponse(event outbox.Event) outboxUnknownEventResponse {
	return outboxUnknownEventResponse{
		ID:             event.ID,
		EventKey:       event.EventKey,
		EventType:      event.EventType,
		AggregateType:  event.AggregateType,
		AggregateID:    event.AggregateID,
		Status:         event.Status,
		Attempts:       event.Attempts,
		MaxAttempts:    event.MaxAttempts,
		AvailableAt:    event.AvailableAt,
		LastAttemptAt:  event.LastAttemptAt,
		UncertainAt:    event.UncertainAt,
		ReconcileAfter: event.ReconcileAfter,
		LastError:      event.LastError,
		CreatedAt:      event.CreatedAt,
		UpdatedAt:      event.UpdatedAt,
	}
}
