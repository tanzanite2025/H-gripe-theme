package admin

import (
	"encoding/json"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"
	appLogger "commerce-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	adminAuditActionApprove   = "approve"
	adminAuditActionRollback  = "rollback"
	adminAuditActionReconcile = "reconcile"

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

func recordAdminAudit(recorder adminAuditRecorder, c *gin.Context, event adminAuditEvent) error {
	if recorder == nil || c == nil {
		return nil
	}
	log := newAdminAuditLog(c, event)
	if err := recorder.CreateAuditLog(log); err != nil {
		fields := []zap.Field{
			zap.String("action", event.Action),
			zap.String("resource", event.Resource),
			zap.Uint("resource_id", event.ResourceID),
			zap.String("path", log.Path),
			zap.String("method", log.Method),
			zap.Error(err),
		}
		if log.UserID > 0 {
			fields = append(fields, zap.Uint("user_id", log.UserID))
		}
		appLogger.Error("admin audit log write failed", fields...)
		return err
	}
	return nil
}

func newAdminAuditLog(c *gin.Context, event adminAuditEvent) *audit.AuditLog {
	createdAt := time.Now().UTC()
	startedAt := event.StartedAt
	if startedAt.IsZero() {
		startedAt = createdAt
	}
	var userID uint
	var username string
	var ipAddress string
	if c != nil {
		userID = c.GetUint("user_id")
		username = strings.TrimSpace(c.GetString("username"))
		if username == "" {
			username = strings.TrimSpace(c.GetString("email"))
		}
		ipAddress = c.ClientIP()
	}

	log := audit.AuditLog{
		UserID:       userID,
		Username:     username,
		Action:       event.Action,
		Resource:     event.Resource,
		ResourceID:   event.ResourceID,
		IPAddress:    ipAddress,
		Changes:      adminAuditJSON(event.Changes),
		OldValue:     adminAuditJSON(event.OldValue),
		NewValue:     adminAuditJSON(event.NewValue),
		Status:       event.Status,
		ErrorMessage: event.ErrorMessage,
		Duration:     int(createdAt.Sub(startedAt).Milliseconds()),
		CreatedAt:    createdAt,
	}
	if c != nil && c.Request != nil {
		log.Method = c.Request.Method
		log.UserAgent = c.Request.UserAgent()
		if c.Request.URL != nil {
			log.Path = c.Request.URL.Path
		}
	}

	return &log
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
