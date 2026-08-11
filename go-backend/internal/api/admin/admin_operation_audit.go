package admin

import (
	"encoding/json"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"

	"github.com/gin-gonic/gin"
)

const (
	adminAuditActionCreate    = "create"
	adminAuditActionUpdate    = "update"
	adminAuditActionDelete    = "delete"
	adminAuditActionExecute   = "execute"
	adminAuditActionProbe     = "probe"
	adminAuditActionRecompute = "recompute"
	adminAuditActionRevoke    = "revoke"
	adminAuditActionSubmit    = "submit"

	adminAuditStatusSuccess = "success"
	adminAuditStatusFailed  = "failed"
)

type adminAuditRecorder interface {
	CreateAuditLog(log *audit.AuditLog) error
}

type adminAuditEvent struct {
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

func adminAuditStartedAt() time.Time {
	return time.Now().UTC()
}

func recordAdminAudit(recorder adminAuditRecorder, c *gin.Context, event adminAuditEvent) {
	if recorder == nil || c == nil {
		return
	}
	createdAt := time.Now().UTC()
	startedAt := event.StartedAt
	if startedAt.IsZero() {
		startedAt = createdAt
	}
	username := strings.TrimSpace(c.GetString("username"))
	if username == "" {
		username = strings.TrimSpace(c.GetString("email"))
	}

	log := audit.AuditLog{
		UserID:       c.GetUint("user_id"),
		Username:     username,
		Action:       event.Action,
		Resource:     event.Resource,
		ResourceID:   event.ResourceID,
		IPAddress:    c.ClientIP(),
		Changes:      adminAuditJSON(event.Changes),
		OldValue:     adminAuditJSON(event.OldValue),
		NewValue:     adminAuditJSON(event.NewValue),
		Status:       event.Status,
		ErrorMessage: event.ErrorMessage,
		Duration:     int(createdAt.Sub(startedAt).Milliseconds()),
		CreatedAt:    createdAt,
	}
	if c.Request != nil {
		log.Method = c.Request.Method
		log.UserAgent = c.Request.UserAgent()
		if c.Request.URL != nil {
			log.Path = c.Request.URL.Path
		}
	}

	_ = recorder.CreateAuditLog(&log)
}

func adminAuditJSON(value interface{}) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}
