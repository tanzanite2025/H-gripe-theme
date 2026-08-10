package seo

import (
	"encoding/json"
	"strings"
	"time"

	"tanzanite/internal/domain/audit"

	"github.com/gin-gonic/gin"
)

const (
	seoAuditActionUpdate = "update"
	seoAuditStatusOK     = "success"
	seoAuditStatusFailed = "failed"

	seoAuditResourceHome    = "seo_home"
	seoAuditResourceArticle = "seo_article"
	seoAuditResourceProduct = "seo_product"
)

type seoAuditRecorder interface {
	CreateAuditLog(log *audit.AuditLog) error
}

type seoAuditEvent struct {
	StartedAt    time.Time
	Resource     string
	ResourceID   uint
	Status       string
	ErrorMessage string
	Changes      interface{}
	OldValue     interface{}
	NewValue     interface{}
}

func recordSEOAudit(recorder seoAuditRecorder, c *gin.Context, event seoAuditEvent) {
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
		Action:       seoAuditActionUpdate,
		Resource:     event.Resource,
		ResourceID:   event.ResourceID,
		IPAddress:    c.ClientIP(),
		Changes:      seoAuditJSON(event.Changes),
		OldValue:     seoAuditJSON(event.OldValue),
		NewValue:     seoAuditJSON(event.NewValue),
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

func seoAuditJSON(value interface{}) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func seoFieldChanges(values map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{}, len(values))
	for key := range values {
		changes[key] = true
	}
	return changes
}
