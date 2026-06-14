package repository

import (
	"tanzanite/internal/domain/feedback"

	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) List(threadKey, status, search string, page, pageSize int) ([]feedback.Feedback, int64, error) {
	var items []feedback.Feedback
	var total int64

	query := r.db.Model(&feedback.Feedback{}).Where("thread_key = ?", threadKey)
	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", "approved")
	}
	if search != "" {
		query = query.Where("content LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *FeedbackRepository) Create(item *feedback.Feedback) error {
	return r.db.Create(item).Error
}

func (r *FeedbackRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&feedback.Feedback{}).Where("id = ?", id).Update("status", status).Error
}
