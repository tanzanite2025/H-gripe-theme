package repository

import (
	"tanzanite/internal/domain/recommendation"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RecommendationEventRepository struct {
	db *gorm.DB
}

func NewRecommendationEventRepository(db *gorm.DB) *RecommendationEventRepository {
	return &RecommendationEventRepository{db: db}
}

// CreateBatch is idempotent on the client-generated event_id.
func (r *RecommendationEventRepository) CreateBatch(events []recommendation.Event) (int64, error) {
	if len(events) == 0 {
		return 0, nil
	}

	result := r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_id"}},
			DoNothing: true,
		}).
		CreateInBatches(&events, 50)

	return result.RowsAffected, result.Error
}

func (r *RecommendationEventRepository) DeleteExpiredByTypes(eventTypes []string, cutoff time.Time, batchLimit int) (int64, error) {
	if r == nil || r.db == nil || len(eventTypes) == 0 {
		return 0, nil
	}
	if cutoff.IsZero() {
		return 0, nil
	}
	if batchLimit <= 0 {
		batchLimit = 5000
	}

	var ids []uint
	if err := r.db.Model(&recommendation.Event{}).
		Select("id").
		Where("event_type IN ?", eventTypes).
		Where("occurred_at < ?", cutoff.UTC()).
		Order("occurred_at ASC").
		Limit(batchLimit).
		Find(&ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	result := r.db.Where("id IN ?", ids).Delete(&recommendation.Event{})
	return result.RowsAffected, result.Error
}
