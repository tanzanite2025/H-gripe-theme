package repository

import shippingdomain "commerce-platform/internal/domain/shipping"

func (r *ShippingRepository) CreateQuoteSnapshot(snapshot *shippingdomain.QuoteSnapshot) error {
	return r.db.Create(snapshot).Error
}

func (r *ShippingRepository) FindQuoteSnapshotByID(id string) (*shippingdomain.QuoteSnapshot, error) {
	var snapshot shippingdomain.QuoteSnapshot
	if err := r.db.First(&snapshot, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}
