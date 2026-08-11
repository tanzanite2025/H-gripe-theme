package repository

import (
	"time"

	"tanzanite/internal/domain/payment"

	"gorm.io/gorm"
)

// Transaction 相关方法

// CreateTransaction 创建交易记录
func (r *PaymentRepository) CreateTransaction(t *payment.Transaction) error {
	return r.db.Create(t).Error
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
