package repository

import (
	"commerce-platform/internal/domain/attribution"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderAttributionRepository struct {
	db *gorm.DB
}

func NewOrderAttributionRepository(db *gorm.DB) *OrderAttributionRepository {
	return &OrderAttributionRepository{db: db}
}

func (r *OrderAttributionRepository) WithTx(tx *gorm.DB) *OrderAttributionRepository {
	return &OrderAttributionRepository{db: tx}
}

func (r *OrderAttributionRepository) Create(value *attribution.OrderAttribution) error {
	if r == nil || r.db == nil || value == nil {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}},
		DoNothing: true,
	}).Create(value).Error
}

func (r *OrderAttributionRepository) FindByOrderID(orderID uint) (*attribution.OrderAttribution, error) {
	if r == nil || r.db == nil || orderID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var value attribution.OrderAttribution
	if err := r.db.Where("order_id = ?", orderID).First(&value).Error; err != nil {
		return nil, err
	}
	return &value, nil
}
