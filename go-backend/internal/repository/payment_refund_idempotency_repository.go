package repository

import (
	"commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRefundIdempotencyRepository struct {
	db *gorm.DB
}

func NewPaymentRefundIdempotencyRepository(db *gorm.DB) *PaymentRefundIdempotencyRepository {
	return &PaymentRefundIdempotencyRepository{db: db}
}

func (r *PaymentRefundIdempotencyRepository) WithTx(tx *gorm.DB) *PaymentRefundIdempotencyRepository {
	return &PaymentRefundIdempotencyRepository{db: tx}
}

func (r *PaymentRefundIdempotencyRepository) TryCreate(record *payment.RefundIdempotency) (bool, error) {
	if record == nil {
		return false, gorm.ErrInvalidData
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(record)
	return result.RowsAffected == 1, result.Error
}

func (r *PaymentRefundIdempotencyRepository) FindByScopeKey(scope, key string) (*payment.RefundIdempotency, error) {
	var record payment.RefundIdempotency
	err := r.db.Where(
		"scope = ? AND idempotency_key = ?",
		scope,
		key,
	).Clauses(clause.Locking{Strength: "UPDATE"}).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *PaymentRefundIdempotencyRepository) BindRefundID(recordID, refundID uint) error {
	result := r.db.Model(&payment.RefundIdempotency{}).
		Where("id = ? AND refund_id IS NULL", recordID).
		Update("refund_id", refundID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
