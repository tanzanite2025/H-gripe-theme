package service

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/repository"
)

// TransactionalNotificationDeliveryRecordService stores delivery metadata
// without storing rendered HTML, plain text, or template variables.
type TransactionalNotificationDeliveryRecordService struct {
	repo *repository.EmailDeliveryRecordRepository
}

func NewTransactionalNotificationDeliveryRecordService(repo *repository.EmailDeliveryRecordRepository) *TransactionalNotificationDeliveryRecordService {
	return &TransactionalNotificationDeliveryRecordService{repo: repo}
}

func (s *TransactionalNotificationDeliveryRecordService) StartEmailDeliveryRecord(
	eventKey string,
	eventID uint,
	eventType string,
	templateCode string,
	locale string,
	templateVersion int,
	recipientEmail string,
	subject string,
	referenceType string,
	referenceNumber string,
	attemptCount int,
	providerCode string,
	attemptedAt time.Time,
) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if attemptedAt.IsZero() {
		attemptedAt = time.Now().UTC()
	} else {
		attemptedAt = attemptedAt.UTC()
	}
	if templateVersion <= 0 {
		templateVersion = 1
	}
	return s.repo.UpsertEmailDeliveryRecordForSending(&notification.EmailDeliveryRecord{
		OutboxEventID:   eventID,
		EventKey:        strings.TrimSpace(eventKey),
		EventType:       strings.TrimSpace(eventType),
		TemplateCode:    strings.TrimSpace(templateCode),
		TemplateLocale:  strings.TrimSpace(locale),
		TemplateVersion: templateVersion,
		RecipientEmail:  strings.TrimSpace(recipientEmail),
		Subject:         strings.TrimSpace(subject),
		ReferenceType:   strings.TrimSpace(referenceType),
		ReferenceNumber: strings.TrimSpace(referenceNumber),
		Status:          notification.EmailDeliveryStatusSending,
		AttemptCount:    attemptCount,
		ProviderCode:    strings.TrimSpace(providerCode),
		FirstAttemptAt:  attemptedAt,
		LastAttemptAt:   attemptedAt,
	})
}

func (s *TransactionalNotificationDeliveryRecordService) MarkEmailDeliveryRecordSent(eventKey string, sentAt time.Time) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.MarkEmailDeliveryRecordSent(eventKey, sentAt)
}

func (s *TransactionalNotificationDeliveryRecordService) MarkEmailDeliveryRecordFailed(eventKey, status, errorMessage string, failedAt time.Time) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.MarkEmailDeliveryRecordFailed(eventKey, status, errorMessage, failedAt)
}

func (s *TransactionalNotificationDeliveryRecordService) ListEmailDeliveryRecordViews(
	filters repository.EmailDeliveryRecordListFilters,
) ([]notification.EmailDeliveryRecordView, int64, error) {
	if s == nil || s.repo == nil {
		return []notification.EmailDeliveryRecordView{}, 0, nil
	}
	records, total, err := s.repo.ListEmailDeliveryRecords(filters)
	if err != nil {
		return nil, 0, err
	}
	views := make([]notification.EmailDeliveryRecordView, 0, len(records))
	for _, record := range records {
		views = append(views, record.ToAdminEmailDeliveryRecordView())
	}
	return views, total, nil
}
