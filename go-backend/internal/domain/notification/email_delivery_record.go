package notification

import (
	"strings"
	"time"
)

const (
	EmailDeliveryStatusSending = "sending"
	EmailDeliveryStatusSent    = "sent"
	EmailDeliveryStatusFailed  = "failed"
	EmailDeliveryStatusUnknown = "unknown"
)

// EmailDeliveryRecord is the durable audit pointer for a transactional email.
// It deliberately does not contain either rendered body format or template
// variables. The Outbox event remains the retry source of truth.
type EmailDeliveryRecord struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	OutboxEventID   uint       `gorm:"not null;index" json:"-"`
	EventKey        string     `gorm:"size:255;not null;uniqueIndex" json:"-"`
	EventType       string     `gorm:"size:80;not null;index" json:"event_type"`
	TemplateCode    string     `gorm:"size:64;not null;index" json:"template_code"`
	TemplateLocale  string     `gorm:"size:16;not null;default:'en'" json:"locale"`
	TemplateVersion int        `gorm:"not null;default:1" json:"template_version"`
	RecipientEmail  string     `gorm:"size:255;not null" json:"-"`
	Subject         string     `gorm:"size:255;not null;default:''" json:"subject"`
	ReferenceType   string     `gorm:"size:32;not null;default:'';index" json:"reference_type"`
	ReferenceNumber string     `gorm:"size:128;not null;default:'';index" json:"reference_number"`
	Status          string     `gorm:"size:20;not null;index" json:"status"`
	AttemptCount    int        `gorm:"not null;default:0" json:"attempt_count"`
	ProviderCode    string     `gorm:"size:64;not null;default:''" json:"provider_code,omitempty"`
	LastError       string     `gorm:"type:text;not null;default:''" json:"last_error,omitempty"`
	FirstAttemptAt  time.Time  `gorm:"not null;index" json:"first_attempt_at"`
	LastAttemptAt   time.Time  `gorm:"not null;index" json:"last_attempt_at"`
	SentAt          *time.Time `gorm:"index" json:"sent_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (EmailDeliveryRecord) TableName() string { return "email_delivery_records" }

// EmailDeliveryRecordView is the safe admin representation. RecipientEmail
// is masked before it crosses the API boundary.
type EmailDeliveryRecordView struct {
	ID              uint       `json:"id"`
	EventType       string     `json:"event_type"`
	TemplateCode    string     `json:"template_code"`
	Locale          string     `json:"locale"`
	TemplateVersion int        `json:"template_version"`
	RecipientEmail  string     `json:"recipient_email"`
	Subject         string     `json:"subject"`
	ReferenceType   string     `json:"reference_type,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty"`
	Status          string     `json:"status"`
	AttemptCount    int        `json:"attempt_count"`
	ProviderCode    string     `json:"provider_code,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
	FirstAttemptAt  time.Time  `json:"first_attempt_at"`
	LastAttemptAt   time.Time  `json:"last_attempt_at"`
	SentAt          *time.Time `json:"sent_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (record EmailDeliveryRecord) ToAdminEmailDeliveryRecordView() EmailDeliveryRecordView {
	return EmailDeliveryRecordView{
		ID:              record.ID,
		EventType:       record.EventType,
		TemplateCode:    record.TemplateCode,
		Locale:          record.TemplateLocale,
		TemplateVersion: record.TemplateVersion,
		RecipientEmail:  maskEmailAddressForAdminDisplay(record.RecipientEmail),
		Subject:         record.Subject,
		ReferenceType:   record.ReferenceType,
		ReferenceNumber: record.ReferenceNumber,
		Status:          record.Status,
		AttemptCount:    record.AttemptCount,
		ProviderCode:    record.ProviderCode,
		LastError:       record.LastError,
		FirstAttemptAt:  record.FirstAttemptAt,
		LastAttemptAt:   record.LastAttemptAt,
		SentAt:          record.SentAt,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

func maskEmailAddressForAdminDisplay(value string) string {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		if value == "" {
			return "-"
		}
		return "***"
	}
	local := parts[0]
	if len(local) == 1 {
		return local + "***@" + parts[1]
	}
	if len(local) == 2 {
		return local[:1] + "***@" + parts[1]
	}
	return local[:1] + "***" + local[len(local)-1:] + "@" + parts[1]
}
