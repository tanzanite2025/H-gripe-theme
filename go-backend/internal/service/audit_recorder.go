package service

import (
	"encoding/json"
	"time"

	"commerce-platform/internal/domain/audit"
	appLogger "commerce-platform/internal/pkg/logger"

	"go.uber.org/zap"
)

// AuditRecorder is the small boundary used by non-HTTP services that need to
// append an audit event without depending on an admin handler.
type AuditRecorder interface {
	CreateAuditLog(log *audit.AuditLog) error
}

type serviceAuditEvent struct {
	StartedAt    time.Time
	Action       string
	Resource     string
	ResourceID   uint
	Status       string
	ErrorMessage string
	Changes      interface{}
	OldValue     interface{}
	NewValue     interface{}
}

func recordServiceAudit(recorder AuditRecorder, event serviceAuditEvent) {
	if recorder == nil {
		return
	}

	createdAt := time.Now().UTC()
	startedAt := event.StartedAt
	if startedAt.IsZero() {
		startedAt = createdAt
	}
	log := &audit.AuditLog{
		Action:       event.Action,
		Resource:     event.Resource,
		ResourceID:   event.ResourceID,
		Changes:      serviceAuditJSON(event.Changes),
		OldValue:     serviceAuditJSON(event.OldValue),
		NewValue:     serviceAuditJSON(event.NewValue),
		Status:       event.Status,
		ErrorMessage: event.ErrorMessage,
		Duration:     int(createdAt.Sub(startedAt).Milliseconds()),
		CreatedAt:    createdAt,
	}
	if err := recorder.CreateAuditLog(log); err != nil {
		appLogger.Error(
			"service audit log write failed",
			zap.String("action", event.Action),
			zap.String("resource", event.Resource),
			zap.Uint("resource_id", event.ResourceID),
			zap.Error(err),
		)
	}
}

func serviceAuditJSON(value interface{}) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}
