package repository

import (
	"commerce-platform/internal/domain/order"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderIdempotencyRepository struct {
	db *gorm.DB
}

func NewOrderIdempotencyRepository(db *gorm.DB) *OrderIdempotencyRepository {
	return &OrderIdempotencyRepository{db: db}
}

func (r *OrderIdempotencyRepository) WithTx(tx *gorm.DB) *OrderIdempotencyRepository {
	return &OrderIdempotencyRepository{db: tx}
}

// TryCreate claims a key using the database unique constraint. A conflict is
// reported as created=false rather than as an error so the caller can replay
// the order already bound to the key.
func (r *OrderIdempotencyRepository) TryCreate(record *order.OrderIdempotency) (bool, error) {
	if record == nil {
		return false, fmt.Errorf("order idempotency record is required")
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(record)
	return result.RowsAffected == 1, result.Error
}

func (r *OrderIdempotencyRepository) FindByUserScopeKey(userID uint, scope, key string) (*order.OrderIdempotency, error) {
	var record order.OrderIdempotency
	query := r.db.Where("user_id = ? AND scope = ? AND idempotency_key = ?", userID, scope, key)
	query = r.lockForUpdate(query)
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *OrderIdempotencyRepository) BindOrderID(recordID, orderID uint) error {
	result := r.db.Model(&order.OrderIdempotency{}).
		Where("id = ? AND order_id IS NULL", recordID).
		Update("order_id", orderID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("order idempotency record %d was not left unbound", recordID)
	}
	return nil
}

func (r *OrderIdempotencyRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}
