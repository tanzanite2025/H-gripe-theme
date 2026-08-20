package repository

import (
	"time"

	"commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Transaction 相关方法

// CreateTransaction 创建交易记录
func (r *PaymentRepository) CreateTransaction(t *payment.Transaction) error {
	return r.db.Create(t).Error
}

// CreateTransactionIfAbsent is used for durable payment-attempt claiming.
// The database unique constraint is the final arbiter if two API replicas
// pass the Redis idempotency layer at the same time.
func (r *PaymentRepository) CreateTransactionIfAbsent(t *payment.Transaction) (bool, error) {
	if t == nil {
		return false, gorm.ErrInvalidData
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(t)
	return result.RowsAffected == 1, result.Error
}

// FindTransactionByID 根据ID查找交易
func (r *PaymentRepository) FindTransactionByID(id uint) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PaymentRepository) FindTransactionByIDForUpdate(id uint) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.lockForUpdate(r.db).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindTransactionByOrderID 根据订单ID查找交易
func (r *PaymentRepository) FindTransactionByOrderID(orderID uint) ([]payment.Transaction, error) {
	var transactions []payment.Transaction
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

func (r *PaymentRepository) FindCompletedTransactionByOrderIDForUpdate(orderID uint) (*payment.Transaction, error) {
	var transaction payment.Transaction
	err := r.lockForUpdate(r.db).
		Where("order_id = ? AND status = ?", orderID, "completed").
		Order("created_at DESC, id DESC").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindTransactionByTransactionID 根据交易ID查找
func (r *PaymentRepository) FindTransactionByTransactionID(transactionID string) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.db.Where("transaction_id = ?", transactionID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PaymentRepository) FindTransactionByTransactionIDForUpdate(transactionID string) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.lockForUpdate(r.db).Where("transaction_id = ?", transactionID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PaymentRepository) FindTransactionByAttemptKeyForUpdate(orderID uint, paymentMethod, attemptKey string) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.lockForUpdate(r.db).
		Where("order_id = ? AND payment_method = ? AND attempt_key = ?", orderID, paymentMethod, attemptKey).
		Order("created_at DESC, id DESC").
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTransaction 更新交易
func (r *PaymentRepository) UpdateTransaction(t *payment.Transaction) error {
	return r.db.Save(t).Error
}

func (r *PaymentRepository) ExpireOpenTransactionsByOrderID(orderID uint, expiredAt time.Time) (int64, error) {
	if expiredAt.IsZero() {
		expiredAt = time.Now()
	}
	result := r.db.Session(&gorm.Session{SkipHooks: true}).Model(&payment.Transaction{}).
		Where("order_id = ? AND status IN ?", orderID, []string{"pending", "processing", "requires_action"}).
		Updates(map[string]interface{}{
			"status":        "expired",
			"error_message": "payment attempt expired before completion",
			"updated_at":    expiredAt,
		})
	return result.RowsAffected, result.Error
}
