package repository

import shippingdomain "commerce-platform/internal/domain/shipping"

import "time"

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

// DeleteExpiredQuoteSnapshots removes a bounded batch of snapshots whose
// validity window has elapsed. Selecting IDs first keeps the operation
// portable across PostgreSQL and SQLite (both used by the application and
// contract tests) while avoiding an unbounded delete transaction.
func (r *ShippingRepository) DeleteExpiredQuoteSnapshots(now time.Time, limit int) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 500
	}

	var ids []string
	if err := r.db.Model(&shippingdomain.QuoteSnapshot{}).
		Where("expires_at <= ?", now.UTC()).
		Order("expires_at ASC").
		Limit(limit).
		Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.db.Where("id IN ?", ids).Delete(&shippingdomain.QuoteSnapshot{})
	return result.RowsAffected, result.Error
}
