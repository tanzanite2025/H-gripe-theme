package repository

import (
	"time"

	"tanzanite/internal/domain/payment"

	"gorm.io/gorm/clause"
)

// ClaimStripeWebhookEvent creates an event inbox row once. Processed events
// are acknowledged without re-running side effects; failed events can be
// claimed again on a later Stripe retry.
func (r *PaymentRepository) ClaimStripeWebhookEvent(eventID, eventType, payload string) (bool, error) {
	event := &payment.StripeWebhookEvent{
		EventID:   eventID,
		EventType: eventType,
		Status:    "processing",
		Payload:   payload,
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(event)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	var existing payment.StripeWebhookEvent
	if err := r.db.Where("event_id = ?", eventID).First(&existing).Error; err != nil {
		return false, err
	}
	if existing.Status == "processed" || existing.Status == "processing" {
		return false, nil
	}

	result = r.db.Model(&payment.StripeWebhookEvent{}).
		Where("event_id = ? AND status = ?", eventID, "failed").
		Updates(map[string]interface{}{
			"status":        "processing",
			"payload":       payload,
			"error_message": "",
			"processed_at":  nil,
		})
	return result.RowsAffected > 0, result.Error
}

func (r *PaymentRepository) MarkStripeWebhookEventProcessed(eventID string) error {
	now := time.Now()
	return r.db.Model(&payment.StripeWebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":        "processed",
			"error_message": "",
			"processed_at":  &now,
		}).Error
}

func (r *PaymentRepository) MarkStripeWebhookEventFailed(eventID string, processingErr error) error {
	message := ""
	if processingErr != nil {
		message = processingErr.Error()
	}
	return r.db.Model(&payment.StripeWebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": message,
		}).Error
}
