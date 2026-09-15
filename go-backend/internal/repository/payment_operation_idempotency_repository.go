package repository

import (
	"commerce-platform/internal/domain/payment"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentOperationIdempotencyRepository struct {
	db *gorm.DB
}

func NewPaymentOperationIdempotencyRepository(db *gorm.DB) *PaymentOperationIdempotencyRepository {
	return &PaymentOperationIdempotencyRepository{db: db}
}

// TryCreate claims the operation key using the database unique constraint.
func (r *PaymentOperationIdempotencyRepository) TryCreate(record *payment.PaymentOperationIdempotency) (bool, error) {
	if record == nil {
		return false, fmt.Errorf("payment operation idempotency record is required")
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(record)
	return result.RowsAffected == 1, result.Error
}

func (r *PaymentOperationIdempotencyRepository) FindByUserScopeKey(userID uint, scope, key string) (*payment.PaymentOperationIdempotency, error) {
	var record payment.PaymentOperationIdempotency
	if err := r.db.Where("user_id = ? AND scope = ? AND idempotency_key = ?", userID, scope, key).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *PaymentOperationIdempotencyRepository) Complete(id uint, statusCode int, contentType, responseBody string) error {
	result := r.db.Model(&payment.PaymentOperationIdempotency{}).
		Where("id = ? AND status = ?", id, payment.PaymentOperationIdempotencyPending).
		Updates(map[string]interface{}{
			"status":        payment.PaymentOperationIdempotencyCompleted,
			"status_code":   statusCode,
			"content_type":  contentType,
			"response_body": responseBody,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("payment operation idempotency record %d was not pending", id)
	}
	return nil
}

func (r *PaymentOperationIdempotencyRepository) Delete(id uint) error {
	return r.db.Delete(&payment.PaymentOperationIdempotency{}, id).Error
}
