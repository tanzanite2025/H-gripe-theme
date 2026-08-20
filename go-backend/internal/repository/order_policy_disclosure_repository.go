package repository

import (
	"commerce-platform/internal/domain/order"

	"gorm.io/gorm"
)

type OrderPolicyDisclosureRepository struct {
	db *gorm.DB
}

func NewOrderPolicyDisclosureRepository(db *gorm.DB) *OrderPolicyDisclosureRepository {
	return &OrderPolicyDisclosureRepository{db: db}
}

func (r *OrderPolicyDisclosureRepository) WithTx(tx *gorm.DB) *OrderPolicyDisclosureRepository {
	return &OrderPolicyDisclosureRepository{db: tx}
}

func (r *OrderPolicyDisclosureRepository) Create(disclosure *order.PolicyDisclosure) error {
	return r.db.Create(disclosure).Error
}

func (r *OrderPolicyDisclosureRepository) FindByOrderID(orderID uint) (*order.PolicyDisclosure, error) {
	var disclosure order.PolicyDisclosure
	err := r.db.
		Where("order_id = ? AND policy_key = ?", orderID, "refund_return_policy").
		First(&disclosure).Error
	if err != nil {
		return nil, err
	}
	return &disclosure, nil
}
