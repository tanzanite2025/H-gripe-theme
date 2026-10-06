package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/notification"

	"gorm.io/gorm"
)

const maxEmailDeliveryRecordErrorMessage = 2000

// EmailDeliveryRecordListFilters contains the small set of filters needed by
// the admin delivery-history view. It intentionally has no body or variable
// search because those values are never persisted in the audit table.
type EmailDeliveryRecordListFilters struct {
	Status       string
	EventType    string
	TemplateCode string
	Search       string
	Page         int
	PageSize     int
}

type EmailDeliveryRecordRepository struct {
	db *gorm.DB
}

func NewEmailDeliveryRecordRepository(db *gorm.DB) *EmailDeliveryRecordRepository {
	return &EmailDeliveryRecordRepository{db: db}
}

// UpsertEmailDeliveryRecordForSending creates the first attempt or refreshes the same record when
// Outbox retries the event. EventKey is the durable idempotency boundary.
func (r *EmailDeliveryRecordRepository) UpsertEmailDeliveryRecordForSending(record *notification.EmailDeliveryRecord) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if record == nil || strings.TrimSpace(record.EventKey) == "" {
		return gorm.ErrInvalidData
	}
	var existing notification.EmailDeliveryRecord
	findResult := r.db.Where("event_key = ?", record.EventKey).Limit(1).Find(&existing)
	if findResult.Error != nil {
		return findResult.Error
	}
	if findResult.RowsAffected > 0 {
		if !existing.FirstAttemptAt.IsZero() && (record.FirstAttemptAt.IsZero() || existing.FirstAttemptAt.Before(record.FirstAttemptAt)) {
			record.FirstAttemptAt = existing.FirstAttemptAt
		}
		if record.FirstAttemptAt.IsZero() {
			record.FirstAttemptAt = time.Now().UTC()
		}
		record.ID = existing.ID
		return r.db.Model(&notification.EmailDeliveryRecord{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"outbox_event_id":  record.OutboxEventID,
			"event_type":       record.EventType,
			"template_code":    record.TemplateCode,
			"template_locale":  record.TemplateLocale,
			"template_version": record.TemplateVersion,
			"recipient_email":  record.RecipientEmail,
			"subject":          record.Subject,
			"reference_type":   record.ReferenceType,
			"reference_number": record.ReferenceNumber,
			"status":           record.Status,
			"attempt_count":    record.AttemptCount,
			"provider_code":    record.ProviderCode,
			"last_error":       record.LastError,
			"first_attempt_at": record.FirstAttemptAt,
			"last_attempt_at":  record.LastAttemptAt,
			"sent_at":          nil,
			"updated_at":       time.Now().UTC(),
		}).Error
	}
	return r.db.Create(record).Error
}

func (r *EmailDeliveryRecordRepository) MarkEmailDeliveryRecordSent(eventKey string, sentAt time.Time) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if sentAt.IsZero() {
		sentAt = time.Now().UTC()
	} else {
		sentAt = sentAt.UTC()
	}
	return r.db.Model(&notification.EmailDeliveryRecord{}).
		Where("event_key = ?", strings.TrimSpace(eventKey)).
		Updates(map[string]interface{}{
			"status":     notification.EmailDeliveryStatusSent,
			"sent_at":    sentAt,
			"last_error": "",
			"updated_at": sentAt,
		}).Error
}

func (r *EmailDeliveryRecordRepository) MarkEmailDeliveryRecordFailed(eventKey, status, errorMessage string, failedAt time.Time) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if failedAt.IsZero() {
		failedAt = time.Now().UTC()
	} else {
		failedAt = failedAt.UTC()
	}
	if status != notification.EmailDeliveryStatusUnknown {
		status = notification.EmailDeliveryStatusFailed
	}
	errorMessage = strings.TrimSpace(errorMessage)
	if len(errorMessage) > maxEmailDeliveryRecordErrorMessage {
		errorMessage = errorMessage[:maxEmailDeliveryRecordErrorMessage]
	}
	return r.db.Model(&notification.EmailDeliveryRecord{}).
		Where("event_key = ?", strings.TrimSpace(eventKey)).
		Updates(map[string]interface{}{
			"status":          status,
			"last_error":      errorMessage,
			"last_attempt_at": failedAt,
			"updated_at":      failedAt,
		}).Error
}

func (r *EmailDeliveryRecordRepository) ListEmailDeliveryRecords(filters EmailDeliveryRecordListFilters) ([]notification.EmailDeliveryRecord, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := r.db.Model(&notification.EmailDeliveryRecord{})
	if status := strings.TrimSpace(filters.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if eventType := strings.TrimSpace(filters.EventType); eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if templateCode := strings.TrimSpace(filters.TemplateCode); templateCode != "" {
		query = query.Where("template_code = ?", templateCode)
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		like := "%" + search + "%"
		query = query.Where("LOWER(subject) LIKE LOWER(?) OR LOWER(reference_number) LIKE LOWER(?) OR LOWER(recipient_email) LIKE LOWER(?) OR LOWER(event_type) LIKE LOWER(?) OR LOWER(template_code) LIKE LOWER(?)", like, like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []notification.EmailDeliveryRecord
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
