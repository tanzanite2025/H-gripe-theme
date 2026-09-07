package repository

import (
	"commerce-platform/internal/domain/orderevidence"

	"gorm.io/gorm"
)

type OrderEvidenceSnapshotRepository struct {
	db *gorm.DB
}

func NewOrderEvidenceSnapshotRepository(db *gorm.DB) *OrderEvidenceSnapshotRepository {
	return &OrderEvidenceSnapshotRepository{db: db}
}

func (r *OrderEvidenceSnapshotRepository) WithTx(tx *gorm.DB) *OrderEvidenceSnapshotRepository {
	return &OrderEvidenceSnapshotRepository{db: tx}
}

func (r *OrderEvidenceSnapshotRepository) Create(snapshot *orderevidence.OrderEvidenceSnapshot) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Create(snapshot).Error
}

func (r *OrderEvidenceSnapshotRepository) FindByOrderID(orderID uint) (*orderevidence.OrderEvidenceSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var snapshot orderevidence.OrderEvidenceSnapshot
	if err := r.db.Where("order_id = ?", orderID).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}
